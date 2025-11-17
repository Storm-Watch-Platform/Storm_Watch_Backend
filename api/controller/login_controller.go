package controller

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/gin-gonic/gin"
)

type LoginController struct {
	LoginUsecase domain.LoginUsecase
	Env          *bootstrap.Env
}

func (lc *LoginController) Login(c *gin.Context) {
	var request domain.LoginRequest

	err := c.ShouldBind(&request) // check định dạng json
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	//
	user, err := lc.LoginUsecase.GetUserByPhone(c, request.Phone)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "User not found with the given phone"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)) != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Invalid credentials"})
		return
	}

	accessToken, err := lc.LoginUsecase.CreateAccessToken(&user, lc.Env.AccessTokenSecret, lc.Env.AccessTokenExpiryHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	refreshToken, err := lc.LoginUsecase.CreateRefreshToken(&user, lc.Env.RefreshTokenSecret, lc.Env.RefreshTokenExpiryHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	loginResponse := domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// --- Bước mới: lấy group của user từ DB và cập nhật WSManager ---
	// groups, err := lc.LoginUsecase.GetGroupsOfUser(user.ID)
	// user.GroupIDs là []primitive.ObjectID
	// groupIDs := make([]string, len(user.GroupIDs))
	// for i, id := range user.GroupIDs {
	// 	groupIDs[i] = id.Hex() // chuyển sang string để cache
	// }
	// // Cập nhật WSManager
	// lc.WSManager.SetUserGroups(user.ID.Hex(), groupIDs)

	// // --- DEBUG: in cache ra console ---
	// cached := lc.WSManager.GetUserGroups(user.ID.Hex())
	// fmt.Printf("DEBUG: userID = %s, cached groups = %+v\n", user.ID.Hex(), cached)

	c.JSON(http.StatusOK, loginResponse)
}
