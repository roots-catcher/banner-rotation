package events

import "time"

type EventType string

const (
	EventShow  EventType = "show"
	EventClick EventType = "click"
)

type BannerEvent struct {
	Type      EventType `json:"type"`
	SlotID    int       `json:"slot_id"`
	BannerID  int       `json:"banner_id"`
	GroupID   int       `json:"group_id"`
	Timestamp time.Time `json:"timestamp"`
}
