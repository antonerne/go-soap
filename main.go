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
	Users []models.User      `json:"users"`
	Books []models.BibleBook `json:"biblebooks"`
}

type Study struct {
	Studies []models.BibleStudy `json:"studies"`
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

		err := database.Collection("users").Drop(ctx)
		if err != nil {
			log.Println(err)
		}
		err = database.Collection("biblebooks").Drop(ctx)
		if err != nil {
			log.Println(err)
		}
		err = database.Collection("biblestudies").Drop(ctx)
		if err != nil {
			log.Println(err)
		}

		jsonfile, err := os.Open("initialUsers.json")
		if err != nil {
			log.Fatal(err)
		}
		defer jsonfile.Close()

		byteValue, _ := ioutil.ReadAll(jsonfile)

		var users Users

		err = json.Unmarshal(byteValue, &users)
		if err != nil {
			log.Fatalln(err)
		}

		userCollection := database.Collection("users")
		booksCollection := database.Collection("biblebooks")
		studyCollection := database.Collection("biblestudies")

		bookMap := make(map[uint]models.BibleBook)

		for _, user := range users.Users {
			user.ID = primitive.NewObjectID()
			user.Creds.SetPassword("InitialPassword")
			user.Creds.MustChange = true
			user.Creds.Locked = false
			_, err := userCollection.InsertOne(ctx, user)
			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Println(len(users.Books))
		for _, book := range users.Books {
			bookMap[book.ID] = book
			_, err := booksCollection.InsertOne(ctx, book)
			if err != nil {
				log.Fatal(err)
			}
		}

		jsonfile, err = os.Open("soapStudy.json")
		if err != nil {
			log.Fatal(err)
		}
		defer jsonfile.Close()

		byteValue, _ = ioutil.ReadAll(jsonfile)

		var study Study

		err = json.Unmarshal(byteValue, &study)
		if err != nil {
			log.Fatal(err)
		}

		for _, st := range study.Studies {
			st.ID = primitive.NewObjectID()
			for _, month := range st.Periods {
				for _, day := range month.StudyDays {
					for k, ref := range day.References {
						book := bookMap[ref.BookID]
						ref.AssignBook(book)
						day.References[k] = ref
					}
				}
			}
			_, err := studyCollection.InsertOne(ctx, st)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
