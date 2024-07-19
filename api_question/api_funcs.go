package apiquestion

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	vars "github.com/PNM-nasa/sthoabackend/vars"
)

// please don't edit it
const (
	typeChosseOne   = "chosse-one"
	typeTrueFalse   = "true-false"
	typeShortAnswer = "short-answer"
)

func CreateQuestion(c fiber.Ctx) error {

	// Verify the admin key from query parameters
	key := c.Query("key")
	if key != vars.ADMIN_KEY {
		return c.Status(fiber.StatusUnauthorized).SendString("error: key not right")
	}

	var dataQuestion map[string]interface{}
	if err := json.Unmarshal(c.Body(), &dataQuestion); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("errBodyToData")
	}

	// Extract and validate fields from the request body
	typeQuestion, ok := dataQuestion["type_question"].(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("error: invalid type_question")
	}
	if typeQuestion != typeChosseOne &&
		typeQuestion != typeTrueFalse &&
		typeQuestion != typeShortAnswer {
		return c.Status(fiber.StatusBadRequest).SendString("error: invalid type_question, type_question must is " + typeChosseOne + " " + typeTrueFalse + " " + typeShortAnswer)
	}

	title, ok := dataQuestion["title"].(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("error: invalid title")
	}

	optionsInterface, ok := dataQuestion["options"].([]interface{})
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("error: invalid options")
	}

	var options []string
	for _, element := range optionsInterface {
		option, ok := element.(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).SendString("error: invalid option element")
		}
		options = append(options, option)
	}

	answer, ok := dataQuestion["answer"].(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("error: invalid answer")
	}
	switch typeQuestion {
	case typeChosseOne:
		if answer != "0" && answer != "1" && answer != "2" && answer != "3" {
			return c.Status(fiber.StatusBadRequest).SendString("error: invalid answer; when type_question is \"chosse-one\", answer must is number(0-3) string")
		}
		break
	case typeTrueFalse: // continue
	case typeShortAnswer: // continue
	}

	// bug : must check "dataQuestion["lesson_id"].(string)" not error
	lessonID, err := primitive.ObjectIDFromHex(dataQuestion["lesson_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("error: invalid lesson_id")
	}

	question := Question{
		LessonID:     lessonID,
		TypeQuestion: typeQuestion,
		Title:        title,
		Options:      options,
		Answer:       answer,
	}

	result, errInsert := Collection.InsertOne(context.TODO(), question)
	if errInsert != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error: unable to insert question")
	}

	log.Println("create question : ", result.InsertedID)

	return c.SendStatus(fiber.StatusOK)
}

const tRandom = "random"

func GetQuestion(c fiber.Ctx) error {
	task := c.Query("task")
	switch task {
	case tRandom:
		return getQuestionRandom(c)
	default:
		return c.SendStatus(fiber.StatusBadRequest)
	}
}

// random get some question
func getQuestionRandom(c fiber.Ctx) error {
	size, err := strconv.Atoi(c.Query("size"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("error invalid Query \"size\"")
	}

	cursor, _ := Collection.Aggregate(
		context.TODO(),
		[]bson.D{
			{{"$sample", bson.D{{"size", size}}}},
		},
	)
	defer cursor.Close(context.TODO())

	var output []Question
	for cursor.Next(context.TODO()) {
		var result Question
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		output = append(output, result)
		log.Println(result) // Print the random document
	}

	ouputByte, _ := json.Marshal(output)
	return c.Status(fiber.StatusOK).Send(ouputByte)
}
