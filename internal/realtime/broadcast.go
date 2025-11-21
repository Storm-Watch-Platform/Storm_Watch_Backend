package realtime

type Broadcaster interface {
	BroadcastPresence(userID string, lat, lon float64, status string, displayUntil string)
	BroadcastSOS(sosID string, lat, lon float64, radiusM int, expiresAt string, body string)
}
