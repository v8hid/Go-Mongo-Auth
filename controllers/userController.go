package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"v8hid/Go-Mongo-Auth/database"
	helper "v8hid/Go-Mongo-Auth/helpers"

	"v8hid/Go-Mongo-Auth/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ucn string = os.Getenv("USERS_COLLECTION_NAME")
var userCollection *mongo.Collection = database.OpenCollection(database.Client, ucn)

func Singup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request format is invalid"})
			return
		}
		role := "USER"
		user.Role = &role
		var validate = validator.New()
		validationErr := validate.Struct(user)

		if validationErr != nil {

			errors := validationErr.(validator.ValidationErrors)
			finalErrors := helper.MakeValidationErrors(errors)
			c.JSON(http.StatusBadRequest, gin.H{"errors": finalErrors})
			return
		}
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error, something went wrong"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This email already exist"})
			return
		}
		password := helper.HashPassword(*user.Password)
		user.Password = &password

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		userId := user.ID.Hex()
		user.User_id = &userId

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, "USER", *user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		_, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := "Unexpected error, user item can not be created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		response := models.UserRes{
			First_name:   user.First_name,
			Last_name:    user.Last_name,
			Email:        user.Email,
			Token:        user.Token,
			RefreshToken: user.Refresh_token,
			User_id:      &userId,
		}
		c.JSON(http.StatusOK, response)
		return
	}
}
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User
		defer cancel()
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request format is invalid"})
			return
		}
		var validate = validator.New()
		validationErr := validate.StructPartial(user, "Email", "Password")
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errors": "Email or password is incorrect"})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email or password is incorrect"})
			return
		}
		passwordIsMatch := helper.VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsMatch {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email or password is incorrect"})
			return
		}
		token, refreshToken, err := helper.GenerateAllTokens(*foundUser.Email, *foundUser.Role, *foundUser.User_id)
		if err != nil {
			log.Print(err)
		}
		helper.UpdateAllTokens(token, refreshToken, *foundUser.User_id)
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response := models.UserRes{
			First_name:   foundUser.First_name,
			Last_name:    foundUser.Last_name,
			Email:        foundUser.Email,
			Token:        foundUser.Token,
			RefreshToken: foundUser.Refresh_token,
			User_id:      foundUser.User_id,
		}
		c.JSON(http.StatusOK, &response)
	}
}
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}
		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}}}}}
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
			return
		}
		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allUsers[0])
	}

}
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
