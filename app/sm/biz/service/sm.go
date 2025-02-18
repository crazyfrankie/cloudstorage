package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/crazyfrankie/cloudstorage/app/sm/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/sm/biz/service/sms"
	"github.com/crazyfrankie/cloudstorage/app/sm/config"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/sm"
)

type SmServer struct {
	repo *repository.SmRepo
	sms  sms.Service
	sm.UnimplementedShortMsgServiceServer
}

func NewSmServer(repo *repository.SmRepo, sms sms.Service) *SmServer {
	return &SmServer{repo: repo, sms: sms}
}

func (s *SmServer) SendSm(ctx context.Context, req *sm.SendSmRequest) (*sm.SendSmResponse, error) {
	code := generateCode()
	hash := generateHMAC(code, config.GetConf().SMS.Secret)

	err := s.repo.Store(ctx, req.GetBiz(), req.GetPhone(), hash)
	if err != nil {
		return nil, err
	}

	// Send
	err = s.sms.Send(ctx, config.GetConf().SMS.TemplateID, []string{code}, req.GetPhone())

	return &sm.SendSmResponse{}, err
}

func (s *SmServer) VerifySm(ctx context.Context, req *sm.VerifySmRequest) (*sm.VerifySmResponse, error) {
	encode := generateHMAC(req.GetCode(), config.GetConf().SMS.Secret)

	err := s.repo.Verify(ctx, req.GetBiz(), req.GetPhone(), encode)
	if err != nil {
		return nil, err
	}

	return &sm.VerifySmResponse{}, err
}

func generateCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	var code strings.Builder
	for i := 0; i < 6; i++ {
		digit := rand.Intn(10)
		code.WriteString(strconv.Itoa(digit))
	}

	return code.String()
}

func generateHMAC(code, key string) string {
	h := hmac.New(sha256.New, []byte(key))

	h.Write([]byte(code))

	return hex.EncodeToString(h.Sum(nil))
}
