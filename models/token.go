package models

import (
	"time"
)

type Token struct {
	ID      string    `json:"id" gorm:"primaryKey;column:id"`
	Expires time.Time `json:"-" gorm:"column:expires"`
}

func (Token) TableName() string {
	return "user_tokens"
}

// ByContacts will provide the sort interface methods for sorting an employee's
// contact information.
type ByToken []Token

func (s ByToken) Len() int           { return len(s) }
func (s ByToken) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByToken) Less(i, j int) bool { return s[i].ID < s[j].ID }
