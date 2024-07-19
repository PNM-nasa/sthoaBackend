package forum

import "go.mongodb.org/mongo-driver/bson/primitive"

type Post struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Tile     string             `json:"title"`
	Body     string
	UserID   primitive.ObjectID
	Comments []Comment
}
