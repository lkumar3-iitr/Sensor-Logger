package ingest

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"sensor_logger/models"
)

// LidarService handles the underlying logic to read from and communicate with
// the Unitree 4D LiDAR-L2 via ENET UDP (default) or TTL UART.
type LidarService struct {
	deviceID    string
	isConnected bool

	// Network configuration for ENET UDP communication
	deviceIP   string // L2 default: 192.168.1.62
	serverIP   string // Default destination: 192.168.1.2
	sendPort   int    // L2 sends on port 6101
	listenPort int    // Server receives on port 6201

	scanCounter uint64 // Monotonically increasing scan index

	// When using real hardware, this would hold the UDP connection handle.
	// Example (with Unilidar SDK):
	// conn *net.UDPConn
}

// NewLidarService initializes a new service for a given LiDAR device ID
// with the L2 default network configuration.
func NewLidarService(deviceID string) *LidarService {
	return &LidarService{
		deviceID:    deviceID,
		isConnected: false,
		deviceIP:    "192.168.1.62",
		serverIP:    "192.168.1.2",
		sendPort:    6101,
		listenPort:  6201,
		scanCounter: 0,
	}
}

// ConnectService contains the logic to open a UDP socket and establish
// connection with the L2 LiDAR hardware.
func (s *LidarService) ConnectService() error {

	// ---------------- MOCK IMPLEMENTATION ----------------
	s.isConnected = true
	fmt.Printf("[LidarService] Connected to device %s at %s:%d (UDP)...\n",
		s.deviceID, s.deviceIP, s.listenPort)
	return nil

	// ---------------- REAL HARDWARE IMPLEMENTATION ----------------
	/*
		// Bind to the server's listening UDP port to receive L2 data
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", s.listenPort))
		if err != nil {
			return fmt.Errorf("failed to resolve UDP address: %w", err)
		}

		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			return fmt.Errorf("failed to bind UDP port %d: %w", s.listenPort, err)
		}

		s.conn = conn
		s.isConnected = true

		// The L2 defaults to power-on self-start, so it should already be
		// streaming data to serverIP:listenPort after power on.
		// If CMD Start mode is configured, send a start command via SDK here.
	*/
}

// DisconnectService contains logic to gracefully close the UDP connection
// and optionally send a standby command to the L2.
func (s *LidarService) DisconnectService() error {

	// ---------------- MOCK IMPLEMENTATION ----------------
	s.isConnected = false
	fmt.Printf("[LidarService] Disconnected from device %s\n", s.deviceID)
	return nil

	// ---------------- REAL HARDWARE IMPLEMENTATION ----------------
	/*
		if s.conn != nil {
			// Optionally send standby mode command before closing
			// In standby: power < 1W, LED off, motors stop, IMU disabled
			s.SetWorkingMode("Standby")

			s.conn.Close()
			s.conn = nil
		}

		s.isConnected = false
	*/
}

