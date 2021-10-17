package models

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AuthorizeJWT(db *mongo.Database, log *LogFile) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > len(BEARER_SCHEMA) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			tokens := db.Collection("tokens")
			tokenString := authHeader[len(BEARER_SCHEMA):]
			creds := new(Credentials)
			token, err := creds.ValidateToken(tokenString)
			if token.Valid {
				claims := creds.GetClaims(token.Claims.(jwt.MapClaims))
				var dbToken Token
				filter := bson.D{primitive.E{Key: "_id", Value: claims.Uuid}}
				tokens.FindOne(ctx, filter).Decode(&dbToken)
				if dbToken.Expires.Before(time.Now()) {
					c.AbortWithStatus(http.StatusUnauthorized)
				}
				c.Writer.Header().Set("userid", claims.Id.Hex())
				roles := ""
				for _, r := range claims.Roles {
					if roles != "" {
						roles += ","
					}
					roles += r
				}
				c.Writer.Header().Set("roles", roles)
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

func AuthorizeRole(roles []string, db *mongo.Database, log *LogFile) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > len(BEARER_SCHEMA) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			tokens := db.Collection("tokens")
			tokenString := authHeader[len(BEARER_SCHEMA):]
			creds := new(Credentials)
			token, err := creds.ValidateToken(tokenString)
			if token.Valid {
				claims := creds.GetClaims(token.Claims.(jwt.MapClaims))
				var dbToken Token
				filter := bson.D{primitive.E{Key: "_id", Value: claims.Uuid}}
				tokens.FindOne(ctx, filter).Decode(&dbToken)
				if dbToken.Expires.Before(time.Now()) {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
				found := false
				for _, role := range roles {
					if strings.ToLower(role) == "user" {
						found = true
					} else {
						for _, r := range claims.Roles {
							if strings.EqualFold(r, role) {
								found = true
							}
						}
					}
				}
				if !found {
					role := ""
					for _, r := range roles {
						if role != "" {
							role += ","
						}
						role += r
					}
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "Not " + role,
					})
					return
				}
				c.Writer.Header().Set("userid", claims.Id.Hex())
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
