package subscriptions

import (
	"notification_system/platforms"
	"notification_system/record"
)

type Subscription struct {
	Platform  platforms.NotificationPlatform
	AlertType record.AlertType
	Devices   []string
}
