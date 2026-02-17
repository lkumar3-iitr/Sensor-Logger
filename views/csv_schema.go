package views

// CSVSchema defines the column layout for each sensor type's CSV output.
// This file serves as the single source of truth for column ordering.

// SensorType identifies a sensor for schema lookups.
type SensorType int

const (
	SensorCamera SensorType = iota
	SensorLidar
	SensorGPS
	SensorIMU
	SensorRadar
	SensorFused
)

var sensorNames = map[SensorType]string{
	SensorCamera: "camera",
	SensorLidar:  "lidar",
	SensorGPS:    "gps",
	SensorIMU:    "imu",
	SensorRadar:  "radar",
	SensorFused:  "fused",
}

func (s SensorType) String() string {
	if n, ok := sensorNames[s]; ok {
		return n
	}
	return "unknown"
}

// SchemaColumns returns the canonical column list for a sensor.
// The actual header writing is handled by the model's CSVHeader() method;
// this is kept here as a human-readable reference and for validation.
var SchemaColumns = map[SensorType][]string{
	SensorCamera: {
		"timestamp_ns", "frame_id", "width", "height",
		"format", "file_path", "size_bytes",
	},
	SensorLidar: {
		"timestamp_ns", "packet_id", "num_points", "model",
		"rotation_deg", "cloud_file_path", "size_bytes",
	},
	SensorGPS: {
		"timestamp_ns", "latitude", "longitude", "altitude",
		"speed", "heading", "hdop", "fix_quality", "num_sats",
	},
	SensorIMU: {
		"timestamp_ns",
		"accel_x", "accel_y", "accel_z",
		"gyro_x", "gyro_y", "gyro_z",
		"mag_x", "mag_y", "mag_z",
		"temperature",
	},
	SensorRadar: {
		"timestamp_ns", "target_id", "range", "azimuth",
		"elevation", "velocity", "rcs",
	},
	SensorFused: {
		"timestamp_ns",
		"cam_frame_id", "cam_file_path", "cam_width", "cam_height",
		"lidar_packet_id", "lidar_num_points", "lidar_cloud_path",
		"gps_lat", "gps_lon", "gps_alt", "gps_speed", "gps_heading",
		"imu_ax", "imu_ay", "imu_az", "imu_gx", "imu_gy", "imu_gz",
		"radar_range", "radar_azimuth", "radar_velocity",
	},
}
