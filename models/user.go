package models

import (
	rd "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID      string           `json:"id" gorm:"primaryKey;column:id"`
	Email   string           `json:"email" gorm:"column:email"`
	Editor  bool             `json:"editor,omitempty" gorm:"column:editor"`
	Name    Name             `json:"name" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Creds   Credentials      `json:"creds,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Studies []UserBibleStudy `json:"studies" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) HasRemote(ipaddress string) bool {
	answer := false
	for _, rip := range u.Remotes {
		if rip.RemoteIP == ipaddress {
			answer = true
		}
	}
	return answer
}

func (u *User) VerifyRemoteToken(token string, ipaddr string) (bool, *ErrorMessage) {
	if u.Creds.NewRemoteToken == token {
		return true, nil
	}
	return false, &ErrorMessage{
		ErrorType:  "new remote",
		StatusCode: http.StatusUnauthorized,
		Message:    "authorization failure",
	}
}

type UserRemote struct {
	ID       uint64 `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	UserID   string `json:"-" gorm:"column:userid"`
	RemoteIP string `json:"remote_ip" gorm:"column:remote_ip"`
}

func (UserRemote) TableName() string {
	return "user_remotes"
}

type Name struct {
	UserID string `json:"-" gorm:"primaryKey;column:userid"`
	First  string `json:"first" gorm:"column:first"`
	Middle string `json:"middle,omitempty" gorm:"column:middle"`
	Last   string `json:"last" gorm:"column:last"`
	Suffix string `json:"suffix,omitempty" gorm:"column:suffix"`
}

func (Name) TableName() string {
	return "user_names"
}

func (n *Name) FullName() string {
	answer := n.First
	if n.Middle != "" {
		answer += fmt.Sprintf(" %s", n.Middle)
	}
	answer += fmt.Sprintf(" %s", n.Last)
	return answer
}

type Credentials struct {
	UserID            string       `json:"-" gorm:"primaryKey;column:userid"`
	Password          string       `json:"-" gorm:"column:password"`
	Expires           time.Time    `json:"expires" gorm:"column:expires"`
	MustChange        bool         `json:"mustchange" gorm:"column:mustchange"`
	Locked            bool         `json:"locked" gorm:"column:locked"`
	BadAttempts       int16        `json:"-" gorm:"column:badattempts"`
	Verified          time.Time    `json:"-" gorm:"column:verified"`
	VerificationToken string       `json:"-" gorm:"column:verificationtoken"`
	ResetToken        string       `json:"-" gorm:"column:resettoken"`
	ResetExpires      time.Time    `json:"-" gorm:"column:resetexpires"`
	NewRemoteToken    string       `json:"-" gorm:"column:newremotetoken"`
	Remotes           []UserRemote `json:"remotes" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PrivateKey        string       `json:"-" gorm:"column:privatekey"`
}

func (Credentials) TableName() string {
	return "user_credentials"
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
	if !c.HasRemote(remote) {
		errMsg := ErrorMessage{
			ErrorType:  "new remote",
			StatusCode: 205,
			Message:    "New Remote",
		}
		return true, &errMsg
	}
	c.BadAttempts = 0
	c.Locked = false
	return true, nil
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
	fmt.Printf("DB Token: %s - Passed Token: %s\n", c.VerificationToken, token)
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

func (c *Credentials) HasRemote(ipaddress string) bool {
	answer := false
	for _, rip := range c.Remotes {
		if rip.RemoteIP == ipaddress {
			answer = true
		}
	}
	return answer
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
		pos, _ := rd.Int(rd.Reader, big.NewInt(int64(len(characters))))
		token += characters[pos.Int64() : pos.Int64()+1]
	}
	return token
}

func (c *Credentials) CreateJWTToken(ID string, email string, editor bool,
	key string) (string, *Token, error) {
	t := new(Token)
	t.ID = uuid.NewString()
	t.Expires = time.Now().Add(time.Hour * 3)
	expiry := t.Expires.Unix()
	claims := &JwtClaims{
		Id:         ID,
		Email:      email,
		Editor:     editor,
		Exp:        expiry,
		Expires:    c.Expires,
		MustChange: c.MustChange,
		Uuid:       t.ID,
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
		case "editor":
			claims.Editor = v.(bool)
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
