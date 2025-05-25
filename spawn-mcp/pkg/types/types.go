package types

import "time"

type DroneType string
type DroneStatus string
type DroneInfo struct {
	ID string
	Type DroneType
	Status DroneStatus
	CreatedAt time.Time
}