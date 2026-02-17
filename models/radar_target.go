package models

// RadarTarget represents a single detected radar target.
type RadarTarget struct {
	TimestampNs int64   `json:"timestamp_ns"`
	TargetID    int     `json:"target_id"`
	Range       float64 `json:"range"`     // metres
	Azimuth     float64 `json:"azimuth"`   // degrees
	Elevation   float64 `json:"elevation"` // degrees
	Velocity    float64 `json:"velocity"`  // m/s  (radial, positive = approaching)
	RCS         float64 `json:"rcs"`       // radar cross-section dBsm
}

func (RadarTarget) CSVHeader() []string {
	return []string{
		"timestamp_ns", "target_id", "range", "azimuth",
		"elevation", "velocity", "rcs",
	}
}

func (r *RadarTarget) CSVRow() []string {
	return []string{
		itoa64(r.TimestampNs),
		itoa(r.TargetID),
		ftoa(r.Range, 3),
		ftoa(r.Azimuth, 2),
		ftoa(r.Elevation, 2),
		ftoa(r.Velocity, 3),
		ftoa(r.RCS, 2),
	}
}
