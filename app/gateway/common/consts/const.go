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
	// SmallFileSizeLimit 小文件阈值：20MB
	SmallFileSizeLimit = 20 * 1024 * 1024
	// ChunkSize 分块大小：5MB
	ChunkSize = 5 * 1024 * 1024
)
