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

var mongoClient *mongo.Client
var s3Client *s3.S3

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10MB limit

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", 400)
		return
	}
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
	var err error
	mongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://root:example@mongo:27017"))
	if err != nil {
		log.Fatal(err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String("http://localstack:4566"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	})
	if err != nil {
		log.Fatal(err)
	}
	s3Client = s3.New(sess)

	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(":8080", nil)
}

