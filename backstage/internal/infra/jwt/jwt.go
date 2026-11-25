package infrajwt

import (
	"errors"
	"time"
)

const (
	defaultAccessTTL = time.Minute * 15
	TokenTypeAccess  = "access"
)

var ErrEmptyJWTSecret = errors.New("jwt secret is required")
var ErrParseAccessTTL = errors.New("parse jwt access_ttl failed")
var ErrEmptyToken = errors.New("token is empty")
var ErrParseJWTToken = errors.New("parse jwt token failed")
var ErrInvalidTokenType = errors.New("token type invalid")
var ErrInvalidTokenUserID = errors.New("token user id invalid")
var ErrInvalidUserID = errors.New("user id must be positive")
var ErrGenerateTokenJTI = errors.New("generate token jti failed")
var ErrSignJWTToken = errors.New("sign jwt token failed")
var ErrInvalidTTL = errors.New("ttl must be positive")
