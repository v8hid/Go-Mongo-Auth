package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBInstance() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var MongoDBURI string = os.Getenv("MONGODB_URL")

	if MongoDBURI == "" {
		log.Fatal("MONGODB_URL is not set on .env file")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDBURI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	return client
}

var Client *mongo.Client = DBInstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	DBName := os.Getenv("DB_NAME")
	if DBName == "" {
		log.Println("DB_NAME is not set on .env file")
		DBName = "cluster0"
	}
	var collection *mongo.Collection = client.Database(DBName).Collection(collectionName)
	return collection
}
