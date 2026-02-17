package ingest

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"sensor-logger/models"
	"sensor-logger/utils"
)

// RadarReader ingests radar targets from a network interface (or simulates).
type RadarReader struct {
	cfg      utils.RadarConfig
	sim      bool
	Out      chan *models.RadarTarget
	dropped  uint64
	produced uint64
}

func NewRadarReader(cfg utils.RadarConfig, simulate bool) *RadarReader {
	buf := cfg.ChannelBuffer
	if buf <= 0 {
		buf = 128
	}
	return &RadarReader{
		cfg: cfg,
		sim: simulate,
		Out: make(chan *models.RadarTarget, buf),
	}
}

func (r *RadarReader) Start(ctx context.Context) {
	go r.run(ctx)
	utils.L().Info("radar reader started   (buffer=%d, simulate=%v)",
		r.cfg.ChannelBuffer, r.sim)
}

func (r *RadarReader) run(ctx context.Context) {
	defer close(r.Out)

	// Radar typically reports at ~20 Hz
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var id int
	for {
		select {
		case <-ctx.Done():
			utils.L().Info("radar reader stopped   (produced=%d, dropped=%d)",
				atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped))
			return
		case <-ticker.C:
			t := r.readTarget(id)
			id++
			select {
			case r.Out <- t:
				atomic.AddUint64(&r.produced, 1)
			default:
				atomic.AddUint64(&r.dropped, 1)
			}
		}
	}
}

func (r *RadarReader) readTarget(id int) *models.RadarTarget {
	ts := utils.NowNano()

	if r.sim {
		return &models.RadarTarget{
			TimestampNs: ts,
			TargetID:    id,
			Range:       10.0 + rand.Float64()*90.0,
			Azimuth:     -30.0 + rand.Float64()*60.0,
			Elevation:   -5.0 + rand.Float64()*10.0,
			Velocity:    -15.0 + rand.Float64()*30.0,
			RCS:         -10.0 + rand.Float64()*30.0,
		}
	}

	// TODO: real radar network read
	return &models.RadarTarget{TimestampNs: ts, TargetID: id}
}

func (r *RadarReader) Stats() (uint64, uint64) {
	return atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped)
}
