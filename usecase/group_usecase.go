package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type groupUsecase struct {
	groupRepository    domain.GroupRepository
	memberRepository   domain.MemberRepository
	userRepository     domain.UserRepository
	locationRepository domain.LocationRepository
	contextTimeout     time.Duration
}

func NewGroupUsecase(
	groupRepo domain.GroupRepository,
	memberRepo domain.MemberRepository,
	userRepo domain.UserRepository,
	locationRepo domain.LocationRepository,
	timeout time.Duration,
) domain.GroupUsecase {
	return &groupUsecase{
		groupRepository:    groupRepo,
		memberRepository:   memberRepo,
		userRepository:     userRepo,
		locationRepository: locationRepo,
		contextTimeout:     timeout,
	}
}

// 📍 Tạo group mới
func (gu *groupUsecase) CreateGroup(c context.Context, name string) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, gu.contextTimeout)
	defer cancel()

	// userObjID, err := primitive.ObjectIDFromHex(userID)
	// if err != nil {
	// 	return nil, errors.New("invalid user id")
	// }

	group := &domain.Group{
		ID:         primitive.NewObjectID(),
		Name:       name,
		InviteCode: primitive.NewObjectID().Hex()[18:], // lấy 6 ký tự cuối làm mã
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		MemberIDs:  []primitive.ObjectID{},
	}

	// 1️⃣ Tạo group
	if err := gu.groupRepository.Create(ctx, group); err != nil {
		return nil, err
	}

	// 2️⃣ Tự động thêm user làm thành viên đầu tiên (owner)
	// member := &domain.Member{
	// 	ID:        primitive.NewObjectID(),
	// 	GroupID:   group.ID,
	// 	UserID:    userObjID,
	// 	Lat:       0,
	// 	Lon:       0,
	// 	UpdatedAt: time.Now(),
	// }
	// _ = gu.memberRepository.Add(ctx, member)

	return group, nil
}

// 📍 Lấy group bằng mã invite
func (gu *groupUsecase) GetByInviteCode(c context.Context, code string) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, gu.contextTimeout)
	defer cancel()

	group, err := gu.groupRepository.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// 📍 Xóa group
func (gu *groupUsecase) DeleteGroup(c context.Context, userID string, groupID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(c, gu.contextTimeout)
	defer cancel()

	// kiểm tra user có tồn tại không
	_, err := gu.userRepository.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	return gu.groupRepository.Delete(ctx, groupID)
}

func (gu *groupUsecase) GetInviteCodeByGroupID(c context.Context, groupID primitive.ObjectID) (string, error) {
	ctx, cancel := context.WithTimeout(c, gu.contextTimeout)
	defer cancel()

	group, err := gu.groupRepository.GetByID(ctx, groupID)
	if err != nil {
		return "", err
	}
	return group.InviteCode, nil
}

func (u *groupUsecase) JoinGroup(ctx context.Context, userID string, inviteCode string) error {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 1️⃣ Tìm group theo invite code
	group, err := u.groupRepository.GetByInviteCode(c, inviteCode)
	if err != nil {
		return fmt.Errorf("invalid invite code")
	}

	// 2️⃣ Convert userID → ObjectID
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}

	// 3️⃣ Kiểm tra xem user đã ở trong group chưa
	existingMembers, err := u.memberRepository.ListByGroup(c, group.ID)
	if err != nil {
		return fmt.Errorf("failed to check members: %v", err)
	}

	for _, m := range existingMembers {
		if m.UserID == uid {
			return fmt.Errorf("user already in group")
		}
	}

	// 3️⃣ Thêm member mới vào collection members
	member := &domain.Member{
		ID:        primitive.NewObjectID(),
		GroupID:   group.ID,
		UserID:    uid,
		Lat:       0,
		Lon:       0,
		UpdatedAt: time.Now(),
	}

	// ➕ Thêm record member mới
	if err := u.memberRepository.Add(c, member); err != nil {
		return fmt.Errorf("failed to add member: %v", err)
	}

	// 4️⃣ Thêm userID vào group.memberIDs
	if err := u.groupRepository.AddMember(c, group.ID, uid); err != nil {
		return fmt.Errorf("failed to update group: %v", err)
	}

	// 5️⃣ Thêm groupID vào user.groupIds
	if err := u.userRepository.AddGroup(c, uid, group.ID); err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}

// 📍 Lấy thông tin Group từ string ID
func (u *groupUsecase) GetGroupByIDString(ctx context.Context, idStr string) (*domain.Group, error) {
	c, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Convert string -> ObjectID
	groupID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, errors.New("invalid group id")
	}

	// Lấy group từ repository
	group, err := u.groupRepository.GetByID(c, groupID)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (gu *groupUsecase) GetMemberInGroup(ctx context.Context, groupID, memberID string) (*domain.GroupMemberDetail, error) {

	// Convert IDs
	gID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, err
	}
	uID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		return nil, err
	}
	// Step 1: Load group
	group, err := gu.groupRepository.GetByID(ctx, gID)
	if err != nil {
		return nil, err
	}

	// Step 2: Check membership
	isMember := false
	for _, id := range group.MemberIDs {
		if id == uID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, fmt.Errorf("user not in this group")
	}

	// Step 3: Load user info
	user, err := gu.userRepository.GetByID(ctx, uID.Hex())
	if err != nil {
		return nil, err
	}

	// Step 4: Load latest location for this user
	loc, err := gu.locationRepository.GetByUserID(ctx, uID.Hex())
	if err != nil {
		// location optional -> return user only
		loc = nil
	}

	// Step 5: Combine response
	resp := &domain.GroupMemberDetail{
		ID:    user.ID.Hex(),
		Name:  user.Name,
		Phone: user.Phone,
	}

	if loc != nil {
		resp.Location = domain.MemberLocation{
			AccuracyM:   loc.AccuracyM,
			Status:      loc.Status,
			UpdatedAt:   loc.UpdatedAt,
			Coordinates: loc.Location,
		}
	}

	return resp, nil
}
