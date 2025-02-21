package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/cache"
	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/dao"
	"github.com/crazyfrankie/cloudstorage/app/file/mws"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
)

type FileServer struct {
	repo   *repository.UploadRepo
	minio  *mws.MinioServer
	worker DownloadWorker
	file.UnimplementedFileServiceServer
}

func NewFileServer(repo *repository.UploadRepo, minio *mws.MinioServer, worker DownloadWorker) *FileServer {
	return &FileServer{repo: repo, minio: minio, worker: worker}
}

// Upload 处理小文件上传
func (s *FileServer) Upload(ctx context.Context, req *file.UploadRequest) (*file.UploadResponse, error) {
	meta, data := req.GetMetadata(), req.GetData()

	// 秒传
	existFile, err := s.repo.QueryByHash(ctx, meta.Hash)
	if err != nil {
		return nil, err
	}
	if existFile.Id != 0 {
		return &file.UploadResponse{Id: int32(existFile.Id)}, nil
	}

	// 查询容量
	enough, err := s.repo.QueryCapacity(ctx, meta.GetUserId(), meta.GetSize())
	if err != nil {
		return nil, err
	}
	if !enough {
		return nil, errors.New("you're on lower capacity")
	}

	// 存到本地
	err = s.saveFile(meta.Path, data)
	if err != nil {
		return nil, err
	}

	// 存 OSS
	go func() {
		newCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err := s.minio.PutToBucket(newCtx, s.minio.BucketName, meta.Name, meta.Size, data)
		if err != nil {
			log.Printf("failed to load oss:%s", meta.Name)
		}
	}()

	// 存数据库
	f := &dao.File{
		Name:     meta.GetName(),
		Hash:     meta.GetHash(),
		Type:     meta.GetContentType(),
		Path:     meta.GetPath(),
		Size:     meta.GetSize(),
		UserId:   meta.GetUserId(),
		FolderId: meta.GetFolderId(),
	}

	err = s.repo.CreateFile(ctx, f)
	if err != nil {
		return nil, err
	}

	return &file.UploadResponse{
		Id: int32(f.Id),
	}, nil
}

// UploadChunkStream v2 分片上传, 处理流式分片上传
func (s *FileServer) UploadChunkStream(stream file.FileService_UploadChunkStreamServer) error {
	var uploadId string
	var filename string
	var partNumber int32 = 0
	var userId int32
	var folderId int64
	parts := make([]minio.CompletePart, 0)

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			// 完成上传
			if err := s.completeMultipartUpload(stream.Context(), uploadId, filename, parts, &dao.File{
				Name:     filename,
				UserId:   userId,
				Type:     filepath.Ext(filename)[1:],
				Path:     filename,
				FolderId: folderId,
			}); err != nil {
				return err
			}
			return stream.SendAndClose(&file.UploadChunkResponse{
				UploadId: uploadId,
			})
		}
		if err != nil {
			return err
		}

		// 初始化上传或获取之前保存的信息
		if partNumber == 0 {
			filename = chunk.Filename
			userId = chunk.UserId
			folderId = chunk.FolderId

			// 检查存储空间
			enough, err := s.repo.QueryCapacity(stream.Context(), userId, chunk.FileSize)
			if err != nil {
				return err
			}
			if !enough {
				return errors.New("insufficient storage space")
			}

			// 初始化分片上传
			uploadId, err = s.initMultipartUpload(stream.Context(), filename)
			if err != nil {
				return err
			}
		}

		// 上传分片
		partNumber++
		etag, err := s.uploadPart(stream.Context(), uploadId, filename, partNumber, chunk.Data)
		if err != nil {
			return err
		}

		// 保存分片信息
		parts = append(parts, minio.CompletePart{
			PartNumber: int(partNumber),
			ETag:       etag,
		})
	}
}

