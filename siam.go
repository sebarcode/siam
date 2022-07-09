package siam

import (
	"github.com/golang-jwt/jwt"
	"github.com/sebarcode/codekit"
)

type AuthJwt struct {
	jwt.StandardClaims
	Data codekit.M
}
