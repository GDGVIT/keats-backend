package firebaseclient

import(
	"fmt"
	"context"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
	"github.com/spf13/viper"
)

var client *auth.Client = nil

func GetClient() (*auth.Client ,error) {
	if client != nil{
		return client,nil
	}
	opt := option.WithCredentialsFile(viper.GetString("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil,fmt.Errorf("Firebase app initialization failed:",err)
	}
	client, err = app.Auth(context.Background())
	if err != nil {
		return nil,fmt.Errorf("Firebase auth client initialization failed:",err)
	}
	return client,nil
}

