package api

import (
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloudstorage/app/gateway/common/consts"
	"github.com/crazyfrankie/cloudstorage/app/gateway/common/response"
	"github.com/crazyfrankie/cloudstorage/app/gateway/common/util"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
)

type FileHandler struct {
	cli file.FileServiceClient
}

func NewFileHandler(cli file.FileServiceClient) *FileHandler {
	return &FileHandler{cli: cli}
}

func (h *FileHandler) RegisterRoute(r *gin.Engine) {
	fileGroup := r.Group("/api/files")
	{
		fileGroup.POST("/upload", h.Upload())
	}
}

// Upload v1 暂时只实现小文件的上传
// TODO 大文件上传也即大文件分块上传以及随之而来的断点续传功能
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

		meta := &file.FileMetaData{
			Name:        name,
			Path:        path,
			Hash:        hash,
			Size:        size,
			ContentType: typ,
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
