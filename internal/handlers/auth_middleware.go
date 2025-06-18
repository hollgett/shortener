package handlers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const (
	token_exp         = 31 * (24 * time.Hour)
	len_user_id       = 8
	token_user_cookie = "uid"
)

var (
	ErrNoCookie error = errors.New("cookie not found")
)

type ctxKey string

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type User struct {
	ID  string
	Err error
}

var (
	UserKeyCtx ctxKey     = "UserID"
	pseudoRand *rand.Rand = rand.New(rand.NewSource(time.Now().Unix()))
)

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(token_user_cookie)
		if err != nil {
			m.logger.Info("get cookie", zap.Error(err))
			userID, errCook := m.SetUserCookie(w)
			if errCook != nil {
				m.logger.Info("set user cookie", zap.Error(err))
				http.Error(w, fmt.Sprintf("set user cookie: %s", err.Error()), http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, SetContext(r, userID, ErrNoCookie))
			return
		}
		userID, err := m.GetUserID(cookie.Value)
		if err != nil {
			m.logger.Info("get user ID", zap.Error(err))
			userID, errCook := m.SetUserCookie(w)
			if errCook != nil {
				m.logger.Info("set user cookie", zap.Error(err))
				http.Error(w, fmt.Sprintf("set user cookie: %s", err.Error()), http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, SetContext(r, userID, nil))
			return
		}
		next.ServeHTTP(w, SetContext(r, userID, nil))
	})
}

func (m *Middleware) buildJWTString() (token string, userID string, err error) {
	userID = generateUserID()
	tokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(token_exp)),
		},
		UserID: userID,
	})

	token, err = tokenJWT.SignedString([]byte(m.secretKey))
	if err != nil {
		err = fmt.Errorf("failed signing token: %w", err)
		return
	}
	return
}

func (m *Middleware) GetUserID(tokenStr string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(m.secretKey), nil
		})
	if err != nil {
		return "", fmt.Errorf("failed parse token: %w", err)
	}

	if !token.Valid {
		return "", errors.New("token is not valid")
	}

	return claims.UserID, nil
}

func generateUserID() string {
	userID := make([]byte, len_user_id)
	for i := range userID {
		if pseudoRand.Intn(2) == 0 {
			userID[i] = byte(pseudoRand.Intn(25) + 97)
		} else {
			userID[i] = byte(pseudoRand.Intn(10) + 48)
		}
	}
	return string(userID)
}

func (m *Middleware) SetUserCookie(w http.ResponseWriter) (string, error) {
	token, userID, err := m.buildJWTString()
	if err != nil {
		return "", fmt.Errorf("failed build jwt: %w", err)
	}
	NewCookie := &http.Cookie{
		Name:  token_user_cookie,
		Value: token,
		Path:  "/",
	}
	http.SetCookie(w, NewCookie)
	return userID, nil
}

func SetContext(r *http.Request, userID string, err error) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), UserKeyCtx, User{
		ID:  userID,
		Err: err,
	}))
}
