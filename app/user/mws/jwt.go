package mws

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/crazyfrankie/cloudstorage/app/user/config"
)

type Claim struct {
	UserId int32
	jwt.MapClaims
}

func GenerateToken(uid int32) (string, error) {
	now := time.Now()
	claims := &Claim{
		UserId: uid,
		MapClaims: jwt.MapClaims{
			"expire_at": now.Add(time.Hour * 24),
			"issue_at":  now.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(config.GetConf().JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
