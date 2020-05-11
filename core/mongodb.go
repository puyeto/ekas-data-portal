package core

import (
	"context"
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

// InitializeMongoDB Initialize MongoDB Connection
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

	Logger.Infof("Mongo DB initialized: %v", dbName)
	return client.Database(dbName)
}

// LogToMongoDB ...
func LogToMongoDB(m models.DeviceData) error {
	collection := MongoDB.Collection("data_" + strconv.FormatInt(int64(m.DeviceID), 10))
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_, err := collection.InsertOne(ctx, m)
	CreateIndexMongo("data_" + strconv.FormatInt(int64(m.DeviceID), 10))
	return err
}

// CreateIndexMongo create a mongodn index
func CreateIndexMongo(colName string) (string, error) {
	mod := mongo.IndexModel{
		Keys: bson.M{
			"datetimestamp": -1, // index in ascending order
		}, Options: nil,
	}
	collection := MongoDB.Collection(colName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return collection.Indexes().CreateOne(ctx, mod)
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
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, bson.M{"_id": m.DeviceID}, data, opts)

	return err
}

// LogCurrentViolationSeenMongoDB update current violation
func LogCurrentViolationSeenMongoDB(m models.DeviceData) error {
	data := bson.M{
		"$set": bson.M{
			"data":         m,
			"datetime":     m.DateTime,
			"datetimeunix": m.DateTimeStamp,
		},
	}

	collection := MongoDB.Collection("current_violations")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, bson.M{"_id": m.DeviceID}, data, opts)

	return err
}
