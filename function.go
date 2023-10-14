package gcp_mongo_sless

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

type Workout struct {
	Id           string    `json:"_id" bson:"_id,omitempty"`
	Record       int64     `json:"record"`
	Sets         int       `json:"sets"`
	Comments     string    `json:"comments"`
	CreationDate time.Time `json:"creation_date" bson:"creation_date"`
	WorkoutDate  string    `json:"workout_date" bson:"workout_date"`
	Day          string    `json:"day"`
	Week         int       `json:"week"`
	WorkoutType  string    `json:"workout_type" bson:"workout_type"`
	Month        string    `json:"month"`
	Year         int       `json:"year"`
}

type WorkoutDto struct {
	Sets        string `json:"sets"`
	Comments    string `json:"comments"`
	WorkoutDate string `json:"workout_date" bson:"workout_date"`
	WorkoutType string `json:"workout_type" bson:"workout_type"`
}

func init() {
	functions.HTTP("createWorkout", createWorkout)
}

func createWorkout(wr http.ResponseWriter, req *http.Request) {

	header := wr.Header()
	//wr.Header().Add("Access-Control-Allow-Origin", "https://workouts-web-static.vercel.app")
	header.Add("Access-Control-Allow-Origin", "*")
	header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	header.Add("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	header.Add("Access-Control-Max-Age", "600")
	header.Add("Accept", "application/json")

	// methods
	// GET
	if req.Method == "GET" {
		_, err := fmt.Fprint(wr, "GET not supported")
		if err != nil {
			log.Printf("> Error when returing GET response: %v", err)
		}

		// PUT
	} else if req.Method == "PUT" {

		var workout Workout
		if err := json.NewDecoder(req.Body).Decode(&workout); err != nil {
			log.Printf(">>> Error decoding JSON: %v", err)
			log.Panicln(err)
		}

		uri := os.Getenv("MONGODB_URI")
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := client.Disconnect(context.TODO()); err != nil {
				panic(err)
			}
		}()
		coll := client.Database("workouts").Collection("workouts")
		wDate := req.URL.Query().Get("workout_date")
		wType := req.URL.Query().Get("workout_type")
		// filter
		var existingWrkt Workout
		filter := bson.D{{"workout_date", wDate}, {"workout_type", wType}}
		coll.FindOne(context.TODO(), filter).Decode(&existingWrkt)
		// data
		var sets, week int
		var comments, workDate, workType, day, month string
		if workout.WorkoutDate != "" {
			workDate = workout.WorkoutDate
		} else {
			workDate = existingWrkt.WorkoutDate
		}
		if workout.Sets != 0 {
			sets = workout.Sets
		} else {
			sets = existingWrkt.Sets
		}
		if workout.Comments != "" {
			comments = workout.Comments
		} else {
			comments = existingWrkt.Comments
		}
		if workout.WorkoutType != "" {
			workType = workout.WorkoutType
		} else {
			workType = existingWrkt.WorkoutType
		}
		if workout.Day != "" {
			day = workout.Day
		} else {
			day = existingWrkt.Day
		}
		if workout.Month != "" {
			month = workout.Month
		} else {
			month = existingWrkt.Month
		}
		if workout.Week != 0 {
			week = workout.Week
		} else {
			week = existingWrkt.Week
		}
		updatedWorkut := bson.D{{"$set", bson.D{
			{"workout_date", workDate},
			{"sets", sets}, {"comments", comments}, {"workout_type", workType},
			{"day", day}, {"month", month}, {"week", week}}}}

		uddateRes, err2 := coll.UpdateOne(context.TODO(), filter, updatedWorkut)
		if err2 != nil {
			log.Panicln(err2)
		}

		fmt.Fprintf(wr, "Updated workout %s", uddateRes.UpsertedID)

		// POST
	} else if req.Method == "POST" {
		var workout WorkoutDto
		if err := json.NewDecoder(req.Body).Decode(&workout); err != nil {
			log.Printf(">>> Error decoding JSON: %v", err)
			log.Panicln(err)
		}

		uri := os.Getenv("MONGODB_URI")
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := client.Disconnect(context.TODO()); err != nil {
				panic(err)
			}
		}()
		coll := client.Database("workouts").Collection("workouts")
		wDateDate, err := time.Parse(time.DateOnly, workout.WorkoutDate)
		wyear, wWeek := wDateDate.ISOWeek()

		sets, err := strconv.Atoi(workout.Sets)

		doc := bson.D{
			{"record", time.Now().Unix()},
			{"sets", sets},
			{"workout_date", workout.WorkoutDate},
			//{"creation_date", time.Now().Format(time.RFC1123)},
			{"creation_date", time.Now()},
			{"workout_type", workout.WorkoutType},
			{"comments", workout.Comments},
			{"day", wDateDate.Weekday().String()},
			{"week", wWeek},
			{"month", time.Now().Month().String()},
			{"year", wyear},
		}

		//time.Parse(time.RFC822, fmt.Sprintf("01 %s %s 00:00 MST", month[0:3], year[2:4])) //RFC822 = "02 Jan 06 15:04 MST"

		res, err := coll.InsertOne(context.TODO(), doc)

		// todo: delete debug
		bodyJson, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("> Error when reading Insert reply data. %v", err)
		}
		fmt.Println(string(bodyJson))

		//wr.Header().Set("Content-Type", "application/json")

		//_, err2 := fmt.Fprintf(wr, "Created workout with _id: %v\n", res.InsertedID)
		_, err2 := fmt.Fprintf(wr, "{ \"created_id\": \"%s\" }", res.InsertedID)

		if err2 != nil {
			log.Printf(">> Error when writing parsed JSON: %v", err2)
		}
	}
}
