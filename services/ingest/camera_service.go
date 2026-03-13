package ingest

import (
	"fmt"
	"math/rand"
	"time"

	"sensor_logger/models"
)

// CameraService handles the underlying logic to read from and communicate with
// the Intel RealSense camera (referred to here generically as Camera)
type CameraService struct {
	deviceID    string
	isConnected bool

	// When using real hardware, this would hold the RealSense pipeline handle.
	// Example (with CGO + librealsense):
	// pipeline *C.rs2_pipeline
	// The pipeline is responsible for:
	// - starting camera streams
	// - synchronizing RGB / Depth / IR / IMU
	// - providing frames to the application
}

// NewCameraService initializes a new service for a given camera ID
func NewCameraService(deviceID string) *CameraService {
	return &CameraService{
		deviceID:    deviceID,
		isConnected: false,
	}
}

// ConnectService contains the logic to open the pipeline and establish
// connection with the camera hardware.
func (s *CameraService) ConnectService() error {

	// ---------------- MOCK IMPLEMENTATION ----------------
	s.isConnected = true
	fmt.Printf("[CameraService] Connected to device %s via CGO binding...\n", s.deviceID)
	return nil

	// ---------------- REAL HARDWARE IMPLEMENTATION ----------------
	/*
		// Create RealSense pipeline
		pipeline := C.rs2_create_pipeline(nil)

		// Configure streams (RGB, Depth, IR, IMU)
		config := C.rs2_create_config()

		// Example configuration:
		C.rs2_config_enable_stream(config, C.RS2_STREAM_COLOR, 0, 1280, 800, C.RS2_FORMAT_RGB8, 30)
		C.rs2_config_enable_stream(config, C.RS2_STREAM_DEPTH, 0, 1280, 720, C.RS2_FORMAT_Z16, 30)
		C.rs2_config_enable_stream(config, C.RS2_STREAM_INFRARED, 1, 1280, 800, C.RS2_FORMAT_Y8, 30)
		C.rs2_config_enable_stream(config, C.RS2_STREAM_INFRARED, 2, 1280, 800, C.RS2_FORMAT_Y8, 30)

		// Start the pipeline
		C.rs2_pipeline_start_with_config(pipeline, config, nil)

		s.pipeline = pipeline
		s.isConnected = true
	*/
}

// DisconnectService contains logic to gracefully halt the pipeline and streams.
func (s *CameraService) DisconnectService() error {

	// ---------------- MOCK IMPLEMENTATION ----------------
	s.isConnected = false
	fmt.Printf("[CameraService] Disconnected from device %s\n", s.deviceID)
	return nil

	// ---------------- REAL HARDWARE IMPLEMENTATION ----------------
	/*
		if s.pipeline != nil {
			C.rs2_pipeline_stop(s.pipeline)
			C.rs2_delete_pipeline(s.pipeline)
		}

		s.isConnected = false
	*/
}

