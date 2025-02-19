package api

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloudstorage/app/gateway/common/response"
	"github.com/crazyfrankie/cloudstorage/app/gateway/mws"
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
		userGroup.GET("/info", h.GetUserInfo(), mws.Auth())
	}
}

func (h *UserHandler) SendCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Phone string `json:"phone"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}
		fmt.Println(req)
		resp, err := h.cli.SendCode(c.Request.Context(), &user.SendCodeRequest{Phone: req.Phone})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

func (h *UserHandler) VerifyCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Phone string `json:"phone"`
			Code  string `json:"code"`
			Biz   string `json:"biz"`
		}
		var req Req
		if err := c.Bind(&req); err != nil {
			return
		}

		resp, err := h.cli.VerifyCode(c.Request.Context(), &user.VerifyCodeRequest{Phone: req.Phone, Code: req.Code, Biz: req.Biz})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}

func (h *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims")
		claim, _ := claims.(*mws.Claim)

		resp, err := h.cli.GetUserInfo(c.Request.Context(), &user.GetUserInfoRequest{UserId: claim.UserId})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, resp)
	}
}
