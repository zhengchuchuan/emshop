package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// EmshopClaims defines emshop business-specific JWT claims
type EmshopClaims struct {
	ID          uint `json:"userid"`
	AuthorityId uint `json:"authority_id"`
	jwt.RegisteredClaims
}

// GetExpirationTime implements jwt.Claims interface
func (c EmshopClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.ExpiresAt, nil
}

// GetIssuedAt implements jwt.Claims interface
func (c EmshopClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.IssuedAt, nil
}

// GetNotBefore implements jwt.Claims interface
func (c EmshopClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.NotBefore, nil
}

// GetIssuer implements jwt.Claims interface
func (c EmshopClaims) GetIssuer() (string, error) {
	return c.RegisteredClaims.Issuer, nil
}

// GetSubject implements jwt.Claims interface
func (c EmshopClaims) GetSubject() (string, error) {
	return c.RegisteredClaims.Subject, nil
}

// GetAudience implements jwt.Claims interface
func (c EmshopClaims) GetAudience() (jwt.ClaimStrings, error) {
	return c.RegisteredClaims.Audience, nil
}

// SetExpiresAt updates the expiration time (helper for refresh token)
func (c *EmshopClaims) SetExpiresAt(exp *jwt.NumericDate) {
	c.RegisteredClaims.ExpiresAt = exp
}