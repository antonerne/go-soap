package models

import (
	"crypto/aes"
	"crypto/cipher"
	rd "crypto/rand"
	"errors"
	"io"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type EntryReference struct {
	ID        uint64 `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	EntryID   string `json:"-" gorm:"column:entry_id"`
	Book      string `json:"book" gorm:"column:book"`
	Chapter   uint8  `json:"chapter" gorm:"column:chapter"`
	VerseList string `json:"verses" gorm:"column:verses"`
}

type EntryText struct {
	ID        uint64 `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	EntryID   string `json:"-" gorm:"column:entry_id"`
	TextType  string `json:"texttype" gorm:"column:text_type"`
	Encrypted bool   `json:"encrypted" gorm:"column:encrypted"`
	EntryText string `json:"entrytext" gorm:"column:entrytext"`
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
	ID        string           `json:"id" gorm:"primaryKey;column:id"`
	UserID    string           `json:"user" gorm:"column:user_id"`
	Key       string           `json:"-" gorm:"column:privacy"`
	EntryDate time.Time        `json:"entrydate" bson:"entrydate"`
	Title     string           `json:"title" bson:"title"`
	Reference []EntryReference `json:"reference" bson:"reference"`
	Texts     []EntryText      `json:"texts" bson:"scripture"`
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
		j, _ := rd.Int(rd.Reader, big.NewInt(3))
		if j.Int64() == 0 {
			lPos, _ := rd.Int(rd.Reader, big.NewInt(26))
			randPasswd += lower[lPos.Int64() : lPos.Int64()+1]
		} else if j.Int64() == 1 {
			uPos, _ := rd.Int(rd.Reader, big.NewInt(26))
			randPasswd += strings.ToUpper(lower[uPos.Int64() : uPos.Int64()+1])
		} else {
			num, _ := rd.Int(rd.Reader, big.NewInt(10))
			randPasswd += strconv.FormatInt(num.Int64(), 10)
		}
	}
	return randPasswd
}

func (e *Entry) SetEntryText(field string, text string, userkey string) error {
	var err error
	found := false
	for i, txt := range e.Texts {
		if !found && strings.EqualFold(txt.TextType, field) {
			found = true
			txt.Encrypted = false
			txt.EntryText = text
			err = txt.EncryptText(userkey, e.Key)
			e.Texts[i] = txt
		}
	}
	if !found {
		txt := EntryText{}
		txt.Encrypted = false
		txt.EntryText = text
		txt.TextType = field
		txt.EntryID = e.ID
		err = txt.EncryptText(userkey, e.Key)
		e.Texts = append(e.Texts, txt)
	}
	if err != nil {
		return err
	}
	return nil
}

func NewEntry(user string, userKey string, entryDate time.Time) (*Entry, error) {
	answer := Entry{
		ID:        uuid.NewString(),
		UserID:    user,
		EntryDate: entryDate,
	}
	err := answer.CreateEntryKey(userKey)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}
