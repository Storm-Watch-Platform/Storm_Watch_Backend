package controller

import (
	"context"
	"net/http"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupController struct {
	GroupUsecase domain.GroupUsecase
}

// 📍 POST /group/create
func (gc *GroupController) CreateGroup(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	// Lấy user ID từ middleware JWT
	// userID := c.GetString("x-user-id")

	group, err := gc.GroupUsecase.CreateGroup(c, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// 📍 GET /group/invite/:code
func (gc *GroupController) GetByInviteCode(c *gin.Context) {
	code := c.Param("code")

	group, err := gc.GroupUsecase.GetByInviteCode(c, code)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "Group not found"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// 📍 DELETE /group/:id — out nhóm
func (gc *GroupController) Delete(c *gin.Context) {
	idStr := c.Param("groupId")
	groupID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid group ID"})
		return
	}

	userID := c.GetString("x-user-id")

	err = gc.GroupUsecase.DeleteGroup(c, userID, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "Group deleted successfully",
	})
}

// 📍 GET /groups/:groupId/invite — lấy mã invite của group
func (gc *GroupController) GetInviteCodeByGroupID(c *gin.Context) {
	idStr := c.Param("groupId")

	groupID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid group ID"})
		return
	}

	code, err := gc.GroupUsecase.GetInviteCodeByGroupID(c, groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "Group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inviteCode": code,
	})
}

// 📍 PUT /groups/join/:inviteCode
func (gc *GroupController) JoinGroup(c *gin.Context) {
	inviteCode := c.Param("code")
	userID := c.GetString("x-user-id")

	if err := gc.GroupUsecase.JoinGroup(c, userID, inviteCode); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "Joined group successfully"})
}

// 📍 GET /groups/:groupId — lấy thông tin group theo ID
func (gc *GroupController) GetGroupByID(c *gin.Context) {
	idStr := c.Param("groupId")

	group, err := gc.GroupUsecase.(interface {
		GetGroupByIDString(c context.Context, idStr string) (*domain.Group, error)
	}).GetGroupByIDString(c, idStr) // gọi hàm usecase vừa tạo
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "Group not found"})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (gc *GroupController) GetMemberInGroup(c *gin.Context) {
	groupID := c.Param("groupId")
	memberID := c.Param("memberId")

	member, err := gc.GroupUsecase.GetMemberInGroup(c, groupID, memberID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, member)
}
