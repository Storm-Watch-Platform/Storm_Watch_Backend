package usecase

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/tokenutil"
)

type loginUsecase struct {
	userRepository domain.UserRepository
	contextTimeout time.Duration
}

func NewLoginUsecase(userRepository domain.UserRepository, timeout time.Duration) domain.LoginUsecase {
	return &loginUsecase{
		userRepository: userRepository,
		contextTimeout: timeout,
	}
}

func (lu *loginUsecase) GetUserByPhone(c context.Context, phone string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(c, lu.contextTimeout)
	defer cancel()
	return lu.userRepository.GetByPhone(ctx, phone)
}

func (lu *loginUsecase) CreateAccessToken(user *domain.User, secret string, expiry int) (accessToken string, err error) {
	return tokenutil.CreateAccessToken(user, secret, expiry)
}

func (lu *loginUsecase) CreateRefreshToken(user *domain.User, secret string, expiry int) (refreshToken string, err error) {
	return tokenutil.CreateRefreshToken(user, secret, expiry)
}

// func (lu *loginUsecase) GetGroupsOfUser(userID string) ([]string, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), lu.contextTimeout)
// 	defer cancel()

// 	user, err := lu.userRepository.GetByID(ctx, userID) // giả sử repo có GetByID
// 	if err != nil {
// 		return nil, err
// 	}

// 	return user.groupIDs, nil // user.GroupIDs là slice string từ document DB
// }
