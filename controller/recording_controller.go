package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"sensor-logger/models"
	"sensor-logger/utils"
	"sensor-logger/views"
)

// RecordingController is the final pipeline stage.
// It reads FusedRecords and writes them to:
//   - A fused CSV   (all sensor columns in one file)
//   - Per-sensor CSVs (camera.csv, lidar.csv, gps.csv, imu.csv, radar.csv)
//   - Camera JPEG frames to disk (optional)
//
// Writing is fully asynchronous with periodic flush, ensuring zero
// back-pressure on the fusion stage.
type RecordingController struct {
	storageCfg *utils.StorageConfig
	sessionDir string

	fusedWriter  *views.CSVWriter
	cameraWriter *views.CSVWriter
	lidarWriter  *views.CSVWriter
	gpsWriter    *views.CSVWriter
	imuWriter    *views.CSVWriter
	radarWriter  *views.CSVWriter

	saveFrames bool
	framesDir  string

	rowsWritten uint64
	wg          sync.WaitGroup
}

// NewRecordingController sets up the session directory tree and CSV writers.
func NewRecordingController(storageCfg *utils.StorageConfig, sensorsCfg *utils.SensorsConfig) (*RecordingController, error) {
	sess := utils.SessionName(storageCfg.Storage.SessionPrefix)
	sessionDir := filepath.Join(storageCfg.Storage.BaseDir, sess)

	if !storageCfg.Storage.Overwrite {
		if _, err := os.Stat(sessionDir); err == nil {
			return nil, fmt.Errorf("session dir %s already exists (overwrite=false)", sessionDir)
		}
	}

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}

	csvCfg := storageCfg.Storage.CSV
	bufSize := csvCfg.BufferSizeKB * 1024

	rc := &RecordingController{
		storageCfg: storageCfg,
		sessionDir: sessionDir,
	}

	// ── Fused CSV ────────────────────────────────────────────────────
	var err error
	rc.fusedWriter, err = views.NewCSVWriter(
		filepath.Join(sessionDir, "fused.csv"), bufSize, csvCfg.WriteHeader,
		models.FusedRecord{}.CSVHeader(),
	)
	if err != nil {
		return nil, err
	}

	// ── Per-sensor CSVs ──────────────────────────────────────────────
	if sensorsCfg.Sensors.Camera.Enabled {
		rc.cameraWriter, err = views.NewCSVWriter(
			filepath.Join(sessionDir, "camera.csv"), bufSize, csvCfg.WriteHeader,
			models.CameraFrame{}.CSVHeader(),
		)
		if err != nil {
			return nil, err
		}

		if sensorsCfg.Sensors.Camera.SaveFrames {
			rc.framesDir = filepath.Join(sessionDir, storageCfg.Storage.Frames.SavePath)
			if err := os.MkdirAll(rc.framesDir, 0755); err != nil {
				return nil, fmt.Errorf("create frames dir: %w", err)
			}
			rc.saveFrames = true
		}
	}

	if sensorsCfg.Sensors.Lidar.Enabled {
		rc.lidarWriter, err = views.NewCSVWriter(
			filepath.Join(sessionDir, "lidar.csv"), bufSize, csvCfg.WriteHeader,
			models.LidarPacket{}.CSVHeader(),
		)
		if err != nil {
			return nil, err
		}
	}

	if sensorsCfg.Sensors.GPS.Enabled {
		rc.gpsWriter, err = views.NewCSVWriter(
			filepath.Join(sessionDir, "gps.csv"), bufSize, csvCfg.WriteHeader,
			models.GPSData{}.CSVHeader(),
		)
		if err != nil {
			return nil, err
		}
	}

	if sensorsCfg.Sensors.IMU.Enabled {
		rc.imuWriter, err = views.NewCSVWriter(
			filepath.Join(sessionDir, "imu.csv"), bufSize, csvCfg.WriteHeader,
			models.IMUData{}.CSVHeader(),
		)
		if err != nil {
			return nil, err
		}
	}

	if sensorsCfg.Sensors.Radar.Enabled {
		rc.radarWriter, err = views.NewCSVWriter(
			filepath.Join(sessionDir, "radar.csv"), bufSize, csvCfg.WriteHeader,
			models.RadarTarget{}.CSVHeader(),
		)
		if err != nil {
			return nil, err
		}
	}

	utils.L().Info("recording controller ready  session=%s", sessionDir)
	return rc, nil
}

