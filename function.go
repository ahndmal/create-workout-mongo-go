package gcp_mongo_sless

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
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

func init() {
	functions.HTTP("createWorkout", createWorkout)
}

func createWorkout(wr http.ResponseWriter, req *http.Request) {

	if req.Method == "GET" {
		fmt.Fprint(wr, "TEST")
	} else if req.Header.Get("Content-Type") != "" {
		var workout Workout
		if err := json.NewDecoder(req.Body).Decode(&workout); err != nil {
			fmt.Fprint(wr, err)
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
		doc := bson.D{
			{"record", time.Now().Unix()},
			{"sets", workout.Sets},
			{"workout_date", workout.WorkoutDate},
			{"creation_date", time.Now().String()},
			{"workout_type", workout.WorkoutType},
			{"comments", workout.Comments},
			{"day", workout.Day},
			{"month", workout.Month},               // todo - take from workout_date
			{"week", workout.Week},                 // todo-take from date
			{"month", time.Now().Month().String()}, // todo-take from date
			{"year", time.Now().Year()},            // take from date
		}
		res, err := coll.InsertOne(context.TODO(), doc)
		//bodyJson, err := ioutil.ReadAll(req.Body)
		//fmt.Fprintf(wr, string(bodyJson))
		wr.Header().Set("Content-Type", "application/json")
		if err != nil {
			return
		}
		fmt.Printf("Created workout with _id: %v\n", res.InsertedID)
	}
}
