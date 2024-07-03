package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func (s *ApiServer) handleLogin(c echo.Context) error {
	payload := new(struct {
		Id       int    `json:"id"`
		Password string `json:"password"`
	})
	if err := c.Bind(payload); err != nil {
		return err
	}

	user, err := s.store.GetAccountById(payload.Id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.EPassword), []byte(payload.Password)); err != nil {
		return echo.ErrUnauthorized
	}

	token, err := generateJWT(user.Id)
	if err != nil {
		return err
	}

	return c.JSON(200, map[string]string{
		"token": token,
	})
}

func generateJWT(userId int) (string, error) {
	claims := JWTClaims{
		Id:  userId,
		Exp: time.Now().Add(time.Hour * 24).Unix(),
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signature := generateSignature(headerEncoded + "." + claimsEncoded)

	return headerEncoded + "." + claimsEncoded + "." + signature, nil
}

func generateSignature(data string) string {
	h := hmac.New(sha256.New, []byte("secretSHHHH"))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			log.Println("No Authorization header found")
			return echo.ErrUnauthorized
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Println("Invalid Authorization header format")
			return echo.ErrUnauthorized
		}

		token := tokenParts[1]
		claims, err := validateJWT(token)
		if err != nil {
			log.Printf("JWT validation failed: %v", err)
			return echo.ErrUnauthorized
		}

		c.Set("user", claims)
		return next(c)
	}
}
func validateJWT(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	signature := generateSignature(parts[0] + "." + parts[1])
	if signature != parts[2] {
		log.Println("Invalid signature")
		return nil, errors.New("invalid signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Printf("Failed to decode claims: %v", err)
		return nil, err
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		log.Printf("Failed to unmarshal claims: %v", err)
		return nil, err
	}

	if time.Now().Unix() > claims.Exp {
		log.Println("Token has expired")
		return nil, errors.New("token expired")
	}

	return &claims, nil
}
