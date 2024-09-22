package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil)) // Hashing and returning as hex string
}

func main() {
	users := map[string]string{
		"admin":      "eUbP9shywUygMx7u",
		"trapLord92": "peUbP9shywUygMx7u2",
	}

	ctx := context.Background()

	// Debug: Check MongoDB URI
	mongoURI := os.Getenv("MONGO_URI")
	fmt.Println("MongoDB URI:", mongoURI)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	// Ping MongoDB to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal("Error pinging MongoDB:", err)
	}

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")

	for username, password := range users {
		// Hash the password
		hashedPassword := hashPassword(password)

		// Debug: Show original and hashed passwords
		fmt.Printf("User: %s, Original Password: %s, Hashed Password: %s\n", username, password, hashedPassword)

		// Store the hashed password
		_, err := collection.InsertOne(ctx, bson.M{
			"username": username,
			"password": hashedPassword, // Storing as hex string
		})
		if err != nil {
			log.Fatal("Error inserting user:", err)
		}
	}

	log.Println("Users inserted successfully")
}
