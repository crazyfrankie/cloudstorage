package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloudstorage/app/gateway/common/consts"
	"github.com/crazyfrankie/cloudstorage/app/gateway/common/response"
	"github.com/crazyfrankie/cloudstorage/app/gateway/common/util"
	"github.com/crazyfrankie/cloudstorage/app/gateway/mws"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
)

type FileHandler struct {
	cli file.FileServiceClient
}

func NewFileHandler(cli file.FileServiceClient) *FileHandler {
	return &FileHandler{cli: cli}
}

func (h *FileHandler) RegisterRoute(r *gin.Engine) {
	fileGroup := r.Group("/api/files", mws.Auth())
	{
		fileGroup.POST("/upload", h.Upload())
		fileGroup.POST("/large-upload", h.UploadLargeFile())
		fileGroup.GET("/download/:id", h.Download())
		fileGroup.GET("/preview/:id", h.Preview())
		fileGroup.POST("/download/task-queue", h.BatchDownloadFiles())
		fileGroup.GET("/download/task/:taskId", h.GetDownloadTask())
		fileGroup.POST("/download/resume", h.ResumeDownload())
		fileGroup.POST("/search", h.SearchFiles())
		fileGroup.POST("/move", h.MoveFile())
		fileGroup.POST("/delete", h.DeleteFile())
		fileGroup.POST("/folder/create", h.CreateFolder())
		fileGroup.POST("/folder/list", h.ListFolder())
		fileGroup.POST("/folder/move", h.MoveFolder())
	}
}

