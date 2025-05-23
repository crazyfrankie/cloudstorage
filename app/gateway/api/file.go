package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crazyfrankie/gem/gerrors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
		fileGroup.POST("/upload/chunk", h.UploadChunk())
		fileGroup.POST("/update", h.UpdateFile())
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
		fileGroup.POST("/share", h.CreateShareLink())
		fileGroup.POST("/save", h.SaveToMyDrive())
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

		if size > consts.SmallFileSizeLimit {
			response.Error(c, gerrors.NewBizError(40001, "Payload Too Large"))
			return
		}

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

// UploadChunk v2
func (h *FileHandler) UploadChunk() gin.HandlerFunc {
	return func(c *gin.Context) {
		f, header, err := c.Request.FormFile("file")
		if err != nil {
			response.Error(c, err)
			return
		}
		defer f.Close()

		// 获取其他参数
		folder := c.PostForm("folder")
		folderId, _ := strconv.Atoi(folder)
		claims := c.MustGet("claims").(*mws.Claim)

		stream, err := h.cli.UploadChunkStream(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		buffer := make([]byte, consts.ChunkSize)
		var partNumber int32 = 0

		for {
			n, err := f.Read(buffer)
			if n > 0 {
				partNumber++
				if err := stream.Send(&file.UploadChunkRequest{
					Filename:   header.Filename,
					PartNumber: partNumber,
					Data:       buffer[:n],
					FileSize:   header.Size,
					UserId:     claims.UserId,
					FolderId:   int64(folderId),
				}); err != nil {
					response.Error(c, fmt.Errorf("failed to send chunk: %v", err))
					return
				}
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				response.Error(c, fmt.Errorf("failed to read file: %v", err))
				return
			}
		}

		resp, err := stream.CloseAndRecv()
		if err != nil {
			response.Error(c, fmt.Errorf("failed to close stream: %v", err))
			return
		}

		response.Success(c, resp)
	}
}

// UpdateFile 更新文件
func (h *FileHandler) UpdateFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileId, _ := strconv.ParseInt(c.PostForm("fileId"), 10, 64)
		deviceId := c.PostForm("deviceId")
		if deviceId == "" {
			response.Error(c, errors.New("deviceId is required"))
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)

		// 解析基础版本，这是用于增量更新的必要参数
		baseVersion, err := strconv.ParseInt(c.PostForm("baseVersion"), 10, 64)
		if err != nil {
			baseVersion = 0 // 如果未提供，默认为0
		}

		// 检查是否为增量更新
		isIncremental := c.PostForm("isIncremental") == "true"

		// 准备请求对象
		req := &file.UpdateFileRequest{
			FileId:        fileId,
			UserId:        claims.UserId,
			DeviceId:      deviceId,
			BaseVersion:   baseVersion,
			IsIncremental: isIncremental,
		}

		// 如果提供了新文件名，添加到请求中
		if newName := c.PostForm("name"); newName != "" {
			req.Name = newName
		}

		if isIncremental {
			var changeList []struct {
				Operation string `json:"operation"`
				Position  int64  `json:"position"`
				Length    int64  `json:"length"`
				Content   string `json:"content"`
			}

			changesStr := c.PostForm("changes")
			if changesStr != "" {
				if err := json.Unmarshal([]byte(changesStr), &changeList); err != nil {
					response.Error(c, fmt.Errorf("invalid changes format: %v", err))
					return
				}

				changes := make([]*file.FileChange, 0, len(changeList))
				for _, ch := range changeList {
					var op file.ChangeOperation
					switch strings.ToUpper(ch.Operation) {
					case "INSERT":
						op = file.ChangeOperation_INSERT
					case "DELETE":
						op = file.ChangeOperation_DELETE
					case "UPDATE":
						op = file.ChangeOperation_UPDATE
					default:
						response.Error(c, fmt.Errorf("unknown operation type: %s", ch.Operation))
						return
					}

					changes = append(changes, &file.FileChange{
						Operation: op,
						Position:  ch.Position,
						Length:    ch.Length,
						Content:   []byte(ch.Content),
					})
				}

				req.Changes = changes
			}
		} else {
			// 处理全量更新
			f, _, err := c.Request.FormFile("file")
			if err != nil {
				response.Error(c, err)
				return
			}
			defer f.Close()

			data, err := io.ReadAll(f)
			if err != nil {
				response.Error(c, err)
				return
			}

			req.Data = data
		}

		resp, err := h.cli.UpdateFile(c.Request.Context(), req)
		if err != nil {
			response.Error(c, err)
			return
		}

		if resp.HasConflict {
			c.JSON(http.StatusConflict, gin.H{
				"code": 40901,
				"msg":  resp.ConflictMessage,
				"data": gin.H{
					"currentVersion": resp.CurrentVersion,
					"neededChanges":  resp.NeededChanges,
					"file":           resp.File,
				},
			})
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

// Preview 文件预览
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

// CreateShareLink 创建分享链接
func (h *FileHandler) CreateShareLink() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FileIds    []int64 `json:"fileIds"`
			FolderId   int64   `json:"folderId"`
			ExpireDays int32   `json:"expireDays"`
			Password   string  `json:"password"`
		}
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)
		resp, err := h.cli.CreateShareLink(c.Request.Context(), &file.CreateShareLinkRequest{
			UserId:     claims.UserId,
			FileIds:    req.FileIds,
			FolderId:   req.FolderId,
			ExpireDays: req.ExpireDays,
			Password:   req.Password,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// SaveToMyDrive 保存到我的网盘
func (h *FileHandler) SaveToMyDrive() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ShareId    string  `json:"shareId"`
			Password   string  `json:"password"`
			ToFolderId int64   `json:"toFolderId"`
			FileIds    []int64 `json:"fileIds"`
		}
		if err := c.Bind(&req); err != nil {
			return
		}

		claims := c.MustGet("claims").(*mws.Claim)
		resp, err := h.cli.SaveToMyDrive(c.Request.Context(), &file.SaveToMyDriveRequest{
			ShareId:    req.ShareId,
			Password:   req.Password,
			UserId:     claims.UserId,
			ToFolderId: req.ToFolderId,
			FileIds:    req.FileIds,
		})
		if err != nil {
			response.Error(c, err)
			return
		}

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
