package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Word struct {
	Word     string `json:"word" bson:"word"`
	ImageURL string `json:"imageUrl" bson:"imageUrl"`
}

type LetterWords struct {
	Letter string `json:"letter" bson:"letter"`
	Words  []Word `json:"words" bson:"words"`
}

var (
	s3Client *s3.Client
	mongoDB  *mongo.Collection
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize AWS S3 client
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		log.Fatal(err)
	}
	s3Client = s3.NewFromConfig(cfg)

	// Initialize MongoDB client with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	// Use the correct database and collection
	mongoDB = client.Database("alphabetgame").Collection("words")

	// Print existing documents count
	count, err := mongoDB.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("Error counting documents: %v", err)
	} else {
		fmt.Printf("Found %d existing documents in words collection\n", count)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	letters := r.MultipartForm.Value["letters"]
	paths := r.MultipartForm.Value["paths"]

	// Group files by letter
	letterGroups := make(map[string][]Word)

	for i, file := range files {
		letter := letters[i]
		path := paths[i]

		src, err := file.Open()
		if err != nil {
			continue
		}
		defer src.Close()

		fileContent, err := io.ReadAll(src)
		if err != nil {
			continue
		}

		// Get word from filename and clean it
		word := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
		word = strings.Title(strings.ToLower(word))

		// Use the full path for S3 key to maintain folder structure
		s3Key := fmt.Sprintf("assets/%s", path)
		_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
			Key:    aws.String(s3Key),
			Body:   bytes.NewReader(fileContent),
		})
		if err != nil {
			continue
		}

		imageURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
			os.Getenv("AWS_BUCKET_NAME"),
			os.Getenv("AWS_REGION"),
			s3Key)

		// Add word to letter group
		letterGroups[letter] = append(letterGroups[letter], Word{
			Word:     word,
			ImageURL: imageURL,
		})
	}

	// Update MongoDB for each letter group
	var results []LetterWords
	for letter, words := range letterGroups {
		// Try to find existing letter document
		var existingDoc LetterWords
		err := mongoDB.FindOne(context.TODO(), bson.M{"letter": letter}).Decode(&existingDoc)

		if err == nil {
			// Update existing letter document
			existingWords := existingDoc.Words
			wordMap := make(map[string]bool)

			// Create map of existing words
			for _, w := range existingWords {
				wordMap[w.Word] = true
			}

			// Add new words or update existing ones
			for _, newWord := range words {
				if _, exists := wordMap[newWord.Word]; exists {
					// Update existing word
					for i, w := range existingWords {
						if w.Word == newWord.Word {
							existingWords[i] = newWord
							break
						}
					}
				} else {
					// Add new word
					existingWords = append(existingWords, newWord)
				}
			}

			_, err = mongoDB.UpdateOne(
				context.TODO(),
				bson.M{"letter": letter},
				bson.M{"$set": bson.M{"words": existingWords}},
			)
		} else {
			// Create new letter document
			letterDoc := LetterWords{
				Letter: letter,
				Words:  words,
			}
			_, err = mongoDB.InsertOne(context.TODO(), letterDoc)
		}

		if err != nil {
			log.Printf("MongoDB operation failed for letter %s: %v", letter, err)
			continue
		}

		results = append(results, LetterWords{
			Letter: letter,
			Words:  words,
		})
	}

	json.NewEncoder(w).Encode(results)
}

func main() {
	http.HandleFunc("/api/upload", uploadHandler)
	http.Handle("/", http.FileServer(http.Dir("static")))

	fmt.Println("Server starting at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
