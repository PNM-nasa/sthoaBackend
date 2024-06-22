package main

// run `go mod tidy` after editing import lib
import (
	"context"
	"encoding/json"
	"fmt"

	"log"

	"strconv"

	// "io"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"

	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func getlessson(c fiber.Ctx) error {
	c.SendString(c.Params("id"))
	return nil
}

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

// run `go run .`
func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("warn : Error loading .env file, don't care in render")
	}

	checkKeyEnv([]string{"MONGODB_URI"})

	// fmt.Println(string(getFileDrive(0, "1wU5nCulZZfmN133siIBSKJZ1cU8SFw8F")))
	MONGODB_URI := os.Getenv("MONGODB_URI")
	ADMIN_KEY := os.Getenv("ADMIN_KEY")

	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(MONGODB_URI).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {

		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	coll := client.Database("sample_mflix").Collection("movies")
	title := "Back to the Future"
	var result bson.M
	err = coll.FindOne(context.TODO(), bson.D{{Key: "title", Value: title}}).
		Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the title %s\n", title)
		return
	}
	if err != nil {
		panic(err)
	}
	df, err := json.MarshalIndent(result, "", "	")
	var djson map[string]interface{}
	json.Unmarshal(df, &djson)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%s\n", df)
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	fmt.Println(djson["year"])

	lesioncoll := client.Database("core_dp").Collection("lessons")
	// lesioncoll.Indexes().CreateOne(
	// 	context.TODO(),

	// )
	//bufffile, err := os.ReadFile("sample-pdf-file.pdf")
	lesson := Lesson{LessonID: 20, Name: "lesson 2", DriveID: "1wU5nCulZZfmN133siIBSKJZ1cU8SFw8F"}
	// lesioncoll.InsertOne(
	// 	context.TODO(),
	// 	bson.D{{"key", lesson.Name}},
	// )

	b, err := lesioncoll.InsertOne(context.TODO(), lesson)
	if err != nil {
		panic(err)

	}
	print(b)

	// newRestaurant := Restaurant{Name: "8282", Cuisine: "Korean"}
	// lesioncoll.InsertOne(context.TODO(), newRestaurant)

	app := fiber.New()

	appv1 := app.Route("/v1")

	appv1.Route("/lesson/:id").
		Get(func(c fiber.Ctx) error {
			id, err := strconv.Atoi(c.Params("id"))
			if err != nil {
				return c.SendString("error: id must is number")
			}
			var lesson Lesson
			err = lesioncoll.FindOne(context.TODO(), bson.D{
				{Key: "lessonid", Value: id}},
			).Decode(&lesson)
			if err != nil {
				return c.SendString("error: not found pdf file with id: " + strconv.Itoa(id))
			}
			log.Println(lesson.DriveID)
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
	log.Fatal(app.Listen(":3000"))
}
