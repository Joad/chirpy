package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessType  string = "chirpy-access"
	RefreshType string = "chirpy-refresh"
)

func HashPassword(password string) (string, error) {
	result, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(id int, issuedAt, expiresAt time.Time, issuer, jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   fmt.Sprint(id),
		})
	return token.SignedString([]byte(jwtSecret))
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no Authorization header")
	}
	token, found := strings.CutPrefix(authHeader, "Bearer ")
	if !found {
		return "", errors.New("incorrect Authorization header")
	}
	return token, nil
}

func ValidatePolkaKey(headers http.Header, polkaKey string) error {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return errors.New("no Authorization header")
	}
	key, found := strings.CutPrefix(authHeader, "ApiKey ")
	if !found {
		return errors.New("no key in header")
	}

	if key != polkaKey {
		return errors.New("key invalid")
	}

	return nil
}

func ValidateJWT(token, jwtSecret string) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil },
	)
	if err != nil {
		return "", err
	}

	issuer, err := parsedToken.Claims.GetIssuer()
	if err != nil {
		return "", err
	}

	if issuer != AccessType {
		return "", errors.New("invalid issuer")
	}

	subject, err := parsedToken.Claims.GetSubject()
	if err != nil {
		return "", err
	}
	return subject, err
}

func RefreshToken(token, jwtSecret string) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil },
	)
	if err != nil {
		return "", err
	}

	issuer, err := parsedToken.Claims.GetIssuer()
	if err != nil {
		return "", err
	}

	if issuer != RefreshType {
		return "", errors.New("invalid issuer")
	}

	subject, err := parsedToken.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	id, err := strconv.Atoi(subject)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Hour)
	return MakeJWT(
		id,
		now,
		expiresAt,
		AccessType,
		jwtSecret,
	)
}
