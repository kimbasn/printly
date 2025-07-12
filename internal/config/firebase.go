package config

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func SetupFirebase(ctx context.Context, serviceAccountFile string) (*firebase.App, error) {
	var opt option.ClientOption

	if serviceAccountFile != "" {
		opt = option.WithCredentialsFile(serviceAccountFile)
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	return app, err
}
