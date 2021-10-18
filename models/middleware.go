package models

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthorizeJWT(db *gorm.DB, log *LogFile) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > len(BEARER_SCHEMA) {
			tokenString := authHeader[len(BEARER_SCHEMA):]
			creds := new(Credentials)
			token, err := creds.ValidateToken(tokenString)
			if token.Valid {
				claims := creds.GetClaims(token.Claims.(jwt.MapClaims))
				var dbToken Token
				db.Find(&dbToken, "id = ?", claims.Uuid)
				if dbToken.Expires.Before(time.Now()) {
					c.AbortWithStatus(http.StatusUnauthorized)
				}
				c.Writer.Header().Set("userid", claims.Id)
				editor := "user"
				if claims.Editor {
					editor = "editor"
				}
				c.Writer.Header().Set("roles", editor)
				c.Next()
				return
			}
			log.WriteToLog(err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		log.WriteToLog("No Authorization header")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func AuthorizeEditor(db *gorm.DB, log *LogFile) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > len(BEARER_SCHEMA) {
			tokenString := authHeader[len(BEARER_SCHEMA):]
			creds := new(Credentials)
			token, err := creds.ValidateToken(tokenString)
			if token.Valid {
				claims := creds.GetClaims(token.Claims.(jwt.MapClaims))
				var dbToken Token
				db.Find(&dbToken, "id = ?", claims.Uuid)
				if dbToken.Expires.Before(time.Now()) {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
				if !claims.Editor {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "Not Editor",
					})
					return
				}
				c.Writer.Header().Set("userid", claims.Id)
				c.Next()
				return
			}
			log.WriteToLog(err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		log.WriteToLog("No Authorization header")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
