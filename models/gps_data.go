package models

// GPSData holds one NMEA-derived GPS fix.
type GPSData struct {
	TimestampNs int64   `json:"timestamp_ns"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Altitude    float64 `json:"altitude"`    // metres above WGS-84 ellipsoid
	Speed       float64 `json:"speed"`       // m/s
	Heading     float64 `json:"heading"`     // degrees from true north
	HDOP        float64 `json:"hdop"`        // horizontal dilution of precision
	FixQuality  int     `json:"fix_quality"` // 0=invalid, 1=GPS, 2=DGPS, 4=RTK …
	NumSats     int     `json:"num_sats"`
}

func (GPSData) CSVHeader() []string {
	return []string{
		"timestamp_ns", "latitude", "longitude", "altitude",
		"speed", "heading", "hdop", "fix_quality", "num_sats",
	}
}

func (g *GPSData) CSVRow() []string {
	return []string{
		itoa64(g.TimestampNs),
		ftoa(g.Latitude, 9),
		ftoa(g.Longitude, 9),
		ftoa(g.Altitude, 3),
		ftoa(g.Speed, 4),
		ftoa(g.Heading, 2),
		ftoa(g.HDOP, 2),
		itoa(g.FixQuality),
		itoa(g.NumSats),
	}
}
