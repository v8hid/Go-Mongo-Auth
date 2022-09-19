package helpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"v8hid/Go-Mongo-Auth/database"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SingedDetails struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

var ucn string = os.Getenv("USERS_COLLECTION_NAME")
var sk string = os.Getenv("SECRET_KEY")
var userCollection *mongo.Collection = database.OpenCollection(database.Client, ucn)

func ValidateToken(signedToken string) (claims *SingedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SingedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(sk), nil
		},
	)
	if err != nil {
		msg = err.Error()
		return
	}
	claims, ok := token.Claims.(*SingedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprint("token is expired")
	}
	return claims, msg
}
func GenerateAllTokens(email string, role string, uid string) (signedToken string, signedRefreshToken string, error error) {

	claims := &SingedDetails{
		email,
		role,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	refreshClaims := struct {
		Uid string
		jwt.StandardClaims
	}{
		uid,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(720)).Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(sk))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(sk))

	if err != nil {
		log.Panic(err)
	}
	return token, refreshToken, err
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	var updateObj primitive.D
	defer cancel()
	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: updateObj},
		},
		&opt,
	)
	if err != nil {
		log.Panic(err)
		return
	}
	return
}
