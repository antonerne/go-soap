package models

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Email string             `json:"email" bson:"email"`
	Name  Name               `json:"name" bson:"name"`
	Creds Credentials        `json:"creds,omitempty" bson:"creds"`
	Roles []string           `json:"roles,omitempty" bson:"roles"`
}

type Name struct {
	First  string `json:"first" bson:"first"`
	Middle string `json:"middle,omitempty" bson:"middle,omitempty"`
	Last   string `json:"last" bson:"last"`
	Suffix string `json:"suffix,omitempty" bson:"suffix,omitempty"`
}

type Credentials struct {
	Password          string    `json:"-" bson:"password"`
	Expires           time.Time `json:"expires" bson:"expires"`
	MustChange        bool      `json:"mustchange" bson:"mustchange"`
	Locked            bool      `json:"locked" bson:"locked"`
	BadAttempts       int16     `json:"-" bson:"badattempts"`
	Verified          time.Time `json:"-" bson:"verified,omitempty"`
	VerificationToken string    `json:"-" bson:"verificationtoken,omitempty"`
	ResetToken        string    `json:"-" bson:"resettoken,omitempty"`
	ResetExpires      time.Time `json:"-" bson:"resetexpires,omitempty"`
	NewRemoteToken    string    `json:"-" bson:"newremotetoken,omitempty"`
	Remotes           []string  `json:"-" bson:"lastremote,omitempty"`
	PrivateKey        string    `json:"-" bson:"privatekey,omitempty"`
}

// SetPassword function will reset the password to a new value, storing the
// new password in the database
func (c *Credentials) SetPassword(passwd string) (bool, *ErrorMessage) {
	cost := rand.Intn(3) + 10
	bytes, err := bcrypt.GenerateFromPassword([]byte(passwd), cost)
	if err != nil {
		errMsg := ErrorMessage{
			ErrorType:  "password",
			StatusCode: 401,
			Message:    err.Error(),
		}
		return false, &errMsg
	} else {
		c.Password = string(bytes)
		c.Expires = time.Now().Add(time.Hour * 24 * 90)
		c.MustChange = false
		c.BadAttempts = 0
		c.Locked = false
		return true, nil
	}
}

// LogIn function will be used to check the password will match the stored
// password, but the employee's email must be verified, if not reject at the
// start, if verified, compare password, then check badAttempts and expiration.
func (c *Credentials) LogIn(passwd string, remote string) (bool, *ErrorMessage) {
	if c.Verified.Before(time.Date(1970, time.January, 2, 0,
		0, 0, 0, time.UTC)) {
		errMsg := ErrorMessage{
			ErrorType:  "credentials",
			StatusCode: 401,
			Message:    "Account Not Verified",
		}
		return false, &errMsg
	}
	err := bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(passwd))
	if err != nil {
		errMsg := ErrorMessage{
			ErrorType:  "password",
			StatusCode: 401,
			Message:    err.Error(),
		}
		c.BadAttempts++
		if c.BadAttempts > 2 {
			c.Locked = true
		}
		return false, &errMsg
	}
	if c.Expires.Before(time.Now()) {
		errMsg := ErrorMessage{
			ErrorType:  "credentials",
			StatusCode: 401,
			Message:    "Account Expired",
		}
		return false, &errMsg
	}
	if c.BadAttempts > 2 {
		errMsg := ErrorMessage{
			ErrorType:  "credentials",
			StatusCode: 401,
			Message:    "Account Locked",
		}
		c.Locked = true
		return false, &errMsg
	}
	c.BadAttempts = 0
	c.Locked = false
	// check if remote access is from a terminal already loaded.
	for _, site := range c.Remotes {
		if strings.EqualFold(site, remote) {
			return true, nil
		}
	}
	errMsg := ErrorMessage{
		ErrorType:  "new remote",
		StatusCode: 205,
		Message:    "New Remote",
	}
	return true, &errMsg
}

