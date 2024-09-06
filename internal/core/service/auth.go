package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/sarthak014/Fast-Bank/internal/core/domain"
	"github.com/sarthak014/Fast-Bank/internal/core/port"
)

type authService struct {
	secretKey []byte
}

func NewAuthService(secretKey string) port.AuthService {
	return &authService{secretKey: []byte(secretKey)}
}

func (s *authService) Generate(userId int) (string, error) {
	claims := domain.JWTClaims{
		Id: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 2)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *authService) Validate(tokenString string) (*domain.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*domain.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *authService) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.ErrUnauthorized
		}

		tokenString := authHeader[7:] //bearer
		claims, err := s.Validate(tokenString)
		if err != nil {
			return echo.ErrUnauthorized
		}

		c.Set("user", claims)
		return next(c)
	}
}