// Start begins consuming fused records and writing CSVs.
// It also starts a periodic flush goroutine.
func (rc *RecordingController) Start(ctx context.Context, fusedCh <-chan *models.FusedRecord) {
	// Periodic flusher
	rc.wg.Add(1)
	go func() {
		defer rc.wg.Done()
		flushMs := rc.storageCfg.Storage.CSV.FlushIntervalMs
		if flushMs <= 0 {
			flushMs = 100
		}
		ticker := time.NewTicker(time.Duration(flushMs) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				rc.flushAll()
				return
			case <-ticker.C:
				rc.flushAll()
			}
		}
	}()

	// Main writer goroutine
	rc.wg.Add(1)
	go func() {
		defer rc.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case rec, ok := <-fusedCh:
				if !ok {
					return
				}
				rc.writeRecord(rec)
			}
		}
	}()

	utils.L().Info("recording controller started")
}

// writeRecord fans out one FusedRecord to fused CSV + per-sensor CSVs + frame saves.
func (rc *RecordingController) writeRecord(rec *models.FusedRecord) {
	// Fused CSV row
	rc.fusedWriter.WriteRow(rec.CSVRow())

	// Per-sensor CSVs
	if rec.Camera != nil && rc.cameraWriter != nil {
		// Save JPEG frame if configured
		if rc.saveFrames && len(rec.Camera.JPEG) > 0 {
			fname := fmt.Sprintf("%d.jpg", rec.Camera.TimestampNs)
			fpath := filepath.Join(rc.framesDir, fname)
			rec.Camera.FilePath = filepath.Join(rc.storageCfg.Storage.Frames.SavePath, fname)
			go func(data []byte, path string) {
				if err := os.WriteFile(path, data, 0644); err != nil {
					utils.L().Error("save frame: %v", err)
				}
			}(rec.Camera.JPEG, fpath)
		}
		rc.cameraWriter.WriteRow(rec.Camera.CSVRow())
	}

	if rec.Lidar != nil && rc.lidarWriter != nil {
		rc.lidarWriter.WriteRow(rec.Lidar.CSVRow())
	}
	if rec.GPS != nil && rc.gpsWriter != nil {
		rc.gpsWriter.WriteRow(rec.GPS.CSVRow())
	}
	if rec.IMU != nil && rc.imuWriter != nil {
		rc.imuWriter.WriteRow(rec.IMU.CSVRow())
	}
	if rec.Radar != nil && rc.radarWriter != nil {
		rc.radarWriter.WriteRow(rec.Radar.CSVRow())
	}

	atomic.AddUint64(&rc.rowsWritten, 1)
}

func (rc *RecordingController) flushAll() {
	if rc.fusedWriter != nil {
		rc.fusedWriter.Flush()
	}
	if rc.cameraWriter != nil {
		rc.cameraWriter.Flush()
	}
	if rc.lidarWriter != nil {
		rc.lidarWriter.Flush()
	}
	if rc.gpsWriter != nil {
		rc.gpsWriter.Flush()
	}
	if rc.imuWriter != nil {
		rc.imuWriter.Flush()
	}
	if rc.radarWriter != nil {
		rc.radarWriter.Flush()
	}
}

// Stop waits for all writer goroutines, then flushes and closes every CSV.
func (rc *RecordingController) Stop() {
	rc.wg.Wait()
	rc.flushAll()

	for _, w := range []*views.CSVWriter{
		rc.fusedWriter, rc.cameraWriter, rc.lidarWriter,
		rc.gpsWriter, rc.imuWriter, rc.radarWriter,
	} {
		if w != nil {
			w.Close()
		}
	}

	rows := atomic.LoadUint64(&rc.rowsWritten)
	utils.L().Info("recording controller stopped  (rows_written=%d, session=%s)", rows, rc.sessionDir)
}

// SessionDir returns the path to the active session directory.
func (rc *RecordingController) SessionDir() string {
	return rc.sessionDir
}

// RowsWritten returns the total number of fused rows persisted.
func (rc *RecordingController) RowsWritten() uint64 {
	return atomic.LoadUint64(&rc.rowsWritten)
}