// UploadChunk v1 处理分片上传
func (s *FileServer) UploadChunk(ctx context.Context, req *file.UploadChunkRequest) (*file.UploadChunkResponse, error) {
	// 第一个分片需要初始化
	if req.PartNumber == 1 && req.UploadId == "" {
		// 检查存储空间
		enough, err := s.repo.QueryCapacity(ctx, req.UserId, req.FileSize)
		if err != nil {
			return nil, err
		}
		if !enough {
			return nil, errors.New("insufficient storage capacity")
		}

		// 初始化分片上传
		uploadID, err := s.initMultipartUpload(ctx, req.Filename)
		if err != nil {
			return nil, err
		}
		req.UploadId = uploadID
	}

	// 上传分片
	etag, err := s.uploadPart(ctx, req.UploadId, req.Filename, req.PartNumber, req.Data)
	if err != nil {
		return nil, err
	}

	// 保存分片ETag到Redis
	if err := s.repo.SavePartETag(ctx, req.UploadId, int(req.PartNumber), etag); err != nil {
		return nil, err
	}

	// 判断是否是最后一个分片
	if req.IsLast {
		// 获取所有分片的ETag
		etags, err := s.repo.GetPartETags(ctx, req.UploadId)
		if err != nil {
			return nil, err
		}

		// 构建完成分片上传请求
		parts := make([]minio.CompletePart, len(etags))
		for partNumber, etag := range etags {
			parts[partNumber-1] = minio.CompletePart{
				PartNumber: partNumber,
				ETag:       etag,
			}
		}

		// 完成分片上传
		if err := s.completeMultipartUpload(ctx, req.UploadId, req.Filename, parts, &dao.File{
			Name:     req.Filename,
			UserId:   req.UserId,
			Type:     filepath.Ext(req.Filename)[1:],
			Path:     req.Filename,
			FolderId: req.FolderId,
		}); err != nil {
			return nil, err
		}
	}

	return &file.UploadChunkResponse{
		UploadId: req.UploadId,
		Etag:     etag,
	}, nil
}

// initMultipartUpload 初始化分片上传
func (s *FileServer) initMultipartUpload(ctx context.Context, filename string) (string, error) {
	return s.minio.CreateMultipartUpload(ctx, s.minio.BucketName, filename)
}

// uploadPart 上传分片
func (s *FileServer) uploadPart(ctx context.Context, uploadId string, filename string, partNumber int32, data []byte) (string, error) {
	partInfo, err := s.minio.PutObjectPart(ctx, s.minio.BucketName, filename, uploadId, int(partNumber), data, int64(len(data)))
	if err != nil {
		return "", err
	}
	return partInfo.ETag, nil
}

// completeMultipartUpload 完成分片上传
func (s *FileServer) completeMultipartUpload(ctx context.Context, uploadId string, filename string, parts []minio.CompletePart, file *dao.File) error {
	// 完成MinIO的分片上传
	_, err := s.minio.CompleteMultipartUpload(ctx, s.minio.BucketName, filename, uploadId, parts)
	if err != nil {
		return err
	}

	// 获取文件信息
	obj, err := s.minio.GetObject(ctx, s.minio.BucketName, filename)
	if err != nil {
		return err
	}
	defer obj.Close()

	// 获取文件详细信息
	stat, err := obj.Stat()
	if err != nil {
		return err
	}

	// 更新文件大小
	file.Size = stat.Size

	// 创建文件记录
	return s.repo.CreateFile(ctx, file)
}

// Download 单个小文件下载
func (s *FileServer) Download(ctx context.Context, req *file.DownloadRequest) (*file.DownloadResponse, error) {
	// 获取文件信息
	fileInfo, err := s.repo.GetFile(ctx, req.FileId, req.UserId)
	if err != nil {
		return nil, err
	}

	// 优先从本地文件系统读取
	data, err := os.ReadFile(fileInfo.Path)
	if err != nil {
		// 本地文件不存在，从minio获取
		obj, err := s.minio.GetObject(ctx, s.minio.BucketName, fileInfo.Name)
		if err != nil {
			return nil, err
		}
		defer obj.Close()

		// 获取对象信息
		info, err := obj.Stat()
		if err != nil {
			return nil, err
		}

		// 读取对象数据
		data = make([]byte, info.Size)
		_, err = obj.Read(data)
		if err != nil && err != io.EOF {
			return nil, err
		}
	}

	return &file.DownloadResponse{
		Data: data,
	}, nil
}

