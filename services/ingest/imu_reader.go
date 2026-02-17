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

// IMUReader ingests inertial measurement data from a serial port (or simulates).
type IMUReader struct {
	cfg      utils.IMUConfig
	sim      bool
	Out      chan *models.IMUData
	dropped  uint64
	produced uint64
}

func NewIMUReader(cfg utils.IMUConfig, simulate bool) *IMUReader {
	buf := cfg.ChannelBuffer
	if buf <= 0 {
		buf = 512
	}
	return &IMUReader{
		cfg: cfg,
		sim: simulate,
		Out: make(chan *models.IMUData, buf),
	}
}

func (r *IMUReader) Start(ctx context.Context) {
	go r.run(ctx)
	utils.L().Info("imu reader started     (rate=%dHz, buffer=%d, simulate=%v)",
		r.cfg.UpdateRateHz, r.cfg.ChannelBuffer, r.sim)
}

func (r *IMUReader) run(ctx context.Context) {
	defer close(r.Out)

	interval := time.Second / time.Duration(r.cfg.UpdateRateHz)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var step float64
	for {
		select {
		case <-ctx.Done():
			utils.L().Info("imu reader stopped     (produced=%d, dropped=%d)",
				atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped))
			return
		case <-ticker.C:
			d := r.read(step)
			step += 0.01

			select {
			case r.Out <- d:
				atomic.AddUint64(&r.produced, 1)
			default:
				atomic.AddUint64(&r.dropped, 1)
			}
		}
	}
}

func (r *IMUReader) read(step float64) *models.IMUData {
	ts := utils.NowNano()

	if r.sim {
		return &models.IMUData{
			TimestampNs: ts,
			AccelX:      0.02*math.Sin(step) + rand.Float64()*0.005,
			AccelY:      0.01*math.Cos(step) + rand.Float64()*0.005,
			AccelZ:      9.81 + rand.Float64()*0.02,
			GyroX:       0.001*math.Sin(step*2) + rand.Float64()*0.0005,
			GyroY:       0.001*math.Cos(step*2) + rand.Float64()*0.0005,
			GyroZ:       0.0005 + rand.Float64()*0.0002,
			MagX:        25.0 + rand.Float64()*0.5,
			MagY:        -10.0 + rand.Float64()*0.5,
			MagZ:        45.0 + rand.Float64()*0.5,
			Temperature: 35.0 + rand.Float64()*2.0,
		}
	}

	// TODO: real serial IMU read
	return &models.IMUData{TimestampNs: ts}
}

func (r *IMUReader) Stats() (uint64, uint64) {
	return atomic.LoadUint64(&r.produced), atomic.LoadUint64(&r.dropped)
}
