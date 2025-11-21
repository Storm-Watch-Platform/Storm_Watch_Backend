package domain

import (
	"context"
)

type LoginRequest struct {
	Phone    string `form:"phone" binding:"required"`
	Password string `form:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserID       string `json:"userID"`
}

type LoginUsecase interface {
	GetUserByPhone(c context.Context, phone string) (User, error)
	CreateAccessToken(user *User, secret string, expiry int) (accessToken string, err error)
	CreateRefreshToken(user *User, secret string, expiry int) (refreshToken string, err error)
	// GetGroupsOfUser(userID string) ([]string, error)
}