// DownloadStream 大文件流式下载
func (s *FileServer) DownloadStream(req *file.DownloadRequest, stream file.FileService_DownloadStreamServer) error {
	// 获取文件信息
	fileInfo, err := s.repo.GetFile(stream.Context(), req.FileId, req.UserId)
	if err != nil {
		return err
	}

	// 尝试从本地读取
	if f, err := os.Open(fileInfo.Path); err == nil {
		defer f.Close()
		err = s.streamFile(f, stream)
	}

	// 从MinIO读取
	obj, err := s.minio.GetObject(stream.Context(), s.minio.BucketName, fileInfo.Name)
	if err != nil {
		return err
	}
	defer obj.Close()

	return s.streamFile(obj, stream)
}

// DownloadTask 处理下载队列
func (s *FileServer) DownloadTask(ctx context.Context, req *file.DownloadTaskRequest) (*file.DownloadTaskResponse, error) {
	taskId := uuid.New().String()

	// 获取文件信息
	downloadFiles := make([]*cache.DownloadedFile, 0, len(req.Files))
	var totalSize int64

	for _, f := range req.Files {
		fileInfo, err := s.repo.GetFile(ctx, f.FileId, req.UserId)
		if err != nil {
			return nil, err
		}

		downloadFiles = append(downloadFiles, &cache.DownloadedFile{
			FileId: f.FileId,
			Name:   fileInfo.Name,
			Path:   f.Path,
			Size:   fileInfo.Size,
			Status: "pending",
		})
		totalSize += fileInfo.Size
	}

	task := &cache.DownloadTask{
		UserId:     req.UserId,
		Status:     "pending",
		FolderName: req.FolderName,
		TotalSize:  totalSize,
		Progress:   0,
		CreatedAt:  time.Now(),
		Files:      downloadFiles,
	}

	if err := s.repo.CreateDownloadTask(ctx, taskId, task); err != nil {
		return nil, err
	}

	return &file.DownloadTaskResponse{
		TaskId: taskId,
	}, nil
}

