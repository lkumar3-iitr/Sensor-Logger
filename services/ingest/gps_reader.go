package ingest

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"sensor-logger/models"
	"sensor-logger/utils"
)

// GPSReader ingests NMEA sentences from a serial port (or simulates fixes).
type GPSReader struct {
	cfg      utils.GPSConfig
	sim      bool
	Out      chan *models.GPSData
	dropped  uint64
	produced uint64
}

func NewGPSReader(cfg utils.GPSConfig, simulate bool) *GPSReader {
	buf := cfg.ChannelBuffer
	if buf <= 0 {
		buf = 64
	}
	return &GPSReader{
		cfg: cfg,
		sim: simulate,
		Out: make(chan *models.GPSData, buf),
	}
}

func (r *GPSReader) Start(ctx context.Context) {
	go r.run(ctx)
	utils.L().Info("gps reader started     (rate=%dHz, buffer=%d, simulate=%v)",
		r.cfg.UpdateRateHz, r.cfg.ChannelBuffer, r.sim)
}

func (r *GPSReader) run(ctx context.Context) {
	defer close(r.Out)

	interval := time.Second / time.Duration(r.cfg.UpdateRateHz)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Simulated starting point: roughly Bengaluru, India
	lat, lon := 12.9716, 77.5946

	for {
		select {
		case <-ctx.Done():
			utils.L().Info("gps reader stopped     (produced=%d, dropped=%d)",
				atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped))
			return
		case <-ticker.C:
			fix := r.readFix(&lat, &lon)
			select {
			case r.Out <- fix:
				atomic.AddUint64(&r.produced, 1)
			default:
				atomic.AddUint64(&r.dropped, 1)
			}
		}
	}
}

func (r *GPSReader) readFix(lat, lon *float64) *models.GPSData {
	ts := utils.NowNano()

	if r.sim {
		// Simulate a slow drive (~30 km/h heading north-east)
		*lat += 0.00001 + rand.Float64()*0.000005
		*lon += 0.00001 + rand.Float64()*0.000005

		return &models.GPSData{
			TimestampNs: ts,
			Latitude:    *lat,
			Longitude:   *lon,
			Altitude:    920.0 + rand.Float64()*2.0,
			Speed:       8.0 + rand.Float64()*2.0, // ~8-10 m/s
			Heading:     45.0 + rand.Float64()*5.0,
			HDOP:        0.8 + rand.Float64()*0.4,
			FixQuality:  1,
			NumSats:     12 + rand.Intn(4),
		}
	}

	// TODO: parse real NMEA from serial
	return &models.GPSData{TimestampNs: ts}
}

func (r *GPSReader) Stats() (uint64, uint64) {
	return atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped)
}
