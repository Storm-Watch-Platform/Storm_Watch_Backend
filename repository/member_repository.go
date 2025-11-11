package repository

import (
	"context"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type memberRepository struct {
	database   mongo.Database
	collection string
}

// Khởi tạo repository
func NewMemberRepository(db mongo.Database, collection string) domain.MemberRepository {
	return &memberRepository{
		database:   db,
		collection: collection,
	}
}

// Thêm member vào group
func (r *memberRepository) Add(c context.Context, m *domain.Member) error {
	col := r.database.Collection(r.collection)
	_, err := col.InsertOne(c, m)
	return err
}

// Cập nhật vị trí của member (lat, lon, updatedAt)
func (r *memberRepository) UpdateLocation(
	c context.Context,
	userID primitive.ObjectID,
	lat, lon float64,
	updatedAt time.Time,
) error {
	col := r.database.Collection(r.collection)
	filter := bson.M{"userId": userID}
	update := bson.M{
		"$set": bson.M{
			"lat":       lat,
			"lon":       lon,
			"updatedAt": updatedAt,
		},
	}
	_, err := col.UpdateOne(c, filter, update)
	return err
}

// Lấy danh sách member theo group
func (r *memberRepository) ListByGroup(c context.Context, groupID primitive.ObjectID) ([]domain.Member, error) {
	col := r.database.Collection(r.collection)

	cursor, err := col.Find(c, bson.M{"groupId": groupID})
	if err != nil {
		return nil, err
	}

	var members []domain.Member
	err = cursor.All(c, &members)
	if members == nil {
		members = []domain.Member{}
	}

	return members, err
}
