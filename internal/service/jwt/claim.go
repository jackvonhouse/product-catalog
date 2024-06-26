package jwt

import "github.com/golang-jwt/jwt/v5"

type AccessTokenClaim struct {
	Username       string `json:"username"`
	RefreshTokenId int    `json:"refresh_token_id"`

	jwt.RegisteredClaims
}