// Upload 小文件的上传
func (h *FileHandler) Upload() gin.HandlerFunc {
	return func(c *gin.Context) {
		f, header, err := c.Request.FormFile("file")
		if err != nil {
			response.Error(c, err)
			return
		}
		defer f.Close()

		// 根据大小选择上传方式
		if header.Size > consts.SmallFileSizeLimit {
			// 重定向到大文件上传
			h.UploadLargeFile()(c)
			return
		}

		var hash string
		hash, err = util.FileHash(f)
		if err != nil {
			response.Error(c, err)
			return
		}

		f.Seek(0, 0)

		name := header.Filename        // 文件名
		path := consts.BasePath + name // 文件本地路径
		strs := strings.Split(name, ".")
		typ := strs[len(strs)-1] // 文件类型
		size := header.Size      // 文件大小

		claims := c.MustGet("claims")
		claim, _ := claims.(*mws.Claim)

		folder, ok := c.GetPostForm("folder")
		if !ok {
			response.Error(c, errors.New("doesn't contain parent id"))
			return
		}
		folderId, _ := strconv.Atoi(folder)
		meta := &file.FileMetaData{
			Name:        name,
			Path:        path,
			Hash:        hash,
			Size:        size,
			ContentType: typ,
			UserId:      claim.UserId,
			FolderId:    int64(folderId),
		}

		var data []byte
		data, err = io.ReadAll(f)
		if err != nil {
			response.Error(c, err)
			return
		}

		resp, err := h.cli.Upload(c.Request.Context(), &file.UploadRequest{
			Metadata: meta,
			Data:     data,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// UploadLargeFile 大文件上传也即大文件分块上传和断点续传
func (h *FileHandler) UploadLargeFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		f, header, err := c.Request.FormFile("file")
		if err != nil {
			response.Error(c, err)
			return
		}
		defer f.Close()

		var hash string
		hash, err = util.FileHash(f)
		if err != nil {
			response.Error(c, err)
			return
		}

		f.Seek(0, 0)
		// 获取文件基本信息
		name := header.Filename
		path := consts.BasePath + name
		size := header.Size
		typ := strings.Split(name, ".")[len(strings.Split(name, "."))-1]
		claims := c.MustGet("claims").(*mws.Claim)

		folder, ok := c.GetPostForm("folder")
		if !ok {
			response.Error(c, errors.New("doesn't contain parent id"))
			return
		}
		folderId, _ := strconv.Atoi(folder)

		// 初始化分块上传，获取uploadId
		initResp, err := h.cli.InitMultipartUpload(c.Request.Context(), &file.InitMultipartUploadRequest{
			Name:   name,
			Size:   size,
			UserId: claims.UserId,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		// 并发上传分块
		const workerCount = 3
		jobs := make(chan struct {
			number int32
			data   []byte
		}, workerCount)
		results := make(chan *file.PartInfo, workerCount)

		// 启动工作协程
		var wg sync.WaitGroup
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					partResp, err := h.cli.UploadPart(c.Request.Context(), &file.UploadPartRequest{
						UploadId:   initResp.UploadId,
						ObjectName: name,
						PartNumber: job.number,
						Data:       job.data,
					})
					if err != nil {
						// 处理错误
						continue
					}
					results <- &file.PartInfo{
						PartNumber: job.number,
						Etag:       partResp.Etag,
					}
				}
			}()
		}

		// 读取文件并分发任务
		partNumber := int32(1)
		for {
			buffer := make([]byte, consts.ChunkSize)
			n, err := f.Read(buffer)
			if err == io.EOF {
				break
			}

			jobs <- struct {
				number int32
				data   []byte
			}{
				number: partNumber,
				data:   buffer[:n],
			}
			partNumber++
		}
		close(jobs)

		// 收集结果
		go func() {
			wg.Wait()
			close(results)
		}()

		var parts []*file.PartInfo
		for part := range results {
			parts = append(parts, part)
		}

		// 按照 PartNumber 对 parts 排序
		sort.Slice(parts, func(i, j int) bool {
			return parts[i].PartNumber < parts[j].PartNumber
		})

		// 完成上传
		resp, err := h.cli.CompleteMultipartUpload(c.Request.Context(), &file.CompleteMultipartUploadRequest{
			UploadId:   initResp.UploadId,
			Parts:      parts,
			ObjectName: name,
			UserId:     claims.UserId,
			Hash:       hash,
			Path:       path,
			FolderId:   int64(folderId),
			Typ:        typ,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// Download 单个文件下载
func (h *FileHandler) Download() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		claims := c.MustGet("claims").(*mws.Claim)

		fileId, _ := strconv.Atoi(id)
		resp, err := h.cli.GetFile(c.Request.Context(), &file.GetFileRequest{
			FileId: int64(fileId),
			UserId: claims.UserId,
		})
		if err != nil {
			response.Error(c, err)
			return
		}
		fileName := resp.GetFile().GetName()
		size := resp.GetFile().Size
		// 根据文件扩展名设置正确的 MIME 类型
		mimeType := getMimeType(resp.GetFile().Type)
		const sizeThreshold = 10 * 1024 * 1024 // 10MB
		if resp.GetFile().GetSize() <= sizeThreshold {
			// 小文件直接下载
			resp, err := h.cli.Download(c.Request.Context(), &file.DownloadRequest{
				FileId: int64(fileId),
				UserId: claims.UserId,
			})
			if err != nil {
				response.Error(c, err)
				return
			}
			setHeader(c, fileName, mimeType)
			c.Header("Content-Length", strconv.FormatInt(size, 10))
			c.Data(http.StatusOK, mimeType, resp.GetData())
			return
		}

		// 大文件流式下载
		stream, err := h.cli.DownloadStream(c.Request.Context(), &file.DownloadRequest{
			FileId: int64(fileId),
			UserId: claims.UserId,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		setHeader(c, fileName, mimeType)
		// 显式设置 chunked transfer encoding
		c.Header("Transfer-Encoding", "true")

		// 使用 Stream 写入响应
		c.Stream(func(w io.Writer) bool {
			chunk, err := stream.Recv()
			if err != nil {
				return false
			}
			_, err = w.Write(chunk.Data)
			return err == nil
		})
	}
}

// BatchDownloadFiles 下载队列
func (h *FileHandler) BatchDownloadFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		type BatchDownloadRequest struct {
			Files []struct {
				FileId   int64  `json:"fileId"`
				OrderNum int32  `json:"orderNum"` // 下载顺序
				Path     string `json:"path"`     // 文件在文件夹中的路径
			} `json:"files"`
			FolderName string `json:"folderName"` // 如果是文件夹下载，保存文件夹名称
		}

		var req BatchDownloadRequest
		if err := c.Bind(&req); err != nil {
			response.Error(c, err)
			return
		}
		claims := c.MustGet("claims").(*mws.Claim)

		files := make([]*file.FileDownloadInfo, 0, len(req.Files))
		for _, f := range req.Files {
			files = append(files, &file.FileDownloadInfo{
				FileId:   f.FileId,
				OrderNum: f.OrderNum,
				Path:     f.Path,
			})
		}

		// 创建下载任务
		resp, err := h.cli.DownloadTask(c.Request.Context(), &file.DownloadTaskRequest{
			UserId:     claims.UserId,
			Files:      files,
			FolderName: req.FolderName,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// GetDownloadTask 获取下载任务状态
func (h *FileHandler) GetDownloadTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		taskId := c.Param("taskId")
		claims := c.MustGet("claims").(*mws.Claim)

		resp, err := h.cli.GetDownloadTask(c.Request.Context(), &file.GetDownloadTaskRequest{
			TaskId: taskId,
			UserId: claims.UserId,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// ResumeDownload 断点续传
func (h *FileHandler) ResumeDownload() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TaskId  string  `json:"taskId"`
			FileIds []int64 `json:"fileIds"` // 需要继续下载的文件ID列表
		}
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)

		resp, err := h.cli.ResumeDownload(c.Request.Context(), &file.ResumeDownloadRequest{
			TaskId:  req.TaskId,
			UserId:  claims.UserId,
			FileIds: req.FileIds,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// CreateFolder 创建文件夹
func (h *FileHandler) CreateFolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name     string `json:"name"`
			ParentId int64  `json:"parentId"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)
		resp, err := h.cli.CreateFolder(c.Request.Context(), &file.CreateFolderRequest{
			Name:     req.Name,
			ParentId: req.ParentId,
			UserId:   claims.UserId,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// ListFolder 获取文件夹内容
func (h *FileHandler) ListFolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			FolderId int64 `json:"folderId"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)
		resp, err := h.cli.ListFolder(c.Request.Context(), &file.ListFolderRequest{
			FolderId: req.FolderId,
			UserId:   claims.UserId,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// MoveFile 移动文件
func (h *FileHandler) MoveFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			FileId     int64 `json:"fileId"`
			ToFolderID int64 `json:"toFolderID"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)
		resp, err := h.cli.MoveFile(c.Request.Context(), &file.MoveFileRequest{
			UserId:     claims.UserId,
			FileId:     req.FileId,
			ToFolderId: req.ToFolderID,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// MoveFolder 移动文件夹
func (h *FileHandler) MoveFolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			FolderId   int64  `json:"folderId"`
			ToFolderID int64  `json:"toFolderID"`
			FolderName string `json:"folderName"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)
		resp, err := h.cli.MoveFolder(c.Request.Context(), &file.MoveFolderRequest{
			UserId:     claims.UserId,
			FolderId:   req.FolderId,
			ToFolderId: req.ToFolderID,
			FolderName: req.FolderName,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// DeleteFile 删除文件
func (h *FileHandler) DeleteFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			FileId int64 `json:"fileId"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)
		resp, err := h.cli.DeleteFile(c.Request.Context(), &file.DeleteFileRequest{
			FileId: req.FileId,
			UserId: claims.UserId,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// DeleteFolder 删除文件夹
func (h *FileHandler) DeleteFolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			FolderId int64 `json:"fileId"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)
		resp, err := h.cli.DeleteFolder(c.Request.Context(), &file.DeleteFolderRequest{
			FolderId: req.FolderId,
			UserId:   claims.UserId,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// SearchFiles 搜索文件
func (h *FileHandler) SearchFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Query string `json:"query"`
			Page  int32  `json:"page"`
			Size  int32  `json:"size"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)
		resp, err := h.cli.Search(c.Request.Context(), &file.SearchRequest{
			UserId: claims.UserId,
			Query:  req.Query,
			Page:   req.Page,
			Size:   req.Size,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

func (h *FileHandler) Preview() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		fileId, _ := strconv.Atoi(id)
		claims := c.MustGet("claims").(*mws.Claim)

		resp, err := h.cli.Preview(c.Request.Context(), &file.PreviewRequest{
			FileId: int64(fileId),
			UserId: claims.UserId,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		// 返回预览信息
		response.Success(c, resp)
	}
}

// 根据文件扩展名获取 MIME 类型
func getMimeType(ext string) string {
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

func setHeader(c *gin.Context, filename, mimeType string) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", url.QueryEscape(filename)))
	c.Header("Content-Type", mimeType)
	c.Header("Cache-Control", "no-cache")
}
