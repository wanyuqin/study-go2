package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type Common struct {
	ID       primitive.ObjectID `json:"id"`
	CreateAt string             `json:"create_at"`
}
