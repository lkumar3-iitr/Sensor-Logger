package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sensor_logger/controller"
)

func main() {
	fmt.Println("Starting Sensor Logger Pipeline...")

	// 1. Setup a global context to manage graceful shutdowns across the go routines
	// A context is used to signal all goroutines to stop.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// A WaitGroup ensures the program waits until all goroutines finish before exiting.
	var wg sync.WaitGroup

	// Set window timeframe to a discrete size, e.g., 5 seconds.
	// We dynamically grow data inside this window and write per 5-seconds logic.
	timeFrameWindow := 5 * time.Second

	// 2. Initialize the Central Fusion controller
	fusionController := controller.NewDataFusionController(timeFrameWindow)

	// 3. Initialize the Camera controller, injecting the channel from the Fusion logic
	// The camera service handles commands inside StartCamera
	cameraD456 := controller.NewCameraController("Camera_D456", fusionController.GetCameraChannel())

	// ============================================ Run Routines ============================================

	// Start the data multiplexing logic inside the Fusion Controller
	wg.Add(1)
	go fusionController.StartIngestListeners(ctx, &wg)

	// Start the cyclic window rotator for SSD persisting
	wg.Add(1)
	go fusionController.StartWindowRotator(ctx, &wg)

	// Start reading physical sensor bytes over C wrapper service inside the Camera Controller
	wg.Add(1)
	go func() {
		err := cameraD456.StartCamera(ctx, &wg)
		if err != nil {
			fmt.Printf("Camera exited loop with err: %v\n", err)
			cancel() // Gracefully halt the other routines if hardware completely fails
		}
	}()

	// 4. Await interrupts to shut everything down cleanly
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\n[Main] Logging started. Press CTRL+C to perform SSD flush and exit cleanly.")
	<-sigChan

	fmt.Println("\n[Main] Shutdown signal detected! Attempting to close goroutines and store final arrays.")
	cameraD456.StopCamera()
	cancel()  // Release the context to break out of controller loops
	wg.Wait() // Wait for all channels and SSD writes to finish

	fmt.Println("All goroutines cleanly exited. Logs successfully flushed.")
}
