package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CollectionUser = "users"
)

type User struct {
	ID       primitive.ObjectID   `bson:"_id"`
	Name     string               `bson:"name"`
	Phone    string               `bson:"phone"`
	Password string               `bson:"password"`
	Role     string               `bson:"role"`
	GroupIDs []primitive.ObjectID `bson:"groupIDs"`
}

type UserRepository interface {
	Create(c context.Context, user *User) error
	Fetch(c context.Context) ([]User, error)
	GetByPhone(c context.Context, phone string) (User, error)
	GetByID(c context.Context, id string) (User, error)
	AddGroup(ctx context.Context, userID primitive.ObjectID, groupID primitive.ObjectID) error
}
