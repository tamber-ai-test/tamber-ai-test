package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Add to top of main.go
var (
	mongoURI    = os.Getenv("MONGO_URI")              // e.g. mongodb://root:example@mongo:27017
	s3Endpoint  = os.Getenv("S3_ENDPOINT")            // e.g. http://localstack:4566
	s3Region    = os.Getenv("S3_REGION")              // e.g. us-east-1
	s3Key       = os.Getenv("AWS_ACCESS_KEY_ID")      // e.g. test
	s3Secret    = os.Getenv("AWS_SECRET_ACCESS_KEY")  // e.g. test
	s3Bucket    = os.Getenv("S3_BUCKET")              // e.g. my-mp3-bucket
)

var mongoClient *mongo.Client
var s3Client *s3.S3


func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // or restrict to specific origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h(w, r)
	}
}


func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received upload request")
	r.ParseMultipartForm(10 << 20) // 10MB limit

	log.Println("Parsing form data")
	file, handler, err := r.FormFile("file")
	if err != nil {
		println("Error retrieving file:", err)
		http.Error(w, "Invalid file", 400)
		return
	}

	log.Printf("Uploaded file: %s, Size: %d bytes\n", handler.Filename, handler.Size)
	defer file.Close()

	tempFile, err := os.CreateTemp("", "upload-*.mp3")
	if err != nil {
		http.Error(w, "Can't create temp file", 500)
		return
	}
	defer os.Remove(tempFile.Name())

	io.Copy(tempFile, file)
	tempFile.Seek(0, 0)

	bucket := "my-mp3-bucket"
	key := handler.Filename

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   tempFile,
	})
	if err != nil {
		http.Error(w, "Failed to upload to S3", 500)
		return
	}

	coll := mongoClient.Database("audio").Collection("uploads")
	_, err = coll.InsertOne(context.Background(), bson.M{
		"filename": key,
		"s3_url":   fmt.Sprintf("s3://%s/%s", bucket, key),
	})
	if err != nil {
		http.Error(w, "Mongo insert failed", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Upload successful")
}

func main() {
	log.Println("Starting server...")
	var err error
	mongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Println("Failed to connect to MongoDB:", err)
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3Region),
		Endpoint:    aws.String(s3Endpoint),
		Credentials: credentials.NewStaticCredentials(s3Key, s3Secret, ""),
	})

	if err != nil {
		log.Fatal(err)
	}
	s3Client = s3.New(sess)

	http.HandleFunc("/upload", withCORS(uploadHandler))
	http.HandleFunc("/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to the MP3 Upload Service!")
	}))

	http.ListenAndServe(":8080", nil)
}

