package mws

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

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

func (b *AuthBuilder) Auth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := b.paths[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		tokenHeader := r.Header.Get("Authorization")
		token := extractToken(tokenHeader)

		claims, err := parseToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized"))
			return
		}

		uid := strconv.Itoa(int(claims.UserId))
		r = r.WithContext(context.WithValue(r.Context(), "user_id", uid))

		next.ServeHTTP(w, r)
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
