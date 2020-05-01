package core

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ekas-data-portal/models"
	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB ...
var MongoDB *mongo.Database

// InitializeMongoCon Initialize MongoDB Connection
func InitializeMongoDB(dbURL, dbName string) *mongo.Database {
	client, err := mongo.NewClient(options.Client().ApplyURI(dbURL))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// defer client.Disconnect(ctx)

	fmt.Println("Mongo DB initialized", dbName)
	return client.Database(dbName)
}

// LogToMongoDB ...
func LogToMongoDB(m models.DeviceData) error {
	collection := MongoDB.Collection("data_" + strconv.FormatInt(int64(m.DeviceID), 10))
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_, err := collection.InsertOne(ctx, m)
	return err
}

// LoglastSeenMongoDB update last seen
func LoglastSeenMongoDB(m models.DeviceData) error {
	data := bson.M{
		"$set": bson.M{
			"last_seen_date": m.DateTime,
			"last_seen_unix": m.DateTimeStamp,
			"updated_at":     time.Now(),
		},
	}

	collection := MongoDB.Collection("a_device_lastseen")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_, err := collection.UpdateOne(ctx, bson.M{"_id": m.DeviceID}, data)
	return err
}
