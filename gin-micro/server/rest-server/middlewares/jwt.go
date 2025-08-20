package middlewares

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// ClaimsValidator defines interface for custom claims validation
type ClaimsValidator interface {
	Validate(claims jwt.Claims) error
}

// JWT provides generic JWT operations for any claims type that implements jwt.Claims
type JWT[T jwt.Claims] struct {
	SigningKey []byte
	Validator  ClaimsValidator // optional custom validator
}

var (
	TokenExpired     = errors.New("Token is expired")
	TokenNotValidYet = errors.New("Token not active yet")
	TokenMalformed   = errors.New("That's not even a token")
	TokenInvalid     = errors.New("Couldn't handle this token:")
)

// NewJWT creates a new generic JWT instance
func NewJWT[T jwt.Claims](signKey string) *JWT[T] {
	return &JWT[T]{
		SigningKey: []byte(signKey),
	}
}

// NewJWTWithValidator creates a JWT instance with custom validator
func NewJWTWithValidator[T jwt.Claims](signKey string, validator ClaimsValidator) *JWT[T] {
	return &JWT[T]{
		SigningKey: []byte(signKey),
		Validator:  validator,
	}
}

// CreateToken creates a JWT token with the provided claims
func (j *JWT[T]) CreateToken(claims T) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// ParseToken parses and validates a JWT token
func (j *JWT[T]) ParseToken(tokenString string, claimsPtr T) (T, error) {
	var zero T
	
	token, err := jwt.ParseWithClaims(tokenString, claimsPtr, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	
	if err != nil {
		// Handle different types of JWT errors in v5
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return zero, TokenMalformed
		case errors.Is(err, jwt.ErrTokenExpired):
			return zero, TokenExpired
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return zero, TokenNotValidYet
		default:
			return zero, TokenInvalid
		}
	}
	
	if token == nil || !token.Valid {
		return zero, TokenInvalid
	}
	
	// Apply custom validation if validator is set
	if j.Validator != nil {
		if err := j.Validator.Validate(token.Claims); err != nil {
			return zero, err
		}
	}
	
	return claimsPtr, nil
}

// RefreshToken refreshes an existing JWT token with new expiration time
func (j *JWT[T]) RefreshToken(tokenString string, claimsPtr T, newExpiry time.Duration) (string, error) {
	// In v5, we use parser options instead of global TimeFunc
	// Parse with custom options to allow expired tokens for refresh
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(tokenString, claimsPtr, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	
	if err != nil {
		return "", err
	}
	
	if token == nil {
		return "", TokenInvalid
	}
	
	// Update expiration time using reflection or type assertion
	// This is a simplified approach - in practice, you'd want to handle this more elegantly
	if regClaims, ok := token.Claims.(interface{ SetExpiresAt(*jwt.NumericDate) }); ok {
		regClaims.SetExpiresAt(jwt.NewNumericDate(time.Now().Add(newExpiry)))
	}
	
	return j.CreateToken(claimsPtr)
}
