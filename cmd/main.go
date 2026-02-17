package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"sensor-logger/controller"
	"sensor-logger/utils"
)

func main() {
	// ── CLI flags ────────────────────────────────────────────────────
	sensorsPath := flag.String("sensors", "config/sensors.yaml", "path to sensors.yaml")
	storagePath := flag.String("storage", "config/storage.yaml", "path to storage.yaml")
	logFile := flag.String("log", "", "optional log file path (stdout is always included)")
	alignMs := flag.Int("align-ms", 33, "fusion alignment interval in milliseconds (~30 Hz)")
	flag.Parse()

	// ── Logger ───────────────────────────────────────────────────────
	logger := utils.InitLogger(utils.INFO, *logFile)
	defer logger.Close()

	utils.L().Info("═══════════════════════════════════════════════════")
	utils.L().Info("  Sensor-Logger  ·  Autonomous Driving Dataset Gen")
	utils.L().Info("  GOMAXPROCS=%d  ·  PID=%d", runtime.GOMAXPROCS(0), os.Getpid())
	utils.L().Info("═══════════════════════════════════════════════════")

	// ── Load configs ─────────────────────────────────────────────────
	sensorsCfg, err := utils.LoadSensorsConfig(*sensorsPath)
	if err != nil {
		utils.L().Fatal("load sensors config: %v", err)
	}
	storageCfg, err := utils.LoadStorageConfig(*storagePath)
	if err != nil {
		utils.L().Fatal("load storage config: %v", err)
	}

	// Resolve relative base_dir to absolute.
	if !filepath.IsAbs(storageCfg.Storage.BaseDir) {
		abs, _ := filepath.Abs(storageCfg.Storage.BaseDir)
		storageCfg.Storage.BaseDir = abs
	}

	// ── Context with OS signal cancellation ──────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Optional fixed duration from config.
	duration := sensorsCfg.Simulation.DurationSeconds
	if duration > 0 {
		var timerCancel context.CancelFunc
		ctx, timerCancel = context.WithTimeout(ctx, time.Duration(duration)*time.Second)
		defer timerCancel()
		utils.L().Info("recording will auto-stop after %ds", duration)
	}

	// ── Pipeline assembly ────────────────────────────────────────────
	//
	//  Sensor goroutines  ──►  buffered channels  ──►  FusionController
	//                                                         │
	//                                                   FusedRecord chan
	//                                                         │
	//                                                  RecordingController
	//                                                   │           │
	//                                              fused.csv    per-sensor CSVs + frames

	// 1. Sensors
	sensorCtrl := controller.NewSensorsController(sensorsCfg)
	sensorCtrl.Start(ctx)

	// 2. Fusion
	fusionCtrl := controller.NewFusionController(*alignMs)
	fusionCtrl.Start(ctx, sensorCtrl)

	// 3. Recording
	recordCtrl, err := controller.NewRecordingController(storageCfg, sensorsCfg)
	if err != nil {
		utils.L().Fatal("init recording controller: %v", err)
	}
	recordCtrl.Start(ctx, fusionCtrl.Out)

	utils.L().Info("pipeline running — press Ctrl+C to stop")

	// ── Stats ticker ─────────────────────────────────────────────────
	statsTicker := time.NewTicker(5 * time.Second)
	defer statsTicker.Stop()

	// ── Main event loop ──────────────────────────────────────────────
	for {
		select {
		case sig := <-sigCh:
			utils.L().Info("received signal: %v — shutting down…", sig)
			cancel()
			goto shutdown

		case <-ctx.Done():
			goto shutdown

		case <-statsTicker.C:
			utils.L().Info("── stats ─────────────────────────")
			sensorCtrl.LogStats()
			utils.L().Info("  fused rows written: %d", recordCtrl.RowsWritten())
			utils.L().Info("──────────────────────────────────")
		}
	}

shutdown:
	// Allow a brief drain period for in-flight data.
	utils.L().Info("draining pipeline…")
	time.Sleep(500 * time.Millisecond)

	recordCtrl.Stop()

	utils.L().Info("session saved to: %s", recordCtrl.SessionDir())
	utils.L().Info("total fused rows: %d", recordCtrl.RowsWritten())

	fmt.Println("\n✓ Sensor-Logger finished. Dataset at:", recordCtrl.SessionDir())
}
