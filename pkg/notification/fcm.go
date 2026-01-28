package notification

import (
	"context"
	"encoding/base64"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"hopSpotAPI/internal/config"
)

type FCMClient struct {
	client *messaging.Client
}

func NewFCMClient(cfg config.Config) (*FCMClient, error) {
	if cfg.FirebaseAuthKey == "" {
		return nil, fmt.Errorf("FIREBASE_AUTH_KEY not set")
	}

	decodedKey, err := base64.StdEncoding.DecodeString(cfg.FirebaseAuthKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode firebase key: %w", err)
	}

	opt := option.WithAuthCredentialsJSON(option.ServiceAccount, decodedKey)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create messaging client: %w", err)
	}

	return &FCMClient{client: fcmClient}, nil
}

// SendToDevice
func (c *FCMClient) SendToDevice(ctx context.Context, token string, title string, body string, data map[string]string) error {
	if token == "" {
		return fmt.Errorf("token is empty")
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := c.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send FCM message: %w", err)
	}

	return nil
}

// SendToMultiple
func (c *FCMClient) SendToMultiple(ctx context.Context, tokens []string, title string, body string, data map[string]string) error {
	if len(tokens) == 0 {
		return fmt.Errorf("tokens list is empty")
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := c.client.SendEachForMulticast(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send FCM messages: %w", err)
	}

	return nil
}
