package common

import (
	"net/http"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessMsg.BizStatusCode(),
		Message: SuccessMsg.BizMessage(),
		Data:    data,
	})
}

func Error(c *gin.Context, err error) {
	if bizErr, ok := gerrors.FromBizStatusError(err); ok {
		c.JSON(http.StatusOK, Response{
			Code:    bizErr.BizStatusCode(),
			Message: bizErr.BizMessage(),
		})
	}

	c.JSON(http.StatusOK, Response{
		Code:    InternalError.BizStatusCode(),
		Message: InternalError.BizMessage(),
	})
}
