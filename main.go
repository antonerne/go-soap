package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/antonerne/go-soap/models"
	"github.com/google/uuid"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DBHOST"), os.Getenv("DBUSER"), os.Getenv("DBPASSWD"),
		os.Getenv("DATABASE"), os.Getenv("DBPORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(
		&models.BibleBook{},
		&models.BibleStudy{},
		&models.BibleStudyPeriod{},
		&models.BibleStudyDay{},
		&models.BibleStudyDayReference{},
	)

	db.AutoMigrate(
		&models.User{},
		&models.UserRemote{},
		&models.Name{},
		&models.Credentials{},
		&models.Token{},
	)

	db.AutoMigrate(
		&models.UserBibleStudy{},
		&models.UserBibleStudyPeriod{},
		&models.UserBibleStudyDay{},
		&models.UserBibleStudyReference{},
	)

	if loadData {

		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM biblebooks")
		db.Exec("DELETE FROM biblestudies")

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

		bookMap := make(map[uint]models.BibleBook)

		for _, user := range users.Users {
			user.ID = uuid.NewString()
			user.Name.UserID = user.ID
			user.Creds.UserID = user.ID
			user.Creds.SetPassword("InitialPassword")
			user.Creds.MustChange = true
			user.Creds.Locked = false
			db.Create(&user)
		}

		for _, book := range users.Books {
			db.Create(&book)
			bookMap[book.ID] = book
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
			db.Create(&st)
		}
	}
}
