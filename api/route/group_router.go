package route

import (
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/api/controller"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/repository"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/usecase"
	"github.com/gin-gonic/gin"
)

func NewGroupRouter(env *bootstrap.Env, timeout time.Duration, db mongo.Database, group *gin.RouterGroup) {
	gr := repository.NewGroupRepository(db, domain.CollectionGroup) // tạo user repository để làm việc với collection user
	mr := repository.NewMemberRepository(db, domain.CollectionMember)
	ur := repository.NewUserRepository(db, domain.CollectionUser)
	lr := repository.NewLocationRepo(db, domain.CollectionLocation)

	gc := &controller.GroupController{ // tạo controller
		GroupUsecase: usecase.NewGroupUsecase(gr, mr, ur, lr, timeout),
	}

	groupRoutes := group.Group("/groups")
	{
		// tạo group mới
		groupRoutes.POST("/create", gc.CreateGroup)

		// get invite code từ group id
		groupRoutes.GET("/:groupId/invite", gc.GetInviteCodeByGroupID)

		// lấy thông tin group bằng invite code
		groupRoutes.GET("/invite/:code", gc.GetByInviteCode)

		// join group + add thông tin group vào user
		// tham gia group bằng code
		groupRoutes.PUT("/join/:code", gc.JoinGroup)

		// xóa group chưa xóa người (index groupid & userid trong member) -> nên chuyển thành xóa member khỏi group
		groupRoutes.DELETE("/:groupId/members", gc.Delete)

		groupRoutes.GET("/:groupId", gc.GetGroupByID)

		groupRoutes.GET("/:groupId/members/:memberId", gc.GetMemberInGroup)
	}
}
