package service

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/dao"
	"github.com/crazyfrankie/cloudstorage/app/file/mws"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
)

type FileServer struct {
	repo  *repository.UploadRepo
	minio *mws.MinioServer
	file.UnimplementedFileServiceServer
}

func NewFileServer(repo *repository.UploadRepo, minio *mws.MinioServer) *FileServer {
	return &FileServer{repo: repo, minio: minio}
}

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

func (s *FileServer) MoveFile(ctx context.Context, req *file.MoveFileRequest) (*file.MoveFileResponse, error) {
	err := s.repo.MoveFile(ctx, req.GetFileId(), req.GetToFolderId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &file.MoveFileResponse{}, nil
}

func (s *FileServer) MoveFolder(ctx context.Context, req *file.MoveFolderRequest) (*file.MoveFolderResponse, error) {
	err := s.repo.MoveFolder(ctx, req.GetFolderId(), req.GetToFolderId(), req.GetUserId(), req.GetFolderName())
	if err != nil {
		return nil, err
	}

	return &file.MoveFolderResponse{}, nil
}

func (s *FileServer) DeleteFile(ctx context.Context, req *file.DeleteFileRequest) (*file.DeleteFileResponse, error) {
	err := s.repo.DeleteFile(ctx, req.GetFileId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &file.DeleteFileResponse{}, nil
}

func (s *FileServer) DeleteFolder(ctx context.Context, req *file.DeleteFolderRequest) (*file.DeleteFolderResponse, error) {
	err := s.repo.DeleteFolder(ctx, req.GetFolderId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &file.DeleteFolderResponse{}, nil
}

func (s *FileServer) SearchFiles(ctx context.Context, req *file.SearchRequest) (*file.SearchResponse, error) {
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
	buffer := make([]byte, 32*1024) // 32KB chunks
	for {
		n, err := r.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := stream.Send(&file.DownloadStreamResponse{
			Data: buffer[:n],
		}); err != nil {
			return err
		}
	}
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
