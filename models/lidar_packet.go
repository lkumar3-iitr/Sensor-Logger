package models

import "time"

// ================== DEVICE INFO ==================

// LidarInfo contains details about the connected Unitree 4D LiDAR-L2 device.
type LidarInfo struct {
	SerialNumber    string
	FirmwareVersion string
	IsConnected     bool

	// Network configuration (ENET UDP default)
	DeviceIP   string // Default: 192.168.1.62
	ServerIP   string // Default destination: 192.168.1.2
	SendPort   int    // L2 sends from port 6101
	ListenPort int    // Server receives on port 6201

	// Capabilities from spec sheet
	WorkingMode    string  // "3D" or "2D"
	NegaMode       bool    // If true, vertical FOV expands to 96°
	IMUEnabled     bool    // IMU reporting enable/disable
	HorizontalFOV  float64 // 360 degrees
	VerticalFOV    float64 // 90° normal, 96° NEGA mode
	SamplingFreqHz int     // 64000 effective points/sec
	ScanFreqHz     float64 // 5.55 Hz circumferential rotation
	MinRange       float64 // 0.05 m
	MaxRange       float64 // 30 m (at 90% reflectivity)
}

// ================== POINT CLOUD DATA ==================

// LidarPoint represents a single point from the L2 scan.
// The L2 outputs 4D data: 3D position + 1D grayscale (reflectivity).
type LidarPoint struct {
	// Cartesian coordinates derived from distance + angles
	// Origin is at the bottom center of the L2 unit (see coordinate system in manual)
	X float64 // mm, +X = opposite direction of the outlet
	Y float64 // mm, +Y = 90° counterclockwise from +X
	Z float64 // mm, +Z = upward

	Distance       float64 // Raw distance value in mm
	HorizontalAngle float64 // Azimuth angle in degrees (0-360)
	VerticalAngle  float64 // Elevation angle in degrees
	Reflectivity   uint8   // Grayscale reflectivity of the detected surface
}

// ================== IMU DATA ==================

// LidarIMUData represents the built-in IMU data from the L2.
// The IMU samples at 1 kHz internally and reports at 500 Hz.
// Axes are parallel to the point cloud coordinate system O-XYZ.
type LidarIMUData struct {
	DeviceTimestamp float64    // Hardware timestamp for correlation
	Acceleration    [3]float64 // X, Y, Z linear acceleration (m/s²)
	AngularVelocity [3]float64 // X, Y, Z rotational velocity (rad/s)
}

// ================== DEVICE STATUS ==================

// LidarStatus reports the operating state of the L2 during a scan.
type LidarStatus struct {
	WorkingState string  // "Sampling", "Standby", or "Interference"
	RPM          float64 // Current rotational speed
	Voltage      float64 // Current supply voltage (nominal 12V DC)
	Temperature  float64 // Device temperature in °C (-10 to 50 operating range)
}

// ================== SCAN FRAME ==================

// LidarScan represents one complete 360° revolution of point cloud data
// collected by the L2 at its 5.55 Hz scan frequency.
// Each scan contains approximately 11,532 points (64000 / 5.55).
type LidarScan struct {
	Timestamp       time.Time // Global time when this scan was received
	DeviceTimestamp float64   // Hardware timestamp from the L2
	ScanIndex       uint64    // Monotonically increasing scan counter

	Points []LidarPoint // Point cloud for this revolution
	IMU    LidarIMUData // Latest IMU reading during this scan
	Status LidarStatus  // Device operating status
}
