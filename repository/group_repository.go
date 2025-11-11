package repository

import (
	"context"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type groupRepository struct {
	database   mongo.Database
	collection string
}

func NewGroupRepository(db mongo.Database, collection string) domain.GroupRepository {
	return &groupRepository{
		database:   db,
		collection: collection,
	}
}

// Tạo group mới
func (r *groupRepository) Create(ctx context.Context, g *domain.Group) error {
	col := r.database.Collection(r.collection)
	_, err := col.InsertOne(ctx, g)
	return err
}

// Xóa group theo ID
func (r *groupRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	col := r.database.Collection(r.collection)
	_, err := col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Lấy group bằng mã mời
func (r *groupRepository) GetByInviteCode(ctx context.Context, code string) (domain.Group, error) {
	col := r.database.Collection(r.collection)
	var group domain.Group
	err := col.FindOne(ctx, bson.M{"inviteCode": code}).Decode(&group)
	return group, err
}

// Lấy group bằng ID
func (r *groupRepository) GetByID(ctx context.Context, id primitive.ObjectID) (domain.Group, error) {
	col := r.database.Collection(r.collection)
	var group domain.Group
	err := col.FindOne(ctx, bson.M{"_id": id}).Decode(&group)
	return group, err
}

func (r *groupRepository) AddMember(ctx context.Context, groupID primitive.ObjectID, userID primitive.ObjectID) error {
	collection := r.database.Collection(r.collection)
	filter := bson.M{"_id": groupID}
	update := bson.M{"$addToSet": bson.M{"memberIDs": userID}}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}
