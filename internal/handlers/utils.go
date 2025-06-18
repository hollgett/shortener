package handlers

import (
	"errors"
	"net/http"
)

func parseUserID(r *http.Request) (User, error) {
	userID := r.Context().Value(UserKeyCtx)
	val, ok := userID.(User)
	if !ok {
		val.Err = errors.New("failed get user id")
		return val, errors.New("failed get user id")
	}

	return val, nil
}
