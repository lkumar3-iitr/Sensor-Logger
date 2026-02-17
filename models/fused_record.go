package models

// FusedRecord is a time-aligned snapshot across all active sensors.
// The fusion controller writes one of these per alignment window.
type FusedRecord struct {
	TimestampNs int64        `json:"timestamp_ns"` // alignment anchor time
	Camera      *CameraFrame `json:"camera,omitempty"`
	Lidar       *LidarPacket `json:"lidar,omitempty"`
	GPS         *GPSData     `json:"gps,omitempty"`
	IMU         *IMUData     `json:"imu,omitempty"`
	Radar       *RadarTarget `json:"radar,omitempty"`
}

// CSVHeader returns the fused CSV header: common timestamp + each sensor block.
func (FusedRecord) CSVHeader() []string {
	h := []string{"timestamp_ns"}
	// Camera cols (prefixed)
	h = append(h, "cam_frame_id", "cam_file_path", "cam_width", "cam_height")
	// LiDAR cols
	h = append(h, "lidar_packet_id", "lidar_num_points", "lidar_cloud_path")
	// GPS cols
	h = append(h, "gps_lat", "gps_lon", "gps_alt", "gps_speed", "gps_heading")
	// IMU cols
	h = append(h, "imu_ax", "imu_ay", "imu_az", "imu_gx", "imu_gy", "imu_gz")
	// Radar cols
	h = append(h, "radar_range", "radar_azimuth", "radar_velocity")
	return h
}

// CSVRow returns a single fused row, using empty strings for missing sensors.
func (f *FusedRecord) CSVRow() []string {
	row := []string{itoa64(f.TimestampNs)}

	// Camera
	if f.Camera != nil {
		row = append(row, utoa64(f.Camera.FrameID), f.Camera.FilePath,
			itoa(f.Camera.Width), itoa(f.Camera.Height))
	} else {
		row = append(row, "", "", "", "")
	}

	// LiDAR
	if f.Lidar != nil {
		row = append(row, utoa64(f.Lidar.PacketID), itoa(f.Lidar.NumPoints),
			f.Lidar.CloudFilePath)
	} else {
		row = append(row, "", "", "")
	}

	// GPS
	if f.GPS != nil {
		row = append(row, ftoa(f.GPS.Latitude, 9), ftoa(f.GPS.Longitude, 9),
			ftoa(f.GPS.Altitude, 3), ftoa(f.GPS.Speed, 4), ftoa(f.GPS.Heading, 2))
	} else {
		row = append(row, "", "", "", "", "")
	}

	// IMU
	if f.IMU != nil {
		row = append(row,
			ftoa(f.IMU.AccelX, 6), ftoa(f.IMU.AccelY, 6), ftoa(f.IMU.AccelZ, 6),
			ftoa(f.IMU.GyroX, 6), ftoa(f.IMU.GyroY, 6), ftoa(f.IMU.GyroZ, 6))
	} else {
		row = append(row, "", "", "", "", "", "")
	}

	// Radar
	if f.Radar != nil {
		row = append(row, ftoa(f.Radar.Range, 3), ftoa(f.Radar.Azimuth, 2),
			ftoa(f.Radar.Velocity, 3))
	} else {
		row = append(row, "", "", "")
	}

	return row
}
