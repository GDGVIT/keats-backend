package firebaseclient

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/storage"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

var client *auth.Client = nil
var bucket *storage.Client = nil

func getApp() (*firebase.App, error) {
	opt := option.WithCredentialsFile(viper.GetString("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase app initialization failed: %v", err)
	}
	return app, nil
}

func GetClient() (*auth.Client, error) {
	if client != nil {
		return client, nil
	}
	app, err := getApp()
	if err != nil {
		return nil, err
	}
	client, err = app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("firebase auth client initialization failed: %v", err)
	}
	return client, nil
}

func GetBucket() (*storage.Client, error) {
	if bucket != nil {
		return bucket, nil
	}
	app, err := getApp()
	if err != nil {
		return nil, err
	}
	bucket, err := app.Storage(context.Background())
	if err != nil {
		return nil, fmt.Errorf("firebase bucket client initialization failed: %v", err)
	}
	return bucket, nil
}

func WriteObject(file *multipart.File, acceptedType []string) (string, error) {
	bucketName := viper.GetString("FIREBASE_BUCKET_NAME")
	fileData := make([]byte, 512)
	_, err := io.ReadAtLeast(*file, fileData, 512)
	if err != nil {
		return "", fmt.Errorf("file parse error")
	}
	// Resets file pointer
	_, err = (*file).Seek(0, 0)
	if err != nil {
		return "", err
	}
	var check = false
	contentType := http.DetectContentType(fileData)
	for _, x := range acceptedType {
		if contentType == x {
			check = true
			break
		}
	}
	if !check {
		return "", fmt.Errorf("invalid file type")
	}
	bucketClient, err := GetBucket()
	if err != nil {
		return "", err
	}
	bucket, err := bucketClient.Bucket(bucketName)
	if err != nil {
		return "", err
	}
	fid := uuid.NewString()
	filePath := "public/" + fid
	wc := bucket.Object(filePath).NewWriter(context.Background())
	if _, err = io.Copy(wc, *file); err != nil {
		return "", err
	}
	if err = wc.Close(); err != nil {
		return "", err
	}

	fileURL := "https://firebasestorage.googleapis.com/v0/b/" + bucketName + "/o/public%2f" + fid + "?alt=media"
	return fileURL, nil
}
