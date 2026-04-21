package notifications

import (
	"context"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

const androidNotificationChannelID = "finance-go-default"

type PushMessage struct {
	Title string
	Body  string
	Data  map[string]string
}

type PushSender interface {
	Send(ctx context.Context, token string, message PushMessage) error
}

type NoopPushSender struct{}

func (NoopPushSender) Send(context.Context, string, PushMessage) error {
	return nil
}

type FirebasePushConfig struct {
	ProjectID       string
	CredentialsJSON string
	CredentialsPath string
}

type FirebaseSender struct {
	client *messaging.Client
}

func NewFirebaseSender(ctx context.Context, cfg FirebasePushConfig) (PushSender, error) {
	var opts []option.ClientOption
	if cfg.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsPath))
	} else if cfg.CredentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
	}

	if cfg.ProjectID == "" {
		cfg.ProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: cfg.ProjectID}, opts...)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FirebaseSender{client: client}, nil
}

func (s *FirebaseSender) Send(ctx context.Context, token string, message PushMessage) error {
	if s == nil || s.client == nil || token == "" {
		return nil
	}

	_, err := s.client.Send(ctx, &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: message.Title,
			Body:  message.Body,
		},
		Data: message.Data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: androidNotificationChannelID,
			},
		},
	})
	return err
}
