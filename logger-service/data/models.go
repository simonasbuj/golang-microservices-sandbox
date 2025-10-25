package data

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func New(mongo *mongo.Client) Models {
	client = mongo

	return Models{
		LogEntry: LogEntry{},
	}
}

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID			string		`bson:"_id,omitempty" json:"id,omitempty"`
	Name 		string 		`bson:"name" json:"name"`
	Data 		string		`bson:"data" json:"data"`
	CreatedAt 	time.Time 	`bson:"created_at" json:"created_at"`
	UpdatedAt 	time.Time 	`bson:"updated_at" json:"updated_at"`
}

func (l *LogEntry) Insert(entry LogEntry) error {
	collection := client.Database("logs").Collection("logs")

	_, err := collection.InsertOne(context.TODO(), LogEntry{
		Name: 		entry.Name,
		Data:	 	entry.Data,
		CreatedAt: 	time.Now(),
		UpdatedAt: 	time.Now(),
	})
	if err != nil {
		log.Println("error inserting into logs: ", err)
		return fmt.Errorf("failed to insert log into mongo: %w", err)
	}

	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 15)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Println("finding all docs error: ", err)
		return nil, fmt.Errorf("failed to fetch logs from mongo: %w", err)
	}

	defer cursor.Close(ctx)

	var logs []*LogEntry

	for cursor.Next(ctx) {
		var item LogEntry

		err := cursor.Decode(item)
		if err != nil {
			log.Println("error decoding log into slice: ", err)
			return nil, fmt.Errorf("failed to decode log into slice: %w", err)
		}

		logs = append(logs, &item)
	}

	return logs, nil
}

func (l *LogEntry) GetOne(id string) (*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 15)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo _id from hex: %w", err)
	}

	var log LogEntry
	err = collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&log)
	if err != nil {
		return  nil, fmt.Errorf("failed to fetch and decode log entry: %w", err)
	}

	return &log, nil
}

func (l *LogEntry) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 15)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	err := collection.Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	return nil
}

func (l *LogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 15)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	docID, err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo _id from hex: %w", err)
	}

	result, err := collection.UpdateOne(
		ctx, 
		bson.M{"_id": docID},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "name", 		Value: l.Name},
				{Key: "data", 		Value: l.Data},
				{Key: "updated_at", Value: time.Now()},
			}},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update log entry: %w", err)
	}

	return result, nil
}