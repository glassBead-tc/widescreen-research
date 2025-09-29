// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package types

import "time"

type DroneType string
type DroneStatus string
type DroneInfo struct {
	ID        string
	Type      DroneType
	Status    DroneStatus
	CreatedAt time.Time
}
