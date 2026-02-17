package models

// IMUData holds one inertial measurement unit reading.
type IMUData struct {
	TimestampNs int64   `json:"timestamp_ns"`
	AccelX      float64 `json:"accel_x"` // m/s²
	AccelY      float64 `json:"accel_y"`
	AccelZ      float64 `json:"accel_z"`
	GyroX       float64 `json:"gyro_x"` // rad/s
	GyroY       float64 `json:"gyro_y"`
	GyroZ       float64 `json:"gyro_z"`
	MagX        float64 `json:"mag_x"` // µT (micro-tesla)
	MagY        float64 `json:"mag_y"`
	MagZ        float64 `json:"mag_z"`
	Temperature float64 `json:"temperature"` // °C
}

func (IMUData) CSVHeader() []string {
	return []string{
		"timestamp_ns",
		"accel_x", "accel_y", "accel_z",
		"gyro_x", "gyro_y", "gyro_z",
		"mag_x", "mag_y", "mag_z",
		"temperature",
	}
}

func (d *IMUData) CSVRow() []string {
	return []string{
		itoa64(d.TimestampNs),
		ftoa(d.AccelX, 6), ftoa(d.AccelY, 6), ftoa(d.AccelZ, 6),
		ftoa(d.GyroX, 6), ftoa(d.GyroY, 6), ftoa(d.GyroZ, 6),
		ftoa(d.MagX, 4), ftoa(d.MagY, 4), ftoa(d.MagZ, 4),
		ftoa(d.Temperature, 2),
	}
}
