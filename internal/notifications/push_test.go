package notifications

import "testing"

func TestBuildFirebaseMessageUsesMinimalAndroidPayload(t *testing.T) {
	msg := buildFirebaseMessage("device-token", PushMessage{
		Title: "Daily expense reminder",
		Body:  "Jangan lupa input pengeluaran hari ini.",
		Data: map[string]string{
			"kind":  "daily_expense_input",
			"type":  "daily_expense_input",
			"route": "/activity",
		},
	})

	if msg.Token != "device-token" {
		t.Fatalf("unexpected token: %s", msg.Token)
	}
	if msg.Notification == nil {
		t.Fatal("expected notification payload")
	}
	if msg.Notification.Title != "Daily expense reminder" {
		t.Fatalf("unexpected title: %s", msg.Notification.Title)
	}
	if msg.Notification.Body != "Jangan lupa input pengeluaran hari ini." {
		t.Fatalf("unexpected body: %s", msg.Notification.Body)
	}
	if msg.Data != nil {
		t.Fatalf("expected no data payload, got: %#v", msg.Data)
	}
	if msg.Android == nil {
		t.Fatal("expected android config")
	}
	if msg.Android.Priority != "high" {
		t.Fatalf("unexpected android priority: %s", msg.Android.Priority)
	}
	if msg.Android.Notification == nil {
		t.Fatal("expected android notification config")
	}
	if msg.Android.Notification.ChannelID != androidNotificationChannelID {
		t.Fatalf("unexpected channel id: %s", msg.Android.Notification.ChannelID)
	}
	if msg.Android.Notification.Sound != "" {
		t.Fatalf("expected empty sound for minimal payload, got %q", msg.Android.Notification.Sound)
	}
}