// StartVerification function will start the Email Verification process, by
// creating a random token.
func (c *Credentials) StartVerification() string {
	c.Verified = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	c.VerificationToken = c.randomToken(8)
	return c.VerificationToken
}

// Verify function will be used to complete the Email Verification process, by
// removing the verification token and setting the verification datetime.
func (c *Credentials) Verify(token string) (bool, *ErrorMessage) {
	if c.VerificationToken == token {
		c.Verified = time.Now()
		c.VerificationToken = ""
		return true, nil
	}
	return false, &ErrorMessage{
		ErrorType:  "verification",
		StatusCode: 401,
		Message:    "Verification failure",
	}
}

// New Remote start will create a token for allowing the user to approve a new
// computer or other device to be used.
func (c *Credentials) StartRemoteToken() string {
	c.NewRemoteToken = c.randomToken(7)
	return c.NewRemoteToken
}

func (c *Credentials) VerifyRemoteToken(token string, ipaddr string) (bool, *ErrorMessage) {
	if c.NewRemoteToken == token {
		c.Remotes = append(c.Remotes, ipaddr)
		return true, nil
	}
	return false, &ErrorMessage{
		ErrorType:  "new remote",
		StatusCode: http.StatusUnauthorized,
		Message:    "authorization failure",
	}
}

// StartForgot function will be used to start the reset (forgot) password
// process, creating a token and an expiration date/time.
func (c *Credentials) StartForgot() string {
	c.ResetToken = c.randomToken(8)
	c.ResetExpires = time.Now().Add(time.Hour)
	return c.ResetToken
}

// CreateRandomPassword function will be used to create a temporary password
// to provide to the employee for log in.  It will be passed to their email
// address
func (c *Credentials) CreateRandomKey(length uint) string {
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

// Unlock account function will reset the account to allow an employee to
// log into the system.
func (c *Credentials) Unlock() bool {
	c.BadAttempts = 0
	c.Locked = false
	return true
}

func (c *Credentials) randomToken(size int) string {
	characters := "0123456789"
	token := ""
	for i := 0; i < size; i++ {
		randCh := rand.Intn(len(characters))
		token += characters[randCh : randCh+1]
	}
	return token
}

func (c *Credentials) CreateJWTToken(ID primitive.ObjectID,
	email string, roles []string, key string) (string, *Token, error) {
	t := new(Token)
	t.ID = primitive.NewObjectID()
	t.Expires = time.Now().Add(time.Hour * 3)
	expiry := t.Expires.Unix()
	claims := &JwtClaims{
		Id:         ID.Hex(),
		Email:      email,
		Roles:      roles,
		Exp:        expiry,
		Expires:    c.Expires,
		MustChange: c.MustChange,
		Uuid:       t.ID.Hex(),
		Locked:     c.Locked,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiry,
		},
	}
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		err := godotenv.Load()
		if err != nil {
			return "", nil, err
		}
		secretKey = os.Getenv("JWT_SECRET")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)

	signedToken, err := token.SignedString([]byte(secretKey))
	return signedToken, t, err
}

func (c *Credentials) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("invalid token - %s", token.Header["alg"])
		}
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			err := godotenv.Load()
			if err != nil {
				return nil, err
			}
			secretKey = os.Getenv("JWT_SECRET")
		}
		return []byte(secretKey), nil
	})
}

func (c *Credentials) GetClaims(iClaims map[string]interface{}) *JwtClaims {
	claims := new(JwtClaims)
	for k, v := range iClaims {
		switch strings.ToLower(k) {
		case "id":
			claims.Id = v.(string)
		case "email":
			claims.Email = v.(string)
		case "roles":
			roles := v.([]interface{})
			for i := range roles {
				claims.Roles = append(claims.Roles, roles[i].(string))
			}
		case "expires":
			claims.Expires, _ = time.Parse(time.RFC3339, v.(string))
		case "mustchange":
			claims.MustChange = v.(bool)
		case "uuid":
			claims.Uuid = v.(string)
		case "locked":
			claims.Locked = v.(bool)
		}
	}
	return claims
}
