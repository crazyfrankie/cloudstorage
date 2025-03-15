package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string
)

type SmCache struct {
	cmd redis.Cmdable
}

func NewSmCache(cmd redis.Cmdable) *SmCache {
	return &SmCache{cmd: cmd}
}

func (c *SmCache) Store(ctx context.Context, biz, phone, code string) error {
	key := c.key(biz, phone)

	res, err := c.cmd.Eval(ctx, luaSetCode, []string{key}, code).Int()

	if err != nil {
		return err
	}

	switch res {
	case 0:
		return nil
	case -1:
		// 发送太频繁
		return errors.New("send too many")
	default:
		return errors.New("internal")
	}
}

func (c *SmCache) Verify(ctx context.Context, biz, phone, inputCode string) error {
	key := c.key(biz, phone)

	ok, err := c.cmd.Eval(ctx, luaVerifyCode, []string{key}, inputCode).Int()
	if err != nil {
		return err
	}
	switch ok {
	case 0:
		return nil
	case -1:
		return errors.New("verify too many")
	}
	return errors.New("internal")
}

func (c *SmCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
