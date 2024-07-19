package apiquestion

import "go.mongodb.org/mongo-driver/bson/primitive"

// Question represents the structure of a question document
type Question struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	LessonID     primitive.ObjectID
	TypeQuestion string `bson:"type_question"`
	Title        string
	Options      []string
	Answer       string
}
