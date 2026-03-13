package models

import "time"

// CameraInfo contains details about the connected camera device
type CameraInfo struct {
	SerialNumber     string
	FirmwareVersion  string
	IsConnected      bool
	DepthFOV         float64 // field of view in degrees
	IsIRFilterActive bool
}

// CameraFrame represents one synchronized snapshot from all camera sensors
type CameraFrame struct {
	Timestamp       time.Time // Time when your program received the frame
	DeviceTimestamp float64   // Time when the camera captured the frame
	FrameIndex      uint64

	RGB   RGBFrame
	Depth DepthFrame
	IR    IRFrame
	IMU   CameraIMUData
}

// ---------------- RGB DATA ----------------

// RGBFrame represents a color frame from the RGB camera
type RGBFrame struct {
	Data   []byte
	Width  int
	Height int
}

// ---------------- DEPTH DATA ----------------

// DepthFrame represents a depth image
type DepthFrame struct {
	Data   []uint16
	Width  int
	Height int
}

// ---------------- INFRARED DATA ----------------

// IRFrame represents stereo infrared images
type IRFrame struct {
	LeftData  []byte
	RightData []byte
	Width     int
	Height    int
}

// ---------------- IMU DATA ----------------

// CameraIMUData represents the 6 Degrees of Freedom data
type CameraIMUData struct {
	DeviceTimestamp float64    // Hardware timestamp to correlate Accel & Gyro packets
	Acceleration    [3]float64 // X, Y, Z linear acceleration
	AngularVelocity [3]float64 // X, Y, Z rotational velocity
}
