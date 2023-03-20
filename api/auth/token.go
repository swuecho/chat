package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	uuid "github.com/iris-contrib/go.uuid"
)

var ErrInvalidToken = errors.New("invalid token")

func GenerateToken(userID int32, secret string) (string, error) {
	expires := time.Now().Add(time.Hour * 8).Unix()
	notBefore := time.Now().Unix()
	issuer := "https://www.bestqa.net"

	claims := jwt.MapClaims{
		"user_id": strconv.FormatInt(int64(userID), 10),
		"exp":     expires,
		"role":    "user",
		"jti":     uuid.Must(uuid.NewV4()).String(),
		"iss":     issuer,
		"nbf":     notBefore,
		"aud":     "bc442bb2b1d848fba5be2aa24312e711",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "dfsafdsafdsafadsfdasdfs"
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateToken(tokenString string, secret string) (int32, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing algorithm and secret key used to sign the token
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token)
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return 0, ErrInvalidToken
	}

	i, err := strconv.Atoi(userIDStr)
	if err != nil {
		return -1, err
	}

	return int32(i), nil
}

func GetExpireSecureCookie(value string, isHttps bool) *http.Cookie {
	utcOffset := time.Now().UTC().Add(-24 * time.Hour)
	options := &http.Cookie{
		Name:     "jwt",
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   isHttps,
		SameSite: http.SameSiteStrictMode,
		Expires:  utcOffset,
	}
	return options
}
