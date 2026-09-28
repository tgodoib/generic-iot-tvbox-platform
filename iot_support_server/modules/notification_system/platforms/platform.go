package platforms

type NotificationPlatform interface {
	Send(title string, body string, deviceId string) error
}
