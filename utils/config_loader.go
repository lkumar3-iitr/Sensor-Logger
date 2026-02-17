package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ─── Sensor-level configs ───────────────────────────────────────────────

type CameraConfig struct {
	Enabled    bool   `yaml:"enabled"`
	DevicePath string `yaml:"device_path"`
	Resolution struct {
		Width  int `yaml:"width"`
		Height int `yaml:"height"`
	} `yaml:"resolution"`
	FPS           int    `yaml:"fps"`
	Format        string `yaml:"format"`
	ChannelBuffer int    `yaml:"channel_buffer"`
	SaveFrames    bool   `yaml:"save_frames"`
}

type LidarConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Address       string `yaml:"address"`
	Port          int    `yaml:"port"`
	Model         string `yaml:"model"`
	RPM           int    `yaml:"rpm"`
	ChannelBuffer int    `yaml:"channel_buffer"`
	PointsPerPkt  int    `yaml:"points_per_packet"`
}

type GPSConfig struct {
	Enabled       bool   `yaml:"enabled"`
	SerialPort    string `yaml:"serial_port"`
	BaudRate      int    `yaml:"baud_rate"`
	UpdateRateHz  int    `yaml:"update_rate_hz"`
	ChannelBuffer int    `yaml:"channel_buffer"`
}

type IMUConfig struct {
	Enabled       bool   `yaml:"enabled"`
	SerialPort    string `yaml:"serial_port"`
	BaudRate      int    `yaml:"baud_rate"`
	UpdateRateHz  int    `yaml:"update_rate_hz"`
	ChannelBuffer int    `yaml:"channel_buffer"`
}

type RadarConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Address       string `yaml:"address"`
	Port          int    `yaml:"port"`
	ChannelBuffer int    `yaml:"channel_buffer"`
}

type SimulationConfig struct {
	Enabled         bool `yaml:"enabled"`
	DurationSeconds int  `yaml:"duration_seconds"`
}

// SensorsConfig is the top-level structure for sensors.yaml.
type SensorsConfig struct {
	Sensors struct {
		Camera CameraConfig `yaml:"camera"`
		Lidar  LidarConfig  `yaml:"lidar"`
		GPS    GPSConfig    `yaml:"gps"`
		IMU    IMUConfig    `yaml:"imu"`
		Radar  RadarConfig  `yaml:"radar"`
	} `yaml:"sensors"`
	Simulation SimulationConfig `yaml:"simulation"`
}

// ─── Storage configs ────────────────────────────────────────────────────

type CSVStorageConfig struct {
	FlushIntervalMs int  `yaml:"flush_interval_ms"`
	BufferSizeKB    int  `yaml:"buffer_size_kb"`
	WriteHeader     bool `yaml:"write_header"`
}

type FrameStorageConfig struct {
	SavePath string `yaml:"save_path"`
	Naming   string `yaml:"naming"` // "timestamp" or "sequence"
}

type StorageConfig struct {
	Storage struct {
		BaseDir       string             `yaml:"base_dir"`
		SessionPrefix string             `yaml:"session_prefix"`
		CSV           CSVStorageConfig   `yaml:"csv"`
		Frames        FrameStorageConfig `yaml:"frames"`
		Overwrite     bool               `yaml:"overwrite"`
	} `yaml:"storage"`
}

// ─── Loaders ────────────────────────────────────────────────────────────

// LoadSensorsConfig reads and parses sensors.yaml.
func LoadSensorsConfig(path string) (*SensorsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read sensors config: %w", err)
	}
	var cfg SensorsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse sensors config: %w", err)
	}
	return &cfg, nil
}

// LoadStorageConfig reads and parses storage.yaml.
func LoadStorageConfig(path string) (*StorageConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read storage config: %w", err)
	}
	var cfg StorageConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse storage config: %w", err)
	}
	return &cfg, nil
}
