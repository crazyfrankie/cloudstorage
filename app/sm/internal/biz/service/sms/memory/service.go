package memory

import (
	"context"
	"fmt"

	"github.com/crazyfrankie/cloudstorage/app/sm/internal/biz/service/sms"
)

type MemorySmService struct {
}

func NewMemorySmService() sms.Service {
	return &MemorySmService{}
}

func (m *MemorySmService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	fmt.Println(args)
	return nil
}
