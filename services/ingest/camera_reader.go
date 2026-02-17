package ingest

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"sensor-logger/models"
	"sensor-logger/utils"
)

// CameraReader ingests frames from a camera device (or simulates them).
// It pumps into a buffered channel that downstream consumers read without blocking.
type CameraReader struct {
	cfg      utils.CameraConfig
	sim      bool
	Out      chan *models.CameraFrame
	dropped  uint64
	produced uint64
}

// NewCameraReader wires up the camera ingest pipeline.
func NewCameraReader(cfg utils.CameraConfig, simulate bool) *CameraReader {
	buf := cfg.ChannelBuffer
	if buf <= 0 {
		buf = 120
	}
	return &CameraReader{
		cfg: cfg,
		sim: simulate,
		Out: make(chan *models.CameraFrame, buf),
	}
}

// Start launches a goroutine that reads/simulates camera frames until ctx is cancelled.
func (r *CameraReader) Start(ctx context.Context) {
	go r.run(ctx)
	utils.L().Info("camera reader started  (fps=%d, buffer=%d, simulate=%v)",
		r.cfg.FPS, r.cfg.ChannelBuffer, r.sim)
}

func (r *CameraReader) run(ctx context.Context) {
	defer close(r.Out)

	interval := time.Second / time.Duration(r.cfg.FPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var seq uint64
	for {
		select {
		case <-ctx.Done():
			utils.L().Info("camera reader stopped  (produced=%d, dropped=%d)",
				atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped))
			return
		case <-ticker.C:
			frame := r.capture(seq)
			seq++

			// Non-blocking send: if the channel is full we drop the frame
			// to guarantee zero back-pressure on the capture goroutine.
			select {
			case r.Out <- frame:
				atomic.AddUint64(&r.produced, 1)
			default:
				atomic.AddUint64(&r.dropped, 1)
				utils.L().Warn("camera: dropped frame %d (consumer too slow)", frame.FrameID)
			}
		}
	}
}

// capture either reads a real device or generates synthetic data.
func (r *CameraReader) capture(seq uint64) *models.CameraFrame {
	ts := utils.NowNano()

	if r.sim {
		// Simulate a JPEG frame (random bytes, realistic size).
		sz := 80_000 + rand.Intn(40_000) // 80-120 KB
		jpeg := make([]byte, sz)
		// fill first few bytes so it looks like a JPEG SOI marker
		jpeg[0], jpeg[1] = 0xFF, 0xD8
		return &models.CameraFrame{
			TimestampNs: ts,
			FrameID:     seq,
			Width:       r.cfg.Resolution.Width,
			Height:      r.cfg.Resolution.Height,
			Format:      r.cfg.Format,
			SizeBytes:   sz,
			JPEG:        jpeg,
		}
	}

	// ── Real device capture (stub) ──────────────────────────────────
	// TODO: integrate V4L2 / GStreamer pipeline for actual frame capture.
	// For now, return an empty frame placeholder.
	return &models.CameraFrame{
		TimestampNs: ts,
		FrameID:     seq,
		Width:       r.cfg.Resolution.Width,
		Height:      r.cfg.Resolution.Height,
		Format:      r.cfg.Format,
	}
}

// Stats returns (produced, dropped) counts atomically.
func (r *CameraReader) Stats() (uint64, uint64) {
	return atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped)
}
