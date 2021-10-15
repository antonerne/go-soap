package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Token struct {
	TokenUUID primitive.ObjectID `json:"id" bson:"_id"`
	Expires   time.Time          `json:"-" bson:"expires"`
}

// ByContacts will provide the sort interface methods for sorting an employee's
// contact information.
type ByToken []Token

func (s ByToken) Len() int           { return len(s) }
func (s ByToken) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByToken) Less(i, j int) bool { return s[i].TokenUUID.Hex() < s[j].TokenUUID.Hex() }