// ReadFrameService processes the camera hardware to construct a populated CameraFrame model.
// This uses global time for synchronization across devices.
func (s *CameraService) ReadFrameService() (models.CameraFrame, error) {

	if !s.isConnected {
		return models.CameraFrame{}, fmt.Errorf("[CameraService] device %s not connected", s.deviceID)
	}

	// ---------------- MOCK IMPLEMENTATION ----------------

	// Simulate frame arrival (≈100 FPS)
	time.Sleep(10 * time.Millisecond)

	frame := models.CameraFrame{
		Timestamp:       time.Now().UTC(),
		DeviceTimestamp: float64(time.Now().UnixNano()) / 1e6,
		FrameIndex:      uint64(rand.Int63()),

		RGB: models.RGBFrame{
			Width:  1280,
			Height: 800,
			Data:   make([]byte, 1280*800*3),
		},
		Depth: models.DepthFrame{
			Width:  1280,
			Height: 720,
			Data:   make([]uint16, 1280*720),
		},
		IR: models.IRFrame{
			Width:     1280,
			Height:    800,
			LeftData:  make([]byte, 1280*800),
			RightData: make([]byte, 1280*800),
		},
		IMU: models.CameraIMUData{
			DeviceTimestamp: float64(time.Now().UnixNano()) / 1e6,
			Acceleration:    [3]float64{0.0, 9.8, 0.0},
			AngularVelocity: [3]float64{0.1, 0.0, 0.0},
		},
	}

	return frame, nil

	// ---------------- REAL HARDWARE IMPLEMENTATION ----------------
	/*
		// Wait for next frame set from the camera
		frames := C.rs2_pipeline_wait_for_frames(s.pipeline, nil)

		// Extract color frame
		colorFrame := C.rs2_extract_frame(frames, C.RS2_STREAM_COLOR)

		// Extract depth frame
		depthFrame := C.rs2_extract_frame(frames, C.RS2_STREAM_DEPTH)

		// Extract infrared frames
		irLeft := C.rs2_extract_frame(frames, C.RS2_STREAM_INFRARED, 1)
		irRight := C.rs2_extract_frame(frames, C.RS2_STREAM_INFRARED, 2)

		// Extract IMU frame
		motionFrame := C.rs2_extract_motion_frame(frames)

		// Convert frame buffers to Go slices
		rgbData := C.rs2_get_frame_data(colorFrame)
		depthData := C.rs2_get_frame_data(depthFrame)

		// Populate CameraFrame struct with real data
	*/
}

// GetInfoService retrieves detailed info from the sensor
func (s *CameraService) GetInfoService() (models.CameraInfo, error) {

	// ---------------- MOCK IMPLEMENTATION ----------------
	return models.CameraInfo{
		SerialNumber:     "SN12345",
		FirmwareVersion:  "v5.0.0",
		IsConnected:      s.isConnected,
		DepthFOV:         90.0,
		IsIRFilterActive: true,
	}, nil

	// ---------------- REAL HARDWARE IMPLEMENTATION ----------------
	/*
		// Get device from the running pipeline
		device := C.rs2_pipeline_get_device(s.pipeline)

		// Query serial number and firmware version
		serial := C.rs2_get_device_info(device, C.RS2_CAMERA_INFO_SERIAL_NUMBER)
		firmware := C.rs2_get_device_info(device, C.RS2_CAMERA_INFO_FIRMWARE_VERSION)

		// --------------------------------------------------
		// Determine if IR emitter is active
		// --------------------------------------------------
		depthSensor := C.rs2_query_sensors(device)

		emitterEnabled := C.rs2_get_option(depthSensor, C.RS2_OPTION_EMITTER_ENABLED)

		isIRActive := false
		if emitterEnabled > 0 {
			isIRActive = true
		}

		// --------------------------------------------------
		// Retrieve depth camera intrinsics to compute FOV
		// --------------------------------------------------

		// Get depth stream profile
		profile := C.rs2_pipeline_get_active_profile(s.pipeline)

		depthStream := C.rs2_get_stream_profile(profile, C.RS2_STREAM_DEPTH)

		// Get intrinsics
		var intrinsics C.rs2_intrinsics
		C.rs2_get_video_stream_intrinsics(depthStream, &intrinsics)

		width := float64(intrinsics.width)
		fx := float64(intrinsics.fx)

		// Compute horizontal field of view
		depthFOV := 2.0 * math.Atan(width/(2.0*fx)) * (180.0 / math.Pi)

		return models.CameraInfo{
			SerialNumber:     C.GoString(serial),
			FirmwareVersion:  C.GoString(firmware),
			IsConnected:      s.isConnected,
			DepthFOV:         depthFOV,
			IsIRFilterActive: isIRActive,
		}, nil
	*/
}

//------------------------------------NOTES------------------------------------------

// s *CameraService -> means 's' is a pointer to a CameraService object.
// struct in Go don't have methods directly, so methods are attached using receiver functions.
// Each CameraService object typically represents one physical camera.
// The functions in camera_service.go provide abstraction over the RealSense pipeline and hardware APIs.
