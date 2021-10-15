package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"soap/models"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Users struct {
	Users []models.User `json:"users"`
}

func main() {
	progArgs := os.Args
	loadData := false
	if len(progArgs) > 1 {
		if strings.ToLower(progArgs[1]) == "true" ||
			strings.ToLower(progArgs[1]) == "process" {
			loadData = true
		}
	}
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprintf("mongodb://%s/", os.Getenv("mongodb"))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	database := client.Database("soap")
	if loadData {
		jsonfile, err := os.Open("initialUsers.json")
		if err != nil {
			log.Fatal(err)
		}
		defer jsonfile.Close()

		byteValue, _ := ioutil.ReadAll(jsonfile)

		var users Users

		json.Unmarshal(byteValue, &users)

		userCollection := database.Collection("users")

		for _, user := range users.Users {
			user.ID = primitive.NewObjectID()
			user.Creds.SetPassword("InitialPassword")
			insertResult, err := userCollection.InsertOne(ctx, user)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(insertResult.InsertedID)
		}
	}
}
