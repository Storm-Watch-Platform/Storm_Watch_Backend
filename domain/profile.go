package domain

import "context"

type Profile struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	ID    string `json:"id"`
}

type ProfileUsecase interface {
	GetProfileByID(c context.Context, userID string) (*Profile, error)
}
