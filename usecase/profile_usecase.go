package usecase

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
)

type profileUsecase struct {
	userRepository domain.UserRepository
	contextTimeout time.Duration
}

func NewProfileUsecase(userRepository domain.UserRepository, timeout time.Duration) domain.ProfileUsecase {
	return &profileUsecase{
		userRepository: userRepository,
		contextTimeout: timeout,
	}
}

func (pu *profileUsecase) GetProfileByID(c context.Context, userID string) (*domain.Profile, error) {
	ctx, cancel := context.WithTimeout(c, pu.contextTimeout)
	defer cancel()

	// Get User from repository
	user, err := pu.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert ObjectIDs to strings
	groupStr := make([]string, len(user.GroupIDs))
	for i, id := range user.GroupIDs {
		groupStr[i] = id.Hex()
	}

	profile := &domain.Profile{
		Name:  user.Name,
		Phone: user.Phone,
		ID:    user.ID.Hex(),
		Group: groupStr,
	}

	return profile, nil
}
