package auth

import "github.com/golang-jwt/jwt"

// UserClaimsInfo custom jwt claims
type UserClaimsInfo struct {
	SubType      string `json:"sub_type"`
	UserName     string `json:"username"`
	BKAppCode    string `json:"bk_app_code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	// https://tools.ietf.org/html/rfc7519#section-4.1
	// aud: 接收jwt一方; exp: jwt过期时间; jti: jwt唯一身份认证; IssuedAt: 签发时间; Issuer: jwt签发者
	*jwt.StandardClaims
}