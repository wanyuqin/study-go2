package po

import "go.mongodb.org/mongo-driver/bson/primitive"

type Common struct {
	ID       primitive.ObjectID `json:"id" copier:"Id"`
	CreateAt string             `json:"create_at"`
}
