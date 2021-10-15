package models

import (
	"crypto/aes"
	"crypto/cipher"
	rd "crypto/rand"
	"errors"
	"io"
	rand "math/rand"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntryReference struct {
	Book      string `json:"book" bson:"book"`
	Chapter   uint8  `json:"chapter" bson:"chapter"`
	VerseList string `json:"verses" bson:"verses"`
}

type EntryText struct {
	Encrypted bool   `json:"encrypted" bson:"encrypted"`
	EntryText string `json:"entrytext" bson:"entrytext"`
}

func (et *EntryText) EncryptText(privkey string, entrykey string) error {
	if !et.Encrypted {
		tempkey := []byte(entrykey)
		c, err := aes.NewCipher([]byte(privkey))
		if err != nil {
			return err
		}

		gcm, err := cipher.NewGCM(c)
		if err != nil {
			return err
		}

		nonceSize := gcm.NonceSize()
		if len(entrykey) < nonceSize {
			return errors.New("entry key too small")
		}

		nonce, cipherkey := tempkey[:nonceSize], tempkey[nonceSize:]
		key, err := gcm.Open(nil, nonce, cipherkey, nil)
		if err != nil {
			return err
		}

		c, err = aes.NewCipher(key)
		if err != nil {
			return err
		}

		gcm, err = cipher.NewGCM(c)
		if err != nil {
			return err
		}

		nonce = make([]byte, gcm.NonceSize())
		if _, err = io.ReadFull(rd.Reader, nonce); err != nil {
			return err
		}

		text := []byte(et.EntryText)

		et.EntryText = string(gcm.Seal(nonce, nonce, text, nil))
		et.Encrypted = true
	}
	return nil
}

func (et *EntryText) DecryptText(privkey string, entrykey string) error {
	if et.Encrypted {
		tempkey := []byte(entrykey)
		c, err := aes.NewCipher([]byte(privkey))
		if err != nil {
			return err
		}

		gcm, err := cipher.NewGCM(c)
		if err != nil {
			return err
		}

		nonceSize := gcm.NonceSize()
		if len(entrykey) < nonceSize {
			return errors.New("entry key too small")
		}

		nonce, cipherkey := tempkey[:nonceSize], tempkey[nonceSize:]
		key, err := gcm.Open(nil, nonce, cipherkey, nil)
		if err != nil {
			return err
		}

		c, err = aes.NewCipher(key)
		if err != nil {
			return err
		}

		gcm, err = cipher.NewGCM(c)
		if err != nil {
			return err
		}

		text := []byte(et.EntryText)
		nonceSize = gcm.NonceSize()
		if len(text) < nonceSize {
			return errors.New("encrypted text too small")
		}

		plainText, err := gcm.Open(nil, nonce, text, nil)
		if err != nil {
			return err
		}
		et.EntryText = string(plainText)
		et.Encrypted = false
	}
	return nil
}

type Entry struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	UserID       primitive.ObjectID `json:"user" bson:"userid"`
	Key          string             `json:"-" bson:"privacy"`
	EntryDate    time.Time          `json:"entrydate" bson:"entrydate"`
	Title        string             `json:"title" bson:"title"`
	Reference    EntryReference     `json:"reference" bson:"reference"`
	Scripture    EntryText          `json:"scripture" bson:"scripture"`
	Observations EntryText          `json:"observations" bson:"observations"`
	Application  EntryText          `json:"application" bson:"application"`
	Prayer       EntryText          `json:"prayer" bson:"prayer"`
}

func (e *Entry) CreateEntryKey(userkey string) error {
	c, err := aes.NewCipher([]byte(userkey))
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rd.Reader, nonce); err != nil {
		return err
	}

	privKey := []byte(e.CreateRandomKey())
	e.Key = string(gcm.Seal(nonce, nonce, privKey, nil))
	return nil
}

func (e *Entry) CreateRandomKey() string {
	lower := "abcdefghijklmnopqrstuvwxyz"
	randPasswd := ""
	for len([]byte(randPasswd)) < 32 {
		j := rand.Intn(3)
		if j == 0 {
			lPos := rand.Intn(26)
			randPasswd += lower[lPos : lPos+1]
		} else if j == 1 {
			uPos := rand.Intn(26)
			randPasswd += strings.ToUpper(lower[uPos : uPos+1])
		} else {
			randPasswd += strconv.FormatInt(rand.Int63n(9), 10)
		}
	}
	return randPasswd
}

func (e *Entry) SetEntryText(field string, text string, userkey string) error {
	var err error
	switch strings.ToLower(field)[:1] {
	case "s":
		e.Scripture.Encrypted = false
		e.Scripture.EntryText = text
		err = e.Scripture.EncryptText(userkey, e.Key)
	case "o":
		e.Observations.Encrypted = false
		e.Observations.EntryText = text
		err = e.Observations.EncryptText(userkey, e.Key)
	case "a":
		e.Application.Encrypted = false
		e.Application.EntryText = text
		err = e.Application.EncryptText(userkey, e.Key)
	case "p":
		e.Prayer.Encrypted = false
		e.Prayer.EntryText = text
		e.Prayer.EncryptText(userkey, e.Key)
	}
	if err != nil {
		return err
	}
	return nil
}

func NewEntry(user primitive.ObjectID, userKey string, entryDate time.Time) (*Entry, error) {
	answer := Entry{
		ID:        primitive.NewObjectID(),
		UserID:    user,
		EntryDate: entryDate,
	}
	err := answer.CreateEntryKey(userKey)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}
