package controller

import (
	"context"

	"sensor-logger/models"
	"sensor-logger/services/ingest"
	"sensor-logger/utils"
)

// SensorsController owns the lifecycle of every sensor reader goroutine.
// It exposes typed output channels that downstream controllers consume.
type SensorsController struct {
	camera *ingest.CameraReader
	lidar  *ingest.LidarReader
	gps    *ingest.GPSReader
	imu    *ingest.IMUReader
	radar  *ingest.RadarReader

	CameraCh chan *models.CameraFrame
	LidarCh  chan *models.LidarPacket
	GPSCh    chan *models.GPSData
	IMUCh    chan *models.IMUData
	RadarCh  chan *models.RadarTarget
}

// NewSensorsController creates reader instances for every enabled sensor.
func NewSensorsController(cfg *utils.SensorsConfig) *SensorsController {
	sc := &SensorsController{}
	sim := cfg.Simulation.Enabled

	if cfg.Sensors.Camera.Enabled {
		sc.camera = ingest.NewCameraReader(cfg.Sensors.Camera, sim)
		sc.CameraCh = sc.camera.Out
	}
	if cfg.Sensors.Lidar.Enabled {
		sc.lidar = ingest.NewLidarReader(cfg.Sensors.Lidar, sim)
		sc.LidarCh = sc.lidar.Out
	}
	if cfg.Sensors.GPS.Enabled {
		sc.gps = ingest.NewGPSReader(cfg.Sensors.GPS, sim)
		sc.GPSCh = sc.gps.Out
	}
	if cfg.Sensors.IMU.Enabled {
		sc.imu = ingest.NewIMUReader(cfg.Sensors.IMU, sim)
		sc.IMUCh = sc.imu.Out
	}
	if cfg.Sensors.Radar.Enabled {
		sc.radar = ingest.NewRadarReader(cfg.Sensors.Radar, sim)
		sc.RadarCh = sc.radar.Out
	}

	return sc
}

// Start launches all enabled sensor goroutines.
func (sc *SensorsController) Start(ctx context.Context) {
	if sc.camera != nil {
		sc.camera.Start(ctx)
	}
	if sc.lidar != nil {
		sc.lidar.Start(ctx)
	}
	if sc.gps != nil {
		sc.gps.Start(ctx)
	}
	if sc.imu != nil {
		sc.imu.Start(ctx)
	}
	if sc.radar != nil {
		sc.radar.Start(ctx)
	}
	utils.L().Info("sensors controller: all enabled readers launched")
}

// LogStats prints current produce/drop counters for each active sensor.
func (sc *SensorsController) LogStats() {
	if sc.camera != nil {
		p, d := sc.camera.Stats()
		utils.L().Info("  camera   produced=%d  dropped=%d", p, d)
	}
	if sc.lidar != nil {
		p, d := sc.lidar.Stats()
		utils.L().Info("  lidar    produced=%d  dropped=%d", p, d)
	}
	if sc.gps != nil {
		p, d := sc.gps.Stats()
		utils.L().Info("  gps      produced=%d  dropped=%d", p, d)
	}
	if sc.imu != nil {
		p, d := sc.imu.Stats()
		utils.L().Info("  imu      produced=%d  dropped=%d", p, d)
	}
	if sc.radar != nil {
		p, d := sc.radar.Stats()
		utils.L().Info("  radar    produced=%d  dropped=%d", p, d)
	}
}
