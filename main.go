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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WordImage struct {
	Letter   string `json:"letter"`
	Word     string `json:"word"`
	ImageURL string `json:"imageUrl"`
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
	mongoDB = client.Database("wordimages").Collection("images")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form data (32MB max)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	letters := r.MultipartForm.Value["letters"]

	var results []WordImage
	for i, file := range files {
		letter := letters[i]

		// Open the uploaded file
		src, err := file.Open()
		if err != nil {
			continue
		}
		defer src.Close()

		// Read file content
		fileContent, err := io.ReadAll(src)
		if err != nil {
			continue
		}

		// Get word from filename
		word := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))

		// Upload to S3
		s3Key := fmt.Sprintf("assets/%s/%s", letter, file.Filename)
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

		wordImage := WordImage{
			Letter:   letter,
			Word:     word,
			ImageURL: imageURL,
		}

		// Save to MongoDB
		_, err = mongoDB.InsertOne(context.TODO(), wordImage)
		if err != nil {
			continue
		}

		results = append(results, wordImage)
	}

	json.NewEncoder(w).Encode(results)
}

func main() {
	http.HandleFunc("/api/upload", uploadHandler)
	http.Handle("/", http.FileServer(http.Dir("static")))

	fmt.Println("Server starting at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
