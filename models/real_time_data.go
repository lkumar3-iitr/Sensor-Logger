package models

import (
	"fmt"
	"sync"
	"time"
)

//In Go, a 'slice' is a dynamic, flexible view of an array.

// RealTimeData acts as a container for all sensor data collected within a specific time window.
// Once the time window has elapsed, the instance is written to SSD and flushed (cleared).
type RealTimeData struct {
	StartTime time.Time
	EndTime   time.Time

	// We use a mutex to ensure safe concurrent access when multiple controllers
	// (Go routines) push their data to this shared instance.
	mutex sync.Mutex

	// Slices to hold data dynamically. They grow as new data comes in during this time window.
	CameraFrames []CameraFrame
	// GPS data, LiDAR data etc. would be added here similarly.
}

// NewRealTimeData creates a new window instance covering the specified duration.
func NewRealTimeData(start time.Time, duration time.Duration) *RealTimeData {
	return &RealTimeData{
		StartTime:    start,
		EndTime:      start.Add(duration),
		CameraFrames: make([]CameraFrame, 0), // Starts empty, grows dynamically based on number of calls
	}
}

// IsInWindow checks if a given timestamp falls within this RealTimeData's time window.
func (rtd *RealTimeData) IsInWindow(timestamp time.Time) bool {
	return (timestamp.Equal(rtd.StartTime) || timestamp.After(rtd.StartTime)) && timestamp.Before(rtd.EndTime)
}

// AddCameraFrame pushes a new camera frame into this time window's storage array.
func (rtd *RealTimeData) AddCameraFrame(frame CameraFrame) {
	rtd.mutex.Lock()
	defer rtd.mutex.Unlock()

	rtd.CameraFrames = append(rtd.CameraFrames, frame)
}

// FlushToSSD is a mock function that simulates writing the window's data
// to persistent storage (like an SSD) and then clearing the data from memory.
func (rtd *RealTimeData) FlushToSSD() error {
	rtd.mutex.Lock()
	defer rtd.mutex.Unlock()

	// TODO: Replace with actual file writing logic (e.g., CSV, binary format, or Database)
	fmt.Printf("[RealTimeData] Writing window [%s - %s] to SSD. Total Camera Frames: %d\n",
		rtd.StartTime.Format(time.RFC3339),
		rtd.EndTime.Format(time.RFC3339),
		len(rtd.CameraFrames))

	// Clear the slices down to size 0. Go's slice re-slicing keeps the underlying array capacity
	// but changes the length, freeing up the logical space for GC or reuse.
	rtd.CameraFrames = nil

	return nil
}
