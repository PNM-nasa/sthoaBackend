package forum

import "go.mongodb.org/mongo-driver/bson/primitive"

type Comment struct {
	ID      primitive.ObjectID `bson:"_id"`
	Body    string
	UserID  int
	Comment []Comment
}
