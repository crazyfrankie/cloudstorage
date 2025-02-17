package util

import (
	"crypto/md5"
	"fmt"
	"io"
)

func FileHash(file io.Reader) (string, error) {
	hash := md5.New() // 使用 MD5 哈希算法
	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	// 返回哈希值的十六进制字符串
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
