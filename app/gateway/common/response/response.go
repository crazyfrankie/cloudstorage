package response

import (
	"github.com/crazyfrankie/cloudstorage/app/gateway/common/consts"
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
		Code:    consts.SuccessMsg.BizStatusCode(),
		Message: consts.SuccessMsg.BizMessage(),
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
		Code:    consts.InternalError.BizStatusCode(),
		Message: consts.InternalError.BizMessage(),
	})
}
