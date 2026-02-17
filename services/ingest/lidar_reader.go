package ingest

import (
	"context"
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	"sensor-logger/models"
	"sensor-logger/utils"
)

// LidarReader ingests LiDAR packets from a network socket (or simulates them).
type LidarReader struct {
	cfg      utils.LidarConfig
	sim      bool
	Out      chan *models.LidarPacket
	dropped  uint64
	produced uint64
}

func NewLidarReader(cfg utils.LidarConfig, simulate bool) *LidarReader {
	buf := cfg.ChannelBuffer
	if buf <= 0 {
		buf = 256
	}
	return &LidarReader{
		cfg: cfg,
		sim: simulate,
		Out: make(chan *models.LidarPacket, buf),
	}
}

func (r *LidarReader) Start(ctx context.Context) {
	go r.run(ctx)
	utils.L().Info("lidar reader started   (model=%s, buffer=%d, simulate=%v)",
		r.cfg.Model, r.cfg.ChannelBuffer, r.sim)
}

func (r *LidarReader) run(ctx context.Context) {
	defer close(r.Out)

	// Velodyne VLP-16 at 600 RPM ≈ 754 packets/sec
	packetsPerSec := float64(r.cfg.RPM) / 60.0 * 75.0 // ~75 packets per rotation
	interval := time.Duration(float64(time.Second) / packetsPerSec)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var seq uint64
	for {
		select {
		case <-ctx.Done():
			utils.L().Info("lidar reader stopped   (produced=%d, dropped=%d)",
				atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped))
			return
		case <-ticker.C:
			pkt := r.readPacket(seq)
			seq++

			select {
			case r.Out <- pkt:
				atomic.AddUint64(&r.produced, 1)
			default:
				atomic.AddUint64(&r.dropped, 1)
			}
		}
	}
}

func (r *LidarReader) readPacket(seq uint64) *models.LidarPacket {
	ts := utils.NowNano()

	if r.sim {
		nPts := r.cfg.PointsPerPkt + rand.Intn(32) - 16
		cloud := make([]byte, nPts*16) // 16 bytes per point (x,y,z,intensity as float32)
		rot := math.Mod(float64(seq)*0.48, 360.0)
		return &models.LidarPacket{
			TimestampNs: ts,
			PacketID:    seq,
			NumPoints:   nPts,
			Model:       r.cfg.Model,
			RotationDeg: rot,
			SizeBytes:   len(cloud),
			RawCloud:    cloud,
		}
	}

	// TODO: real UDP socket read from Velodyne/Ouster
	return &models.LidarPacket{
		TimestampNs: ts,
		PacketID:    seq,
		Model:       r.cfg.Model,
	}
}

func (r *LidarReader) Stats() (uint64, uint64) {
	return atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped)
}
