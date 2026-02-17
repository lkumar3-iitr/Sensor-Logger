# Sensor-Logger

**High-throughput, multi-threaded sensor data logger for autonomous driving dataset generation.**

Built in Go with goroutines, buffered channels, and lock-free patterns to guarantee **zero-lag capture** from cameras, LiDAR, GPS, IMU, and radar — all fused and written to CSV in real time.

---

## Architecture

```
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│  Camera   │  │  LiDAR   │  │   GPS    │  │   IMU    │  │  Radar   │
│ goroutine │  │ goroutine│  │ goroutine│  │ goroutine│  │ goroutine│
└────┬──────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘
     │ ch           │ ch          │ ch          │ ch          │ ch
     ▼              ▼             ▼             ▼             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Fusion Controller                                │
│  • Drains each channel into latest-value slot (non-blocking)        │
│  • Snapshots all sensors at fixed cadence (default 30 Hz)           │
│  • Emits FusedRecord to output channel                              │
└──────────────────────────────┬──────────────────────────────────────┘
                               │ FusedRecord ch
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Recording Controller                              │
│  • Writes fused.csv  (all sensors in one row)                       │
│  • Writes per-sensor CSVs (camera.csv, lidar.csv, gps.csv, …)      │
│  • Saves camera JPEG frames to disk (optional)                      │
│  • Periodic flush (configurable interval) — never blocks hot path   │
└─────────────────────────────────────────────────────────────────────┘
```

### Zero-Lag Guarantees

| Technique | Where |
|---|---|
| **Buffered channels** (120–512 deep) | Every sensor reader |
| **Non-blocking sends** with frame drop | Reader → channel push |
| **Latest-value slots** (tiny mutex) | Fusion drains |
| **Buffered `bufio.Writer`** (256 KB) | CSV writers |
| **Periodic flush** (100 ms default) | Recording controller |
| **Fire-and-forget** frame saves | JPEG writes in separate goroutines |

---

## Project Structure

```
Sensor-Logger/
├── cmd/
│   └── main.go                  # Entry point & pipeline wiring
├── config/
│   ├── sensors.yaml             # Sensor device & simulation settings
│   └── storage.yaml             # Output directory & CSV tuning
├── controller/
│   ├── fusion_controller.go     # Time-aligns all sensor streams
│   ├── recording_controller.go  # CSV + frame persistence
│   └── sensors_controller.go    # Lifecycle manager for all readers
├── models/
│   ├── camera_frame.go          # Camera data model + CSV serialisation
│   ├── fused_record.go          # Unified multi-sensor record
│   ├── gps_data.go              # GPS fix model
│   ├── helpers.go               # Shared formatting (itoa, ftoa …)
│   ├── imu_data.go              # IMU 9-axis model
│   ├── lidar_packet.go          # LiDAR point-cloud metadata
│   └── radar_target.go          # Radar target model
├── services/
│   └── ingest/
│       ├── camera_reader.go     # Camera capture goroutine
│       ├── gps_reader.go        # GPS NMEA reader goroutine
│       ├── imu_reader.go        # IMU serial reader goroutine
│       ├── lidar_reader.go      # LiDAR UDP reader goroutine
│       └── radar_reader.go      # Radar reader goroutine
├── utils/
│   ├── config_loader.go         # YAML config parser
│   ├── logger.go                # Thread-safe levelled logger
│   └── time_stamp.go            # Nanosecond timestamp utilities
├── views/
│   ├── csv_schema.go            # Column definitions per sensor
│   └── data_export.go           # High-perf buffered CSV writer
├── go.mod
└── README.md
```

---

## Quick Start

```bash
# Build
cd Sensor-Logger
go build -o sensor-logger ./cmd/

# Run with simulated sensors (default)
./sensor-logger

# Run with custom configs
./sensor-logger -sensors config/sensors.yaml -storage config/storage.yaml

# Run for exactly 10 seconds (override config)
# Set simulation.duration_seconds: 10 in sensors.yaml

# Press Ctrl+C to stop recording at any time
```

### CLI Flags

