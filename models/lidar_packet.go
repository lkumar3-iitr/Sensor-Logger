package models

// LidarPacket holds one rotation/packet of 3-D point cloud data.
// Only metadata is written to CSV; the raw cloud is stored separately.
type LidarPacket struct {
	TimestampNs   int64   `json:"timestamp_ns"`
	PacketID      uint64  `json:"packet_id"`
	NumPoints     int     `json:"num_points"`
	Model         string  `json:"model"`           // VLP-16, OS1-64, etc.
	RotationDeg   float64 `json:"rotation_deg"`    // azimuth at packet start
	CloudFilePath string  `json:"cloud_file_path"` // path to .pcd / .bin
	SizeBytes     int     `json:"size_bytes"`
	RawCloud      []byte  `json:"-"` // raw binary point cloud – not in CSV
}

func (LidarPacket) CSVHeader() []string {
	return []string{
		"timestamp_ns", "packet_id", "num_points", "model",
		"rotation_deg", "cloud_file_path", "size_bytes",
	}
}

func (l *LidarPacket) CSVRow() []string {
	return []string{
		itoa64(l.TimestampNs),
		utoa64(l.PacketID),
		itoa(l.NumPoints),
		l.Model,
		ftoa(l.RotationDeg, 2),
		l.CloudFilePath,
		itoa(l.SizeBytes),
	}
}
