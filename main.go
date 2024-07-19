package main

// run `go mod tidy` after editing import lib
import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	"log"

	"strconv"

	// "io"
	"os"
	"time"

	apiquestion "github.com/PNM-nasa/sthoabackend/api_question"
	"github.com/PNM-nasa/sthoabackend/forum"
	"github.com/PNM-nasa/sthoabackend/random"
	vars "github.com/PNM-nasa/sthoabackend/vars"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"

	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func connect(uri string) (*mongo.Client, context.Context,
	context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)

	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

func close(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {

	// CancelFunc to cancel to context
	defer cancel()

	// client provides a method to close
	// a mongoDB connection.
	defer func() {

		// client.Disconnect method also has deadline.
		// returns error if any,
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func ping(client *mongo.Client, ctx context.Context) error {

	// mongo.Client has Ping to ping mongoDB, deadline of
	// the Ping method will be determined by cxt
	// Ping method return error if any occurred, then
	// the error can be handled.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	fmt.Println("connected successfully")
	return nil
}

type Lesson struct {
	LessonID int64
	Name     string
	DriveID  string
	Data     []byte `bson:"data,omitempty"`
}

type Restaurant struct {
	Name         string
	RestaurantId string        `bson:"restaurant_id,omitempty"`
	Cuisine      string        `bson:"cuisine,omitempty"`
	Address      interface{}   `bson:"address,omitempty"`
	Borough      string        `bson:"borough,omitempty"`
	Grades       []interface{} `bson:"grades,omitempty"`
}

type Question struct {
	LessonID int64
	Form     string
	Content  string
	Options  []string
	Answer   string
}

type FormQuestion struct {
	truefalse string
	chossOne  string
}

// run `go run .`
func main() {

	// reset "lessons" forder
	os.RemoveAll("lessons")
	os.Mkdir("lessons", 0755)

	// formQuestion := &FormQuestion{
	// 	truefalse: "truefalse",
	// 	chossOne:  "chossOne",
	// }

	err := godotenv.Load()
	if err != nil {
		log.Println("warn : Error loading .env file, don't care in render")
	}
	checkKeyEnv([]string{"MONGODB_URI"})

	MONGODB_URI := os.Getenv("MONGODB_URI")
	ADMIN_KEY := os.Getenv("ADMIN_KEY")

	vars.ADMIN_KEY = ADMIN_KEY

	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(MONGODB_URI).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	lesioncoll := client.Database("core_dp").Collection("lessons")
	questionColl := client.Database("core_dp").Collection("questions")
	userColl := client.Database("core_dp").Collection("user")
	forumColl := client.Database("core_dp").Collection("forumColl")

	apiquestion.Setup(questionColl)

	// lesioncoll.Indexes().CreateOne(
	// 	context.TODO(),

	// )
	//bufffile, err := os.ReadFile("sample-pdf-file.pdf")
	// lesson := Lesson{LessonID: 20, Name: "lesson 2", DriveID: "1wU5nCulZZfmN133siIBSKJZ1cU8SFw8F"}

	// b, err := lesioncoll.InsertOne(context.TODO(), lesson)
	// if err != nil {
	// 	panic(err)

	// }
	// print(b)

	// newRestaurant := Restaurant{Name: "8282", Cuisine: "Korean"}
	// lesioncoll.InsertOne(context.TODO(), newRestaurant)
	// print("formQuestion.chossOne : ", formQuestion.chossOne)
	// question := Question{LessonID: 0, Form: formQuestion.chossOne, Content: "quesion 1 ", Options: []string{"sf", "sfd", "dfd", "dfd"}, Answer: "0"}
	// questionColl.InsertOne(
	// 	context.TODO(),
	// 	question,
	// )

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
		AllowOrigins:     "http://localhost:5000",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	appv1 := app.Route("/v1")

	appv2 := app.Route("/v2")

	// ?key=ADMIN_KEY
	// body:
	//   type_question string | typeChosseOne typeTrueFalse typeShortAnswer
	//   title         string
	//   options 	   []string
	//   answer		   string
	//     type_question is typeChosseOne   | number(0-3) string | ex : "0"
	//     type_question is typeTrueFalse   | [4]string and only container "t" or "f" | ex : "tftf"
	//     type_question is typeShortAnswer | string | ex : "this is answer"
	appv2.Route("/question").Post(apiquestion.CreateQuestion)
	/**
	*  ?task=random
	*  	 ?size int | lesson_id stringHEX | ...
	 */
	appv2.Route("/question").Get(apiquestion.GetQuestion)

	appv1.Route("/lesson/:id").
		Get(func(c fiber.Ctx) error {
			id, err := strconv.Atoi(c.Params("id"))
			if err != nil {
				return c.SendString("error: id must is number")
			}
			println(id)
			var lesson Lesson
			err = lesioncoll.FindOne(context.TODO(), bson.D{
				{Key: "lessonid", Value: id}},
			).Decode(&lesson)
			if err != nil {
				return c.SendString("error: not found pdf file with id: " + strconv.Itoa(id))
			}
			log.Println(lesson.DriveID, lesson.Name)
			data, err := getFileDrive(id, lesson.DriveID)
			if err != nil {
				panic(err)
			}
			return c.Send(data)
		}).
		Route("/:name/:driveid").
		Post(func(c fiber.Ctx) error {
			key := c.Query("admin-key", "0")
			if key != ADMIN_KEY {
				return c.SendString("error key")
			}
			idfile, err := strconv.Atoi(c.Params("id"))
			if err != nil {
				return c.SendString("error: id must is number")
			}
			namefile := c.Params("name")
			driveID := c.Params("driveid")
			data, err := getFileDrive(idfile, driveID)

			if err != nil {
				return c.SendString(`error : id file drive `)
			}

			if false {
				log.Fatalf(string(data))
				return nil
			}

			lesson := Lesson{LessonID: int64(idfile), Name: namefile, DriveID: driveID}
			lesioncoll.InsertOne(
				context.TODO(),
				lesson,
			)
			return nil
		})
	appv1.Route("question").Get(func(c fiber.Ctx) error {
		return c.SendStatus(200)
	})
	appv1.Route("question/:lessonid/:amount").Get(func(c fiber.Ctx) error {
		lessonID, err := strconv.Atoi(c.Params("lessonid"))
		if err != nil {
			return c.SendString("error: id must is number")
		}
		amount, err := strconv.Atoi(c.Params("amount"))
		if err != nil {
			return c.SendString("error: amount must is number")
		}

		cursor, err := questionColl.Find(
			context.TODO(),
			bson.D{
				{"lessonid", lessonID},
			},
		)
		if err != nil {
			return nil
		}
		var questions []Question
		// cursor.Decode(&questions)
		for cursor.Next(context.TODO()) {
			//Create a value into which the single document can be decoded
			var elem Question
			err := cursor.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}

			questions = append(questions, elem)
			p, _ := json.MarshalIndent(questions, "", "	")
			print(string(p))
		}

		var output []Question
		for index, item := range questions {
			if amount <= 0 {
				break
			}

			if float64(amount)/float64(len(questions)-index) >= rand.Float64() {

				output = append(output, item)
				amount--
				fmt.Println(strconv.Itoa(index))
			}

		}
		// outputjson, err := json.Marshal(output)
		outputjson, err := json.MarshalIndent(output, "", "	")
		if err != nil {
			panic(err)
		}

		return c.Send(outputjson)
	})
	appv1.Route("user").Get(func(c fiber.Ctx) error {
		accessToken := c.Query("access_token")
		var user User
		response, errhttp := http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + accessToken)
		if errhttp != nil {
			panic(errhttp)
		}
		body, _ := ioutil.ReadAll(response.Body)
		print(string(body))
		var usergg Usergg
		err = json.Unmarshal(body, &usergg)
		if err != nil {
			panic(err)
		}

		amountUser, errAmountUser := userColl.CountDocuments(
			context.TODO(),
			bson.D{
				{"email", usergg.Email},
			},
		)
		if errAmountUser != nil {
			panic(errAmountUser)
		}

		if amountUser == 0 {
			print("create user")
			user.PhotoUrl = usergg.Picture
			user.Name = usergg.Email
			user.Lever = 0
			user.Email = usergg.Email
			user.Token = random.Createtoken16()
			userColl.InsertOne(
				context.TODO(),
				user,
			)
		} else {
			print("loading data user")
			err = userColl.FindOne(
				context.TODO(),
				bson.D{
					{"email", usergg.Email},
				},
			).Decode(&user)
		}

		data := map[string]string{
			"photoUrl": user.PhotoUrl,
			"name":     user.Name,
			"lever":    strconv.Itoa(user.Lever),
			"email":    user.Email,
			"token":    user.Token,
		}
		jsonString, _ := json.Marshal(data)
		c.SendString(string(jsonString))
		return nil
	})
	appv1.Route("user/callback").Get(func(c fiber.Ctx) error {
		print("brrr")
		token := c.Query("token")
		var user User
		amountUser, errAmountUser := userColl.CountDocuments(
			context.TODO(),
			bson.D{
				{"token", token},
			},
		)
		if errAmountUser != nil {
			panic(errAmountUser)
		}

		if amountUser == 0 {
			return c.SendStatus(400)
		} else {
			print("loading data user")
			err = userColl.FindOne(
				context.TODO(),
				bson.D{
					{"token", token},
				},
			).Decode(&user)
		}

		data := map[string]string{
			"photoUrl": user.PhotoUrl,
			"name":     user.Name,
			"lever":    strconv.Itoa(user.Lever),
			"email":    user.Email,
			"token":    user.Token,
		}
		jsonString, _ := json.Marshal(data)
		c.SendString(string(jsonString))
		return nil
	})

	appv1.Route("forum/createpost").Post(func(c fiber.Ctx) error {
		var data map[string]string
		println(string(c.Body()))
		json.Unmarshal(c.Body(), &data)

		println(data["a"])
		var post forum.Post
		post.Tile = data["title"]
		post.Body = data["body"]

		token := data["token"]
		var user User
		errorGetUser := userColl.FindOne(
			context.TODO(),
			bson.D{
				{"token", token},
			},
		).Decode(&user)
		if errorGetUser == mongo.ErrNoDocuments {
			return c.SendStatus(400)
		}
		if errorGetUser != nil {
			panic(errorGetUser)
		}
		post.UserID = user.ID
		forumColl.InsertOne(
			context.TODO(),
			post,
		)

		return c.SendStatus(200)
	})
	appv1.Route("forum/view").Get(func(c fiber.Ctx) error {
		//forumColl.
		return nil
	})
	log.Fatal(app.Listen(":4000"))
}
