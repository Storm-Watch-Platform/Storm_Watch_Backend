package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CollectionGroup  = "groups"
	CollectionMember = "members"
)

// Group: mỗi gia đình / nhóm cứu hộ
type Group struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name       string               `bson:"name" json:"name"`
	InviteCode string               `bson:"inviteCode" json:"inviteCode"`
	MemberIDs  []primitive.ObjectID `bson:"memberIDs" json:"memberIDs"`
	ExpiresAt  time.Time            `bson:"expiresAt" json:"expiresAt"`
	CreatedAt  time.Time            `bson:"createdAt" json:"createdAt"`
}

type GroupMemberDetail struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`

	Location MemberLocation `json:"location"`
}

type MemberLocation struct {
	AccuracyM   float64  `json:"accuracy_m"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	UpdatedAt   int64    `json:"updated_at"`
	Coordinates GeoPoint `json:"location"`
}

// Member: lưu thông tin từng người trong nhóm + vị trí realtime
type Member struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID   primitive.ObjectID `bson:"groupId" json:"groupId"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Lat       float64            `bson:"lat" json:"lat"`
	Lon       float64            `bson:"lon" json:"lon"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

//
// INTERFACES (Repository layer)
//

// Repository interface
type GroupRepository interface {
	Create(ctx context.Context, g *Group) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetByInviteCode(ctx context.Context, code string) (Group, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (Group, error)
	AddMember(ctx context.Context, groupID primitive.ObjectID, userID primitive.ObjectID) error
}

type MemberRepository interface {
	Add(c context.Context, m *Member) error
	UpdateLocation(c context.Context, userID primitive.ObjectID, lat, lon float64, updatedAt time.Time) error
	ListByGroup(c context.Context, groupID primitive.ObjectID) ([]Member, error)
}

type GroupUsecase interface {
	CreateGroup(c context.Context, name string) (*Group, error)
	GetByInviteCode(c context.Context, code string) (*Group, error)
	DeleteGroup(c context.Context, userID string, groupID primitive.ObjectID) error
	GetInviteCodeByGroupID(c context.Context, groupID primitive.ObjectID) (string, error)
	JoinGroup(ctx context.Context, userID string, inviteCode string) error
	GetMemberInGroup(ctx context.Context, groupID, memberID string) (*GroupMemberDetail, error)
}
