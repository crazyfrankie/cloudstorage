package api

import (
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloudstorage/rpc_gen/user"
)

type UserHandler struct {
	cli user.UserServiceClient
}

func NewUserHandler(cli user.UserServiceClient) *UserHandler {
	return &UserHandler{cli: cli}
}

func (h *UserHandler) RegisterRoute(r *gin.Engine) {
	userGroup := r.Group("api/user")
	{
		userGroup.POST("/send-code", h.SendCode())
		userGroup.POST("/verify-code", h.VerifyCode())
	}
}

func (h *UserHandler) SendCode() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func (h *UserHandler) VerifyCode() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func (h *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
