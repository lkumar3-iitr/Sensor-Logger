package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sensor_logger/models"
	"sensor_logger/services/ingest"
)

// Each LiDAR connected to the system would have its own controller instance.
// LidarController is an MVC Controller that manages a single Unitree L2 LiDAR,
// its commands, the service processing logic, and coordinates data.
type LidarController struct {
	deviceID string
	// The LiDAR service containing data processing logic
	service *ingest.LidarService
	// A send-only channel carrying LidarScan objects to the fusion pipeline
	dataChannel chan<- models.LidarScan
}

// NewLidarController injects the dependencies so it can route its scans
// out to the global data collection pipeline.
func NewLidarController(id string, channel chan<- models.LidarScan) *LidarController {
	lService := ingest.NewLidarService(id)
	return &LidarController{
		deviceID:    id,
		service:     lService,
		dataChannel: channel,
	}
}

// StartLidar is a device command route that starts the scan loop.
// This function could act as a router payload handler where clients
// call standard commands from a gRPC/HTTP endpoint for instance.
func (l *LidarController) StartLidar(ctx context.Context, waitGroup *sync.WaitGroup) error {
	defer waitGroup.Done()

	// 1. Establish the connection to the underlying device
	err := l.service.ConnectService()
	if err != nil {
		fmt.Printf("[LidarController] Failed to connect: %v\n", err)
		return err
	}

	// Retrieve properties to verify device is functional
	info, _ := l.service.GetInfoService()
	fmt.Printf("[LidarController] Hardware Started. FOV: %.0f°x%.0f°, Mode: %s, Connected: %v\n",
		info.HorizontalFOV, info.VerticalFOV, info.WorkingMode, info.IsConnected)

	// 2. Main operational loop — L2 scans at 5.55 Hz (one revolution every ~180ms)
	fmt.Printf("[LidarController] Entering continuous scan loop for %s...\n", l.deviceID)
	for {
		select {
		case <-ctx.Done():
			// The caller sent a stop signal, trigger shutdown commands.
			fmt.Printf("[LidarController] Halt command received.\n")
			l.service.DisconnectService()
			return nil
		default:
			// Fetch the assembled full-revolution scan from the service
			scan, err := l.service.ReadScanService()
			if err != nil {
				// Retry or reconnect logic here rather than crashing
				time.Sleep(1 * time.Second)
				continue
			}

			// We dynamically collect data into the global synchronizing channel.
			// The channel's listener (fusion_controller.go) will figure out
			// which time frame window instance in RealTimeData to inject it into.
			l.dataChannel <- scan
		}
	}
}

// StopLidar provides an external command trigger to disconnect and free up
// hardware pipeline resources for this LiDAR instance outside of the scan loop.
func (l *LidarController) StopLidar() error {
	return l.service.DisconnectService()
}
