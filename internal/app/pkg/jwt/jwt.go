package jwt

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
	"emshop/gin-micro/server/rest-server/middlewares"
)

// EmshopJWT provides business-specific JWT operations for emshop
type EmshopJWT struct {
	jwt *middlewares.JWT[*EmshopClaims]
}

// NewEmshopJWT creates a new business JWT instance
func NewEmshopJWT(signKey string) *EmshopJWT {
	return &EmshopJWT{
		jwt: middlewares.NewJWT[*EmshopClaims](signKey),
	}
}

// NewEmshopJWTWithValidator creates a JWT instance with custom validator
func NewEmshopJWTWithValidator(signKey string, validator middlewares.ClaimsValidator) *EmshopJWT {
	return &EmshopJWT{
		jwt: middlewares.NewJWTWithValidator[*EmshopClaims](signKey, validator),
	}
}

// CreateToken creates a JWT token with emshop claims
func (ej *EmshopJWT) CreateToken(userID, authorityID uint, issuer string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &EmshopClaims{
		ID:          userID,
		AuthorityId: authorityID,
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
		},
	}
	
	return ej.jwt.CreateToken(claims)
}

// ParseToken parses and validates an emshop JWT token
func (ej *EmshopJWT) ParseToken(tokenString string) (*EmshopClaims, error) {
	claims := &EmshopClaims{}
	return ej.jwt.ParseToken(tokenString, claims)
}

// RefreshToken refreshes an existing JWT token with new expiration
func (ej *EmshopJWT) RefreshToken(tokenString string, newExpiry time.Duration) (string, error) {
	claims := &EmshopClaims{}
	return ej.jwt.RefreshToken(tokenString, claims, newExpiry)
}

// ExtractUserInfo extracts user information from emshop claims
func (ej *EmshopJWT) ExtractUserInfo(claims *EmshopClaims) (userID, authorityID uint) {
	return claims.ID, claims.AuthorityId
}

// ValidateAuthority validates if the user has required authority level
func (ej *EmshopJWT) ValidateAuthority(claims *EmshopClaims, minAuthority uint) bool {
	return claims.AuthorityId >= minAuthority
}