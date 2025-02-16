package common

import (
	"github.com/crazyfrankie/gem/gerrors"
)

const (
	BaseURL = "https://cloud.crazyfrank.top/file/"
)

var (
	SuccessMsg     = gerrors.NewBizError(00000, "success")
	InternalError  = gerrors.NewBizError(50000, "internal error")
	FileNameExists = gerrors.NewBizError(10000, "file name conflict")
)
