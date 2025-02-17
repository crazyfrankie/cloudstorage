package mws

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	SecretKey = []byte("tN7tZ2oQ3mB2nI3gD3qD3tX0yG3qQ6cE")
)

type Claim struct {
	UserId int32
	jwt.MapClaims
}

type AuthBuilder struct {
	paths map[string]struct{}
}

func NewAuthBuilder() *AuthBuilder {
	return &AuthBuilder{
		paths: make(map[string]struct{}),
	}
}

func (b *AuthBuilder) IgnorePath(path string) *AuthBuilder {
	b.paths[path] = struct{}{}
	return b
}

func (b *AuthBuilder) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := b.paths[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		tokenHeader := c.GetHeader("Authorization")
		token := extractToken(tokenHeader)

		claims, err := parseToken(token)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("claims", claims)

		c.Next()
	}
}

func extractToken(token string) string {
	if token == "" {
		return ""
	}

	strs := strings.Split(token, " ")
	if strs[0] == "Bearer" {
		return strs[1]
	}

	return ""
}

func parseToken(token string) (*Claim, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claim{}, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claim); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, errors.New("token is invalid")
}
