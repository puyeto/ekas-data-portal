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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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

	return upsert(data, m.DeviceID, "a_device_lastseen")
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

	return upsert(data, m.DeviceID, "current_violations")
}

func upsert(data bson.M, deviceID uint32, table string) error {
	collection := MongoDB.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, bson.M{"_id": deviceID}, data, opts)

	return err
}
