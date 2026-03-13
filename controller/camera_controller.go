package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sensor_logger/models"
	"sensor_logger/services/ingest"
)

// Each camera connected to the system would have its own controller instance
// CameraController is an MVC Controller that manages a single camera,
// its commands, the service processing logic, and coordinates data.
type CameraController struct {
	deviceID string
	// The camera service containing data processing logic
	service *ingest.CameraService
	// Global list of time windows used for dynamically tracking data among devices
	// By pushing to them over a channel or via a manager
	// chan is a Go keyword that represents a channel type.

	//a send-only channel carrying CameraFrame objects
	dataChannel chan<- models.CameraFrame
}

// NewCameraController injects the dependencies so it can route its frames
// out to the global data collection pipeline.
func NewCameraController(id string, channel chan<- models.CameraFrame) *CameraController {
	cService := ingest.NewCameraService(id)
	return &CameraController{
		deviceID:    id,
		service:     cService,
		dataChannel: channel,
	}
}

// StartCamera is a device command route that starts streaming frames.
// This function could act as a router payload handler where clients
// call standard commands from a gRPC/HTTP endpoint for instance.
func (c *CameraController) StartCamera(ctx context.Context, waitGroup *sync.WaitGroup) error {
	defer waitGroup.Done()

	// 1. Establish the connection to the underlying device
	err := c.service.ConnectService()
	if err != nil {
		fmt.Printf("[CameraController] Failed to connect: %v\n", err)
		return err
	}

	// Retrieve properties to verify device is functional
	info, _ := c.service.GetInfoService()
	fmt.Printf("[CameraController] Hardware Started. FOV: %.2f, Connected: %v\n", info.DepthFOV, info.IsConnected)

	// 2. Main operational loop using Goroutines for high frequency streaming
	fmt.Printf("[CameraController] Entering continuous capture loop for %s...\n", c.deviceID)
	for {
		select {
		case <-ctx.Done():
			// The caller sent a stop signal, trigger shutdown commands.
			fmt.Printf("[CameraController] Halt command received.\n")
			c.service.DisconnectService()
			return nil
		default:
			// Fetch the data block processed via CGO over the service
			frame, err := c.service.ReadFrameService()
			if err != nil {
				// Retry or reconnect logic here rather than crashing
				time.Sleep(1 * time.Second)
				continue
			}

			// We dynamically collect data into the global synchronizing channel.
			// The channel's listener (like `fusion_controller.go`) will figure out
			// which time frame window instance in `RealTimeData.go` to inject it into.
			c.dataChannel <- frame
		}
	}
}

// StopCamera provides an external command trigger to disconnect and free up
// hardware pipeline resources for this camera instances outside of the capture loop.
func (c *CameraController) StopCamera() error {
	return c.service.DisconnectService()
}
