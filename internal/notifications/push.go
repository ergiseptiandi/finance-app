package notifications

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

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
	client    *messaging.Client
	projectID string
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

	log.Printf("firebase push enabled for project %s", cfg.ProjectID)

	return &FirebaseSender{
		client:    client,
		projectID: cfg.ProjectID,
	}, nil
}

func (s *FirebaseSender) Send(ctx context.Context, token string, message PushMessage) error {
	if s == nil || s.client == nil || token == "" {
		return nil
	}

	msg := buildFirebaseMessage(token, message)
	logFirebaseSendRequest(s.projectID, msg)

	messageID, err := s.client.Send(ctx, msg)
	if err != nil {
		log.Printf("fcm send response: project_id=%s error=%v", s.projectID, err)
		return err
	}

	log.Printf("fcm send response: project_id=%s message_id=%s", s.projectID, messageID)
	return err
}

func buildFirebaseMessage(token string, message PushMessage) *messaging.Message {
	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: message.Title,
			Body:  message.Body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: androidNotificationChannelID,
				Sound:     "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: message.Title,
						Body:  message.Body,
					},
					Sound: "default",
					Badge: nil,
				},
			},
		},
		Data: message.Data,
	}

	if msg.Data == nil {
		msg.Data = map[string]string{}
	}

	return msg
}

func logFirebaseSendRequest(projectID string, msg *messaging.Message) {
	if msg == nil {
		return
	}

	payload := map[string]any{
		"project_id": projectID,
		"message": map[string]any{
			"token": msg.Token,
			"notification": map[string]any{
				"title": msg.Notification.Title,
				"body":  msg.Notification.Body,
			},
			"android": map[string]any{
				"priority": msg.Android.Priority,
				"notification": map[string]any{
					"channel_id": msg.Android.Notification.ChannelID,
				},
			},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("fcm send request: project_id=%s marshal_error=%v", projectID, err)
		return
	}

	log.Printf("fcm send request: %s", raw)
}

// IsTokenInvalid checks whether an FCM error indicates the push token is
// no longer valid (unregistered, expired, or malformed). When true the
// caller should clear the stored token so we stop sending to it.
func IsTokenInvalid(err error) bool {
	if err == nil {
		return false
	}
	if messaging.IsUnregistered(err) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "NOT_FOUND") ||
		strings.Contains(msg, "INVALID_ARGUMENT") ||
		strings.Contains(msg, "not a valid FCM registration token")
}
