package controller

import (
	"context"
	"sync"
	"time"

	"sensor-logger/models"
	"sensor-logger/utils"
)

// FusionController merges all sensor streams into a single time-aligned
// fused record stream. It runs one goroutine per sensor channel plus a
// merge goroutine that emits FusedRecords at a fixed alignment cadence.
//
// Design goals:
//   - Zero back-pressure on sensor goroutines (non-blocking reads with latest-value semantics).
//   - Lock-free latest-value slots via atomic-style mutex (tiny critical section).
//   - Deterministic output rate decoupled from sensor rates.
type FusionController struct {
	mu sync.Mutex

	latestCamera *models.CameraFrame
	latestLidar  *models.LidarPacket
	latestGPS    *models.GPSData
	latestIMU    *models.IMUData
	latestRadar  *models.RadarTarget

	Out chan *models.FusedRecord // downstream consumers read this

	alignIntervalMs int
}

// NewFusionController creates a fusion stage.
// alignMs controls how often a fused snapshot is emitted (e.g. 33 ms ≈ 30 Hz).
func NewFusionController(alignMs int) *FusionController {
	if alignMs <= 0 {
		alignMs = 33 // default ~30 Hz
	}
	return &FusionController{
		Out:             make(chan *models.FusedRecord, 256),
		alignIntervalMs: alignMs,
	}
}

// Start launches drain goroutines for each sensor channel plus the merge ticker.
func (fc *FusionController) Start(ctx context.Context, sc *SensorsController) {
	// Drain each sensor channel into its latest-value slot.
	if sc.CameraCh != nil {
		go fc.drainCamera(ctx, sc.CameraCh)
	}
	if sc.LidarCh != nil {
		go fc.drainLidar(ctx, sc.LidarCh)
	}
	if sc.GPSCh != nil {
		go fc.drainGPS(ctx, sc.GPSCh)
	}
	if sc.IMUCh != nil {
		go fc.drainIMU(ctx, sc.IMUCh)
	}
	if sc.RadarCh != nil {
		go fc.drainRadar(ctx, sc.RadarCh)
	}

	go fc.merge(ctx)
	utils.L().Info("fusion controller started (align_interval=%dms)", fc.alignIntervalMs)
}

// ─── drain goroutines ───────────────────────────────────────────────────
// Each one reads as fast as the sensor produces, keeping only the newest value.

func (fc *FusionController) drainCamera(ctx context.Context, ch <-chan *models.CameraFrame) {
	for {
		select {
		case <-ctx.Done():
			return
		case f, ok := <-ch:
			if !ok {
				return
			}
			fc.mu.Lock()
			fc.latestCamera = f
			fc.mu.Unlock()
		}
	}
}

func (fc *FusionController) drainLidar(ctx context.Context, ch <-chan *models.LidarPacket) {
	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-ch:
			if !ok {
				return
			}
			fc.mu.Lock()
			fc.latestLidar = p
			fc.mu.Unlock()
		}
	}
}

func (fc *FusionController) drainGPS(ctx context.Context, ch <-chan *models.GPSData) {
	for {
		select {
		case <-ctx.Done():
			return
		case g, ok := <-ch:
			if !ok {
				return
			}
			fc.mu.Lock()
			fc.latestGPS = g
			fc.mu.Unlock()
		}
	}
}

func (fc *FusionController) drainIMU(ctx context.Context, ch <-chan *models.IMUData) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-ch:
			if !ok {
				return
			}
			fc.mu.Lock()
			fc.latestIMU = d
			fc.mu.Unlock()
		}
	}
}

func (fc *FusionController) drainRadar(ctx context.Context, ch <-chan *models.RadarTarget) {
	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-ch:
			if !ok {
				return
			}
			fc.mu.Lock()
			fc.latestRadar = r
			fc.mu.Unlock()
		}
	}
}

// ─── merge: snapshot latest values at a fixed cadence ───────────────────

func (fc *FusionController) merge(ctx context.Context) {
	defer close(fc.Out)

	ticker := time.NewTicker(time.Duration(fc.alignIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.L().Info("fusion controller stopped")
			return
		case <-ticker.C:
			fc.mu.Lock()
			rec := &models.FusedRecord{
				TimestampNs: utils.NowNano(),
				Camera:      fc.latestCamera,
				Lidar:       fc.latestLidar,
				GPS:         fc.latestGPS,
				IMU:         fc.latestIMU,
				Radar:       fc.latestRadar,
			}
			// Clear slots so we don't duplicate stale data.
			fc.latestCamera = nil
			fc.latestLidar = nil
			fc.latestGPS = nil
			fc.latestIMU = nil
			fc.latestRadar = nil
			fc.mu.Unlock()

			// Non-blocking send
			select {
			case fc.Out <- rec:
			default:
				utils.L().Warn("fusion: output channel full, dropping fused record")
			}
		}
	}
}
