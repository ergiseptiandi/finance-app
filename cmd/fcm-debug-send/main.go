package main

import (
	"context"
	"encoding/json"
	"fmt"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"finance-backend/internal/config"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
	"github.com/joho/godotenv"
)

const androidChannelID = "finance-go-default"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("no .env file loaded; using system environment")
	}

	token := flag.String("token", "", "target FCM device token")
	mode := flag.String("mode", "android-channel", "payload mode: notification-only | android-priority | android-channel")
	title := flag.String("title", "FCM Debug", "notification title")
	body := flag.String("body", "Minimal backend FCM debug send.", "notification body")
	timeout := flag.Duration("timeout", 15*time.Second, "send timeout")
	flag.Parse()

	if strings.TrimSpace(*token) == "" {
		log.Fatal("token is required: go run ./cmd/fcm-debug-send -token <FCM_DEVICE_TOKEN>")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client, projectID, err := newMessagingClient(ctx, cfg)
	if err != nil {
		log.Fatalf("firebase sender init failed: %v", err)
	}

	msg, err := buildDebugMessage(strings.TrimSpace(*mode), strings.TrimSpace(*token), strings.TrimSpace(*title), strings.TrimSpace(*body))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("fcm debug send: project_id=%s mode=%s token=%s", projectID, strings.TrimSpace(*mode), redactToken(*token))
	logDebugPayload(projectID, msg)

	messageID, err := client.Send(ctx, msg)
	if err != nil {
		log.Fatalf("fcm debug send failed: %v", err)
	}

	log.Printf("fcm debug send completed: project_id=%s mode=%s message_id=%s", projectID, strings.TrimSpace(*mode), messageID)
}

func redactToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 16 {
		return token
	}
	return token[:8] + "..." + token[len(token)-8:]
}

func buildDebugMessage(mode, token, title, body string) (*messaging.Message, error) {
	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	switch mode {
	case "notification-only":
		return msg, nil
	case "android-priority":
		msg.Android = &messaging.AndroidConfig{
			Priority: "high",
		}
		return msg, nil
	case "android-channel":
		msg.Android = &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: androidChannelID,
			},
		}
		return msg, nil
	default:
		return nil, fmt.Errorf("invalid mode %q; use notification-only, android-priority, or android-channel", mode)
	}
}

func newMessagingClient(ctx context.Context, cfg config.Config) (*messaging.Client, string, error) {
	projectID := cfg.Push.FirebaseProjectID
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	var opts []option.ClientOption
	if cfg.Push.FirebaseCredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.Push.FirebaseCredentialsPath))
	} else if cfg.Push.FirebaseCredentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.Push.FirebaseCredentialsJSON)))
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, opts...)
	if err != nil {
		return nil, "", err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, "", err
	}

	return client, projectID, nil
}

func logDebugPayload(projectID string, msg *messaging.Message) {
	payload := map[string]any{
		"project_id": projectID,
		"message": map[string]any{
			"token": msg.Token,
			"notification": map[string]any{
				"title": msg.Notification.Title,
				"body":  msg.Notification.Body,
			},
		},
	}

	if msg.Android != nil {
		android := map[string]any{}
		if msg.Android.Priority != "" {
			android["priority"] = msg.Android.Priority
		}
		if msg.Android.Notification != nil && msg.Android.Notification.ChannelID != "" {
			android["notification"] = map[string]any{
				"channel_id": msg.Android.Notification.ChannelID,
			}
		}
		if len(android) > 0 {
			payload["message"].(map[string]any)["android"] = android
		}
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("fcm debug payload marshal error: %v", err)
		return
	}

	log.Printf("fcm debug payload: %s", raw)
}
