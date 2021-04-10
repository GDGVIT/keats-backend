package firebaseclient

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/storage"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

var client *auth.Client = nil
var bucket *storage.Client = nil

func GetClient() (*auth.Client, error) {
	if client != nil {
		return client, nil
	}
	opt := option.WithCredentialsFile(viper.GetString("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase app initialization failed: %v", err)
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
	opt := option.WithCredentialsFile(viper.GetString("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase app initialization failed: %v", err)
	}
	bucket, err := app.Storage(context.Background())
	if err != nil {
		return nil, fmt.Errorf("firebase bucket client initialization failed: %v", err)
	}
	return bucket, nil
}
