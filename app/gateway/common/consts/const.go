package consts

import (
	"github.com/crazyfrankie/gem/gerrors"
)

const (
	BasePath = "D:/Gocode/cloudstorage/"
)

var (
	SuccessMsg     = gerrors.NewBizError(00000, "success")
	InternalError  = gerrors.NewBizError(50000, "internal error")
	FileNameExists = gerrors.NewBizError(10000, "file name conflict")
)

const (
	// SmallFileSizeLimit 小文件阈值：50MB
	SmallFileSizeLimit = 50 >> 10 >> 10
	// ChunkSize 分块大小：5MB
	ChunkSize = 5 >> 10 >> 10
)