// ReadScanService processes the LiDAR UDP packets to construct a populated LidarScan model.
// Each scan represents one 360° revolution at 5.55 Hz (~11,532 points per revolution).
// This uses global time for synchronization across devices.
func (s *LidarService) ReadScanService() (models.LidarScan, error) {

	if !s.isConnected {
		return models.LidarScan{}, fmt.Errorf("[LidarService] device %s not connected", s.deviceID)
	}

	// ---------------- MOCK IMPLEMENTATION ----------------

	// Simulate one full 360° scan at 5.55 Hz rotation speed
	time.Sleep(180 * time.Millisecond)

	// Generate ~11,532 simulated points (64000 effective points / 5.55 Hz)
	numPoints := 11532
	points := make([]models.LidarPoint, numPoints)

	for i := 0; i < numPoints; i++ {
		hAngle := float64(i) / float64(numPoints) * 360.0 // Spread across 360°
		vAngle := rand.Float64() * 90.0                    // 0-90° vertical FOV

		// Simulate a random distance between 0.05m and 30m (in mm)
		distance := 50.0 + rand.Float64()*29950.0 // 50mm to 30000mm

		// Convert spherical to Cartesian (L2 coordinate system)
		hRad := hAngle * math.Pi / 180.0
		vRad := vAngle * math.Pi / 180.0

		points[i] = models.LidarPoint{
			X:               distance * math.Cos(vRad) * math.Cos(hRad),
			Y:               distance * math.Cos(vRad) * math.Sin(hRad),
			Z:               distance * math.Sin(vRad),
			Distance:        distance,
			HorizontalAngle: hAngle,
			VerticalAngle:   vAngle,
			Reflectivity:    uint8(rand.Intn(256)),
		}
	}

	s.scanCounter++

	scan := models.LidarScan{
		Timestamp:       time.Now().UTC(),
		DeviceTimestamp: float64(time.Now().UnixNano()) / 1e6,
		ScanIndex:       s.scanCounter,
		Points:          points,
		IMU: models.LidarIMUData{
			DeviceTimestamp: float64(time.Now().UnixNano()) / 1e6,
			Acceleration:    [3]float64{0.0, 0.0, 9.81},
			AngularVelocity: [3]float64{0.0, 0.0, 0.0},
		},
		Status: models.LidarStatus{
			WorkingState: "Sampling",
			RPM:          333.0, // 5.55 Hz * 60
			Voltage:      12.0,
			Temperature:  25.0,
		},
	}

	return scan, nil

	// ---------------- REAL HARDWARE IMPLEMENTATION ----------------
	/*
		// Read UDP packets from the L2 until one full revolution is assembled.
		// The L2 sends data on UDP port 6101 → server port 6201.
		//
		// Using the Unilidar SDK:
		//   1. Read raw UDP packet buffer
		//   2. Parse point cloud data (distance, angles, reflectivity)
		//   3. Parse embedded IMU data (if IMU enabled)
		//   4. Parse working status (RPM, voltage, temperature)
		//   5. Convert to Cartesian using L2 coordinate system definition
		//
		// buf := make([]byte, 65535)
		// n, _, err := s.conn.ReadFromUDP(buf)
		// if err != nil {
		//     return models.LidarScan{}, fmt.Errorf("UDP read error: %w", err)
		// }
		// scan := parseUnilidarPacket(buf[:n])
	*/
}

// GetInfoService retrieves detailed info from the LiDAR sensor.
func (s *LidarService) GetInfoService() (models.LidarInfo, error) {

	// ---------------- MOCK IMPLEMENTATION ----------------
	return models.LidarInfo{
		SerialNumber:    "UL2-2024-001",
		FirmwareVersion: "v1.1",
		IsConnected:     s.isConnected,
		DeviceIP:        s.deviceIP,
		ServerIP:        s.serverIP,
		SendPort:        s.sendPort,
		ListenPort:      s.listenPort,
		WorkingMode:     "3D",
		NegaMode:        false,
		IMUEnabled:      true,
		HorizontalFOV:   360.0,
		VerticalFOV:     90.0,
		SamplingFreqHz:  64000,
		ScanFreqHz:      5.55,
		MinRange:        0.05,
		MaxRange:        30.0,
	}, nil

	// ---------------- REAL HARDWARE IMPLEMENTATION ----------------
	/*
		// Query device info using Unilidar SDK
		// The SDK provides functions to query firmware version, serial number,
		// current working mode, and network configuration.
		//
		// info := sdk.GetDeviceInfo(s.conn)
		// return models.LidarInfo{...}, nil
	*/
}

// SetWorkingMode configures the L2 working mode.
// Supported modes: "Normal", "Standby", "3D", "2D", "NEGA"
// Changes to 3D/2D and NEGA mode require saving parameters and restarting the radar.
func (s *LidarService) SetWorkingMode(mode string) error {

	// ---------------- MOCK IMPLEMENTATION ----------------
	fmt.Printf("[LidarService] Working mode set to: %s\n", mode)
	return nil

	// ---------------- REAL HARDWARE IMPLEMENTATION ----------------
	/*
		// Send mode configuration command via Unilidar SDK
		// For mode changes like 3D/2D or NEGA, the SDK will:
		//   1. Send the configuration command
		//   2. Save parameters
		//   3. Require radar restart to take effect
		//
		// err := sdk.SetMode(s.conn, mode)
		// if err != nil {
		//     return fmt.Errorf("failed to set mode: %w", err)
		// }
	*/
}

// ------------------------------------ NOTES ------------------------------------------

// s *LidarService -> means 's' is a pointer to a LidarService object.
// Each LidarService object typically represents one physical L2 LiDAR unit.
// The L2 communicates via ENET UDP by default:
//   - Device IP: 192.168.1.62, Gateway: 192.168.1.1, Subnet Mask: 255.255.255.0
//   - Sends data from port 6101, destination server receives on port 6201
//   - Ensure the server IP (default 192.168.1.2) does not conflict with the L2 IP
// Alternative TTL UART communication is available via the GH1.25-4Y serial port
//   at 4,000,000 bps baud rate using the provided adapter module.