| Flag | Default | Description |
|---|---|---|
| `-sensors` | `config/sensors.yaml` | Path to sensor configuration |
| `-storage` | `config/storage.yaml` | Path to storage configuration |
| `-log` | *(none)* | Optional log file path |
| `-align-ms` | `33` | Fusion alignment interval in ms (~30 Hz) |

---

## Output

A session directory is created under `data/` (configurable):

```
data/drive_20260214_153045/
├── fused.csv          # All sensors merged, one row per alignment tick
├── camera.csv         # Per-frame camera metadata
├── lidar.csv          # Per-packet LiDAR metadata
├── gps.csv            # Per-fix GPS data
├── imu.csv            # Per-sample IMU readings
├── radar.csv          # Per-target radar detections
└── frames/            # Camera JPEG frames (if save_frames: true)
    ├── 1739530225000000000.jpg
    ├── 1739530225033333333.jpg
    └── …
```

### Sample `fused.csv`

```csv
timestamp_ns,cam_frame_id,cam_file_path,cam_width,cam_height,lidar_packet_id,lidar_num_points,lidar_cloud_path,gps_lat,gps_lon,gps_alt,gps_speed,gps_heading,imu_ax,imu_ay,imu_az,imu_gx,imu_gy,imu_gz,radar_range,radar_azimuth,radar_velocity
1739530225033000000,42,frames/1739530225033000000.jpg,1920,1080,317,384,,12.971612345,77.594623456,920.123,8.4321,45.23,0.019876,0.010234,9.812345,0.001023,0.000987,0.000512,45.678,12.34,-5.432
```

---

## Configuration

### sensors.yaml

- **Enable/disable** each sensor independently
- **Channel buffer sizes** — tune per sensor for zero-drop operation
- **Simulation mode** — generates realistic synthetic data without hardware
- **Device addresses** — serial ports, network IPs for real hardware

### storage.yaml

- **Base directory** and session naming
- **CSV flush interval** — trade durability vs throughput (100 ms default)
- **Buffer size** — larger = fewer syscalls = faster (256 KB default)
- **Frame saving** — toggle JPEG persistence, choose naming scheme

---

## Concurrency Model

```
Main goroutine
  │
  ├── Camera reader goroutine        (ticker @ 30 fps → buffered chan)
  ├── LiDAR reader goroutine         (ticker @ ~750 pps → buffered chan)
  ├── GPS reader goroutine           (ticker @ 10 Hz → buffered chan)
  ├── IMU reader goroutine           (ticker @ 100 Hz → buffered chan)
  ├── Radar reader goroutine         (ticker @ 20 Hz → buffered chan)
  │
  ├── Fusion drain: camera           (reads chan → latest-value slot)
  ├── Fusion drain: lidar            (reads chan → latest-value slot)
  ├── Fusion drain: gps              (reads chan → latest-value slot)
  ├── Fusion drain: imu              (reads chan → latest-value slot)
  ├── Fusion drain: radar            (reads chan → latest-value slot)
  ├── Fusion merge goroutine         (ticker @ 30 Hz → snapshot → FusedRecord chan)
  │
  ├── Recording writer goroutine     (reads FusedRecord → CSV rows)
  ├── Recording flusher goroutine    (ticker @ 10 Hz → flush all buffers)
  └── Frame saver goroutines         (fire-and-forget JPEG writes)
```

**Total goroutines: ~14+** all coordinated via `context.Context` for graceful shutdown.

---

## Extending for Real Hardware

Each reader has a `TODO` stub for real device integration:

- **Camera**: Integrate V4L2 / GStreamer / OpenCV CGo bindings
- **LiDAR**: UDP socket reader for Velodyne VLP-16 / Ouster packets
- **GPS**: NMEA 0183 serial parser (GGA, RMC sentences)
- **IMU**: Serial protocol for BNO055 / MPU-9250
- **Radar**: CAN bus or Ethernet interface for Continental / Delphi

The simulation mode (`simulation.enabled: true`) lets you develop and test the full pipeline without any hardware.

---

## License

MIT