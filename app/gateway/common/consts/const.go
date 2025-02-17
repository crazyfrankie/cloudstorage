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
