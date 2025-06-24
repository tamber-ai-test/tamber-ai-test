package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rs/zerolog/log"
)

var (
	mongoURI   = os.Getenv("MONGO_URI")
	s3Endpoint = os.Getenv("S3_ENDPOINT")
	s3Region   = os.Getenv("S3_REGION")
	s3Key      = os.Getenv("AWS_ACCESS_KEY_ID")
	s3Secret   = os.Getenv("AWS_SECRET_ACCESS_KEY")
	s3Bucket   = os.Getenv("S3_BUCKET")
)

var mongoClient *mongo.Client
var s3Client *s3.S3

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h(w, r)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Received upload request")
	r.ParseMultipartForm(10 << 20)

	log.Info().Msg("Parsing form data")
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Error().Err(err).Msg("Error retrieving file")
		http.Error(w, "Invalid file", 400)
		return
	}
	defer file.Close()

	log.Info().Str("filename", handler.Filename).Int64("size", handler.Size).Msg("Uploaded file")

	tempFile, err := os.CreateTemp("", "upload-*.mp3")
	if err != nil {
		log.Error().Err(err).Msg("Error creating temp file")
		http.Error(w, "Can't create temp file", 500)
		return
	}
	defer os.Remove(tempFile.Name())

	io.Copy(tempFile, file)
	tempFile.Seek(0, 0)

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(handler.Filename),
		Body:   tempFile,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload to S3")
		http.Error(w, "Failed to upload to S3", 500)
		return
	}

	coll := mongoClient.Database("audio").Collection("uploads")
	_, err = coll.InsertOne(context.Background(), bson.M{
		"filename": handler.Filename,
		"s3_url":   fmt.Sprintf("s3://%s/%s", s3Bucket, handler.Filename),
	})
	if err != nil {
		log.Error().Err(err).Msg("Mongo insert failed")
		http.Error(w, "Mongo insert failed", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Upload successful")
}

func main() {
	log.Info().Msg("Starting server...")

	var err error
	mongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	log.Info().Msg("Connected to MongoDB")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3Region),
		Endpoint:    aws.String(s3Endpoint),
		Credentials: credentials.NewStaticCredentials(s3Key, s3Secret, ""),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create AWS session")
	}
	s3Client = s3.New(sess)

	http.HandleFunc("/upload", withCORS(uploadHandler))
	http.HandleFunc("/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to the MP3 Upload Service!")
	}))

	log.Info().Msg("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

