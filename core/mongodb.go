package core

import (
	"context"
	"fmt"
	"log"
	"time"

	// ...
	_ "github.com/go-sql-driver/mysql"
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
