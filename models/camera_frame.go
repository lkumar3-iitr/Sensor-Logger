package models

// CameraFrame holds a single captured frame with its metadata.
// The raw JPEG bytes travel through the channel; CSV only stores the metadata row.
type CameraFrame struct {
	TimestampNs int64  `json:"timestamp_ns"` // nanosecond-precision capture time
	FrameID     uint64 `json:"frame_id"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Format      string `json:"format"`    // MJPEG, RAW, PNG …
	FilePath    string `json:"file_path"` // relative path where the frame JPEG was saved
	SizeBytes   int    `json:"size_bytes"`
	JPEG        []byte `json:"-"` // raw image data – NOT written to CSV
}

// CSVHeader returns the ordered column names for the camera CSV.
func (CameraFrame) CSVHeader() []string {
	return []string{
		"timestamp_ns", "frame_id", "width", "height",
		"format", "file_path", "size_bytes",
	}
}

// CSVRow serialises one frame into a CSV-compatible string slice.
func (f *CameraFrame) CSVRow() []string {
	return []string{
		itoa64(f.TimestampNs),
		utoa64(f.FrameID),
		itoa(f.Width),
		itoa(f.Height),
		f.Format,
		f.FilePath,
		itoa(f.SizeBytes),
	}
}
