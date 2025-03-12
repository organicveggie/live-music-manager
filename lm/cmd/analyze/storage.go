package analyze

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	databaseName   = "lm"
	collectionName = "tracks"
)

type StorageHandler struct {
	mongoURI   string
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

func newStorageHandler(mongoURI string) (*StorageHandler, error) {
	h := StorageHandler{
		mongoURI: mongoURI,
	}

	var err error
	h.client, err = mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("error establishing connection to MongoDB at %q: %v", mongoURI, err)
	}

	h.db = h.client.Database(databaseName)
	h.collection = h.db.Collection(collectionName)

	return &h, nil
}

func (sh *StorageHandler) Close(ctx context.Context) error {
	return sh.client.Disconnect(ctx)
}

func (sh *StorageHandler) SaveMetadata(ctx context.Context, metadata *Metadata) error {
	if _, err := sh.collection.InsertOne(ctx, metadata); err != nil {
		return fmt.Errorf("error saving metadata to MongoDB: %v", err)
	}
	return nil
}