// GetDownloadTask 获取下载任务状态
func (s *FileServer) GetDownloadTask(ctx context.Context, req *file.GetDownloadTaskRequest) (*file.GetDownloadTaskResponse, error) {
	// 从缓存获取任务信息
	task, err := s.repo.GetDownloadTaskInfo(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	// 检查用户权限
	if task.UserId != req.UserId {
		return nil, errors.New("permission denied")
	}

	// 转换为响应格式
	files := make([]*file.FileProgress, 0, len(task.Files))
	for _, f := range task.Files {
		files = append(files, &file.FileProgress{
			FileId:     f.FileId,
			Name:       f.Name,
			Path:       f.Path,
			Size:       f.Size,
			Status:     f.Status,
			Downloaded: f.Downloaded,
		})
	}

	return &file.GetDownloadTaskResponse{
		TaskId:     req.TaskId,
		Status:     task.Status,
		FolderName: task.FolderName,
		TotalSize:  task.TotalSize,
		Progress:   task.Progress,
		Files:      files,
	}, nil
}

// ResumeDownload 断点续传
func (s *FileServer) ResumeDownload(ctx context.Context, req *file.ResumeDownloadRequest) (*file.ResumeDownloadResponse, error) {
	// 获取原任务信息
	task, err := s.repo.GetDownloadTaskInfo(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	// 检查用户权限
	if task.UserId != req.UserId {
		return nil, errors.New("permission denied")
	}

	// 筛选需要继续下载的文件
	fileMap := make(map[int64]struct{})
	for _, fid := range req.FileIds {
		fileMap[fid] = struct{}{}
	}

	var remainingFiles []*cache.DownloadedFile
	for _, f := range task.Files {
		if _, ok := fileMap[f.FileId]; ok {
			remainingFiles = append(remainingFiles, &cache.DownloadedFile{
				FileId:     f.FileId,
				Name:       f.Name,
				Path:       f.Path,
				Size:       f.Size,
				Status:     "pending",
				Downloaded: f.Downloaded, // 保留已下载进度
			})
		}
	}

	// 创建新的下载任务
	newTaskId := uuid.New().String()
	newTask := &cache.DownloadTask{
		UserId:     req.UserId,
		Status:     "pending",
		FolderName: task.FolderName,
		Files:      remainingFiles,
		Progress:   0, // 新任务从0开始计算进度
		TotalSize:  calculateTotalSize(remainingFiles),
		CreatedAt:  time.Now(),
	}

	// 保存新任务
	if err := s.repo.CreateDownloadTask(ctx, newTaskId, newTask); err != nil {
		return nil, err
	}

	return &file.ResumeDownloadResponse{
		NewTaskId: newTaskId,
	}, nil
}

// calculateTotalSize 计算总大小
func calculateTotalSize(files []*cache.DownloadedFile) int64 {
	var total int64
	for _, f := range files {
		total += f.Size - f.Downloaded // 只计算未下载的部分
	}
	return total
}

// GetFile 获取文件信息
func (s *FileServer) GetFile(ctx context.Context, req *file.GetFileRequest) (*file.GetFileResponse, error) {
	fileInfo, err := s.repo.GetFile(ctx, req.FileId, req.UserId)
	if err != nil {
		return nil, err
	}

	utime := time.Unix(fileInfo.Utime, 0).Format(time.DateTime)
	return &file.GetFileResponse{
		File: &file.File{
			Id:       int32(fileInfo.Id),
			Name:     fileInfo.Name,
			FolderId: fileInfo.FolderId,
			UserId:   fileInfo.UserId,
			Size:     fileInfo.Size,
			Type:     fileInfo.Type,
			Utime:    utime,
		},
	}, nil
}

// CreateFileStore 创建资源空间
func (s *FileServer) CreateFileStore(ctx context.Context, req *file.CreateFileStoreRequest) (*file.CreateFileStoreResponse, error) {
	store := &dao.FileStore{
		UserId: req.GetUserId(),
	}

	id, err := s.repo.CreateFileStore(ctx, store)
	if err != nil {
		return nil, err
	}

	return &file.CreateFileStoreResponse{Id: id}, nil
}

// CreateFolder 创建文件夹
func (s *FileServer) CreateFolder(ctx context.Context, req *file.CreateFolderRequest) (*file.CreateFolderResponse, error) {
	folder := &dao.Folder{
		Name:     req.GetName(),
		UserId:   req.GetUserId(),
		ParentId: req.GetParentId(),
	}

	err := s.repo.CreateFolder(ctx, folder)
	if err != nil {
		return nil, err
	}
	utime := time.Unix(folder.Utime, 0).Format(time.DateTime)

	return &file.CreateFolderResponse{Folder: &file.Folder{
		Id:       folder.Id,
		Name:     folder.Name,
		UserId:   folder.UserId,
		ParentId: folder.ParentId,
		Path:     folder.Path,
		Utime:    utime,
	}}, nil
}

// ListFolder 展示文件夹及文件
func (s *FileServer) ListFolder(ctx context.Context, req *file.ListFolderRequest) (*file.ListFolderResponse, error) {
	fs, fds, err := s.repo.ListFolder(ctx, req.GetFolderId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	files := make([]*file.File, 0, len(fs))
	for _, f := range fs {
		utime := time.Unix(f.Utime, 0).Format(time.DateTime)
		files = append(files, &file.File{
			Id:       int32(f.Id),
			Name:     f.Name,
			Size:     f.Size,
			Type:     f.Type,
			FolderId: f.FolderId,
			UserId:   f.UserId,
			Utime:    utime,
		})
	}

	folders := make([]*file.Folder, 0, len(fds))
	for _, fd := range fds {
		utime := time.Unix(fd.Utime, 0).Format(time.DateTime)
		folders = append(folders, &file.Folder{
			Id:       fd.Id,
			Name:     fd.Name,
			ParentId: fd.ParentId,
			Path:     fd.Path,
			UserId:   fd.UserId,
			Utime:    utime,
		})
	}

	return &file.ListFolderResponse{
		Folders: folders,
		Files:   files,
	}, nil
}

// MoveFile 移动文件
func (s *FileServer) MoveFile(ctx context.Context, req *file.MoveFileRequest) (*file.MoveFileResponse, error) {
	err := s.repo.MoveFile(ctx, req.GetFileId(), req.GetToFolderId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &file.MoveFileResponse{}, nil
}

// MoveFolder 移动文件夹
func (s *FileServer) MoveFolder(ctx context.Context, req *file.MoveFolderRequest) (*file.MoveFolderResponse, error) {
	err := s.repo.MoveFolder(ctx, req.GetFolderId(), req.GetToFolderId(), req.GetUserId(), req.GetFolderName())
	if err != nil {
		return nil, err
	}

	return &file.MoveFolderResponse{}, nil
}

// DeleteFile 删除文件
func (s *FileServer) DeleteFile(ctx context.Context, req *file.DeleteFileRequest) (*file.DeleteFileResponse, error) {
	err := s.repo.DeleteFile(ctx, req.GetFileId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &file.DeleteFileResponse{}, nil
}

// DeleteFolder 删除文件夹
func (s *FileServer) DeleteFolder(ctx context.Context, req *file.DeleteFolderRequest) (*file.DeleteFolderResponse, error) {
	err := s.repo.DeleteFolder(ctx, req.GetFolderId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &file.DeleteFolderResponse{}, nil
}

// Search 搜索文件和文件夹
func (s *FileServer) Search(ctx context.Context, req *file.SearchRequest) (*file.SearchResponse, error) {
	fs, fds, err := s.repo.Search(ctx, req.GetUserId(), req.GetQuery(), req.GetPage(), req.GetSize())
	if err != nil {
		return nil, err
	}

	files := make([]*file.File, 0, len(fs))
	for _, f := range fs {
		utime := time.Unix(f.Utime, 0).Format(time.DateTime)
		files = append(files, &file.File{
			Id:       int32(f.Id),
			Name:     f.Name,
			Size:     f.Size,
			Type:     f.Type,
			FolderId: f.FolderId,
			UserId:   f.UserId,
			Utime:    utime,
		})
	}

	folders := make([]*file.Folder, 0, len(fds))
	for _, fd := range fds {
		utime := time.Unix(fd.Utime, 0).Format(time.DateTime)
		folders = append(folders, &file.Folder{
			Id:       fd.Id,
			Name:     fd.Name,
			ParentId: fd.ParentId,
			Path:     fd.Path,
			UserId:   fd.UserId,
			Utime:    utime,
		})
	}

	return &file.SearchResponse{Files: files, Folders: folders}, nil
}

// Preview 预览
func (s *FileServer) Preview(ctx context.Context, req *file.PreviewRequest) (*file.PreviewResponse, error) {
	// 获取文件信息
	fileInfo, err := s.repo.GetFile(ctx, req.FileId, req.UserId)
	if err != nil {
		return nil, err
	}

	// 判断文件类型
	previewType := s.getPreviewType(fileInfo.Type)
	if previewType == file.PreviewType_UNKNOWN {
		return nil, errors.New("file type not supported for preview")
	}

	// 生成预览URL
	presignedURL, err := s.minio.PresignedGetObject(ctx, s.minio.BucketName, fileInfo.Name, time.Hour)
	if err != nil {
		return nil, err
	}

	// 设置预览相关的参数
	return &file.PreviewResponse{
		PreviewUrl:  presignedURL.String(),
		ContentType: s.getMimeType(fileInfo.Type),
		Type:        previewType,
	}, nil
}

// CreateShareLink 创建分享链接
func (s *FileServer) CreateShareLink(ctx context.Context, req *file.CreateShareLinkRequest) (*file.CreateShareLinkResponse, error) {
	shareId := uuid.New().String()
	expireAt := time.Now().AddDate(0, 0, int(req.ExpireDays))

	// 创建分享记录
	share := &dao.ShareLink{
		Id:       shareId,
		UserId:   req.UserId,
		FolderId: req.FolderId,
		Password: req.Password,
		ExpireAt: expireAt,
		Status:   1,
	}

	// 如果是分享文件，创建文件关联
	if len(req.FileIds) > 0 {
		for _, fileId := range req.FileIds {
			shareFile := &dao.ShareFile{
				ShareId: shareId,
				FileId:  fileId,
			}
			if err := s.repo.CreateShareFile(ctx, shareFile); err != nil {
				return nil, err
			}
		}
	}

	// 保存分享记录
	if err := s.repo.CreateShareLink(ctx, share); err != nil {
		return nil, err
	}

	// 生成分享链接
	shareUrl := fmt.Sprintf("%s/share/%s", "", shareId)

	return &file.CreateShareLinkResponse{
		ShareId:  shareId,
		ShareUrl: shareUrl,
		Password: req.Password,
		ExpireAt: expireAt.Unix(),
	}, nil
}

// SaveToMyDrive 保存到个人网盘
func (s *FileServer) SaveToMyDrive(ctx context.Context, req *file.SaveToMyDriveRequest) (*file.SaveToMyDriveResponse, error) {
	// 验证分享是否有效
	share, err := s.repo.GetShareLink(ctx, req.ShareId)
	if err != nil {
		return nil, err
	}

	// 检查密码
	if share.Password != "" && share.Password != req.Password {
		return nil, errors.New("invalid password")
	}

	// 检查过期时间
	if share.ExpireAt.Before(time.Now()) {
		return nil, errors.New("share link expired")
	}

	// 获取要保存的文件列表
	var files []*dao.File
	var folders []*dao.Folder
	if share.FolderId != 0 {
		// 如果是文件夹分享，获取文件夹下的所有文件
		files, folders, err = s.repo.ListFolder(ctx, req.GetToFolderId(), req.GetUserId())
	} else {
		// 如果是文件分享，获取选中的文件
		files, err = s.repo.GetFilesByIds(ctx, req.FileIds)
	}
	if err != nil {
		return nil, err
	}

	// 检查存储空间
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}
	enough, err := s.repo.QueryCapacity(ctx, req.UserId, totalSize)
	if err != nil || !enough {
		return nil, errors.New("insufficient storage space")
	}

	// 复制文件到用户的网盘
	for _, f := range files {
		newFile := &dao.File{
			UserId:   req.UserId,
			Name:     f.Name,
			Hash:     f.Hash,
			Type:     f.Type,
			Size:     f.Size,
			FolderId: req.ToFolderId,
			Path:     f.Path,
			Status:   1,
		}
		if err := s.repo.CreateFile(ctx, newFile); err != nil {
			return nil, err
		}
	}

	for _, fd := range folders {
		newFolder := &dao.Folder{
			Id:       fd.Id,
			Name:     fd.Name,
			ParentId: fd.ParentId,
			UserId:   req.GetUserId(),
			Path:     fd.Path,
			Status:   0,
		}
		if err := s.repo.CreateFolder(ctx, newFolder); err != nil {
			return nil, err
		}
	}

	return &file.SaveToMyDriveResponse{}, nil
}

// GetUserFileStore 获取用户资源空间信息
func (s *FileServer) GetUserFileStore(ctx context.Context, req *file.GetUserFileStoreRequest) (*file.GetUserFileStoreResponse, error) {
	store, err := s.repo.FindFileStoreById(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &file.GetUserFileStoreResponse{FileStore: &file.FileStore{
		UserId:      store.UserId,
		Capacity:    store.Capacity,
		CurrentSize: store.CurrentSize,
	}}, nil
}

func (s *FileServer) saveFile(path string, data []byte) error {
	newFile, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = newFile.Write(data)
	newFile.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *FileServer) streamFile(r io.Reader, stream file.FileService_DownloadStreamServer) error {
	buffer := make([]byte, 5*1024*1024)
	var totalSent int64

	for {
		n, err := r.Read(buffer)
		if n > 0 {
			if err := stream.Send(&file.DownloadStreamResponse{
				Data: buffer[:n],
			}); err != nil {
				return fmt.Errorf("failed to send chunk: %v", err)
			}
			totalSent += int64(n)
		}

		if err == io.EOF {
			// 确保所有数据都发送完毕
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read: %v", err)
		}
	}

	log.Printf("Total bytes sent: %d", totalSent)
	return nil
}

func (s *FileServer) getPreviewType(fileType string) file.PreviewType {
	switch strings.ToLower(fileType) {
	case "jpg", "jpeg", "png", "gif":
		return file.PreviewType_IMAGE
	case "pdf":
		return file.PreviewType_PDF
	case "doc", "docx", "xls", "xlsx":
		return file.PreviewType_DOCUMENT
	case "txt", "md", "json":
		return file.PreviewType_TEXT
	default:
		return file.PreviewType_UNKNOWN
	}
}

// 根据文件扩展名获取 MIME 类型
func (s *FileServer) getMimeType(ext string) string {
	switch strings.ToLower(ext) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "pdf":
		return "application/pdf"
	case "doc", "docx":
		return "application/msword"
	case "xls", "xlsx":
		return "application/vnd.ms-excel"
	case "txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
