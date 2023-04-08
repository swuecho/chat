package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

func CheckPermission(userID int, ctx context.Context) bool {
	contextUserID, ok := ctx.Value("user_id").(int)
	if !ok {
		return false
	}
	role, ok := ctx.Value("role").(string)
	if !ok {
		return false
	}

	switch role {
	case "admin":
		return true
	case "member":
		return userID == contextUserID
	default:
		return false
	}
}

func extractBearerToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	tokenParts := strings.Split(bearerToken, " ")
	if len(tokenParts) == 2 {
		return tokenParts[1]
	}
	return ""
}

type contextKey string

const (
	roleContextKey contextKey = "role"
	userContextKey contextKey = "user"
	guidContextKey contextKey = "guid"
)

func IsAuthorizedMiddleware(handler http.Handler) http.Handler {
	jwtSigningKey := []byte(appConfig.JWT.SECRET)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/verify" || r.URL.Path == "/login" || r.URL.Path == "/signup" || strings.HasPrefix(r.URL.Path, "/static") {
			handler.ServeHTTP(w, r)
			return
		}

		bearerToken := extractBearerToken(r)
		if bearerToken != "" {
			token, err := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("there was an error in jwt method")
				}
				return jwtSigningKey, nil
			})

			if err != nil {
				fmt.Fprint(w, err.Error())
				return
			}

			if token.Valid {
				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					log.Println("can not get claims")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				userID, ok := claims["user_id"].(string)
				if !ok {
					log.Println("can not get user id")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				role, ok := claims["role"].(string)
				if !ok {
					log.Println("can not get user role")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), userContextKey, userID)
				ctx = context.WithValue(ctx, roleContextKey, role)
				// superuser
				if strings.HasPrefix(r.URL.Path, "/admin") && role != "admin" {
					w.WriteHeader(http.StatusForbidden)
					RespondWithError(w, http.StatusForbidden, "error.NotAdmin", "Not Admin")
					return
				}

				// TODO: get trace id and add it to context
				//traceID := r.Header.Get("X-Request-Id")
				//if len(traceID) > 0 {
				//ctx = context.WithValue(ctx, guidContextKey, traceID)
				//}
				// Store user ID and role in the request context
				// pass token to request
				handler.ServeHTTP(w, r.WithContext(ctx))
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			RespondWithError(w, http.StatusUnauthorized, "error.NotAuthorized", "Not Authorized")
		}
	})
}
