package bootstrap

import (
	"context"
	"log"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/mongo"
)

func NewMongoDatabase(env *Env) mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ✅ Lấy connection string trực tiếp từ file .env
	mongodbURI := env.MongoURI
	if mongodbURI == "" {
		log.Fatal("MongoDB URI is not set in environment (.env)")
	}

	client, err := mongo.NewClient(mongodbURI)
	if err != nil {
		log.Fatal("Error creating MongoDB client:", err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	if err := client.Ping(ctx); err != nil {
		log.Fatal("Could not ping MongoDB:", err)
	}

	log.Println("✅ Connected to MongoDB successfully!")
	return client
}

func CloseMongoDBConnection(client mongo.Client) {
	if client == nil {
		return
	}

	if err := client.Disconnect(context.TODO()); err != nil {
		log.Fatal(err)
	}

	log.Println("Connection to MongoDB closed.")
}
