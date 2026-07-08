package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Server represents an MCP server that exposes tools.
// In the first version, this is a local in-process server.
// Future versions will support the standard MCP protocol over stdio/HTTP.
type Server struct {
	config   MCPConfig
	tools    map[string]ToolHandler
	registry *Registry
}

// ToolHandler is a function that handles an MCP tool call.
type ToolHandler func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

// NewServer creates a new MCP server instance.
func NewServer(config MCPConfig, registry *Registry) *Server {
	return &Server{
		config:   config,
		tools:    make(map[string]ToolHandler),
		registry: registry,
	}
}

// RegisterHandler registers a tool handler for this server.
func (s *Server) RegisterHandler(toolName string, handler ToolHandler) {
	s.tools[toolName] = handler
}

// CallTool invokes a tool on this server by name.
func (s *Server) CallTool(ctx context.Context, toolName string, input map[string]interface{}) (map[string]interface{}, error) {
	handler, ok := s.tools[toolName]
	if !ok {
		return nil, fmt.Errorf("tool %q not found on server %q", toolName, s.config.Name)
	}

	log.Printf("[mcp-server] %s.%s called with input: %v", s.config.Name, toolName, input)

	// Add timeout context if configured
	if s.config.TimeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(s.config.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	output, err := handler(ctx, input)
	if err != nil {
		log.Printf("[mcp-server] %s.%s error: %v", s.config.Name, toolName, err)
		return nil, err
	}

	log.Printf("[mcp-server] %s.%s completed", s.config.Name, toolName)
	return output, nil
}

// ListTools returns all tools registered on this server.
func (s *Server) ListTools() []MCPTool {
	var tools []MCPTool
	for name := range s.tools {
		if tool, ok := s.registry.GetTool(s.config.Name, name); ok {
			tools = append(tools, tool)
		}
	}
	return tools
}

// Config returns the server configuration.
func (s *Server) Config() MCPConfig {
	return s.config
}

// CreateMockSUMOServer creates a mock SUMO traffic simulation server.
func CreateMockSUMOServer(registry *Registry) *Server {
	config := MockMCPConfig("SUMO", "simulation")
	config.Description = "SUMO traffic simulation - mock implementation"

	server := NewServer(config, registry)

	// Register: run_simulation
	server.RegisterHandler("run_simulation", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		// Simulate processing time
		time.Sleep(500 * time.Millisecond)

		vehicles := 50
		if v, ok := input["vehicle_count"]; ok {
			if vf, ok := toFloat64(v); ok {
				vehicles = int(vf)
			}
		}

		return map[string]interface{}{
			"simulation_id":     fmt.Sprintf("sumo-%d", rand.Intn(10000)),
			"total_vehicles":    vehicles,
			"average_speed_kmh": 45.3 + rand.Float64()*15,
			"congestion_level":  "moderate",
			"duration_seconds":  rand.Intn(300) + 60,
			"emissions": map[string]interface{}{
				"co2_kg":  rand.Float64() * 100,
				"nox_g":  rand.Float64() * 50,
			},
		}, nil
	})

	// Register: analyze_traffic
	server.RegisterHandler("analyze_traffic", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(300 * time.Millisecond)

		return map[string]interface{}{
			"peak_hour":          "17:30",
			"bottleneck_section": "section_3",
			"average_delay_sec":  rand.Intn(120) + 30,
			"recommendations": []string{
				"Adjust traffic light timing at intersection 3",
				"Add extra lane at section_3",
			},
		}, nil
	})

	// Register: get_road_network
	server.RegisterHandler("get_road_network", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(200 * time.Millisecond)

		return map[string]interface{}{
			"total_nodes":    120,
			"total_edges":    350,
			"total_length_km": 45.7,
			"network_type":   "urban_grid",
		}, nil
	})

	return server
}

// CreateMockMATLABServer creates a mock MATLAB computation server.
func CreateMockMATLABServer(registry *Registry) *Server {
	config := MockMCPConfig("MATLAB", "calculation")
	config.Description = "MATLAB numerical computation - mock implementation"

	server := NewServer(config, registry)

	// Register: matrix_operation
	server.RegisterHandler("matrix_operation", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(400 * time.Millisecond)

		operation, _ := input["operation"].(string)
		rows, _ := input["rows"].(float64)
		cols, _ := input["cols"].(float64)

		return map[string]interface{}{
			"operation":   operation,
			"matrix_size": map[string]interface{}{"rows": rows, "cols": cols},
			"result": map[string]interface{}{
				"determinant":  rand.Float64() * 100,
				"rank":         rand.Intn(10) + 1,
				"condition":    rand.Float64() * 50,
			},
			"computation_time_ms": rand.Intn(200) + 50,
		}, nil
	})

	// Register: fft_analysis
	server.RegisterHandler("fft_analysis", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(600 * time.Millisecond)

		signalLen := 1024
		if v, ok := input["signal_length"]; ok {
			if vf, ok := toFloat64(v); ok {
				signalLen = int(vf)
			}
		}

		frequencies := make([]map[string]interface{}, 5)
		for i := 0; i < 5; i++ {
			frequencies[i] = map[string]interface{}{
				"frequency_hz": float64(i+1) * 10.0,
				"magnitude":    rand.Float64() * 100,
				"phase":        rand.Float64() * 3.14,
			}
		}

		return map[string]interface{}{
			"signal_length": signalLen,
			"frequencies":   frequencies,
			"dominant_freq": 50.0,
		}, nil
	})

	// Register: optimize
	server.RegisterHandler("optimize", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(800 * time.Millisecond)

		iterations := 100
		if v, ok := input["iterations"]; ok {
			if vf, ok := toFloat64(v); ok {
				iterations = int(vf)
			}
		}

		return map[string]interface{}{
			"optimal_value":   rand.Float64() * 200,
			"iterations":      iterations,
			"converged":       true,
			"tolerance":       1e-6,
			"optimal_params": []float64{rand.Float64(), rand.Float64(), rand.Float64()},
		}, nil
	})

	return server
}

// CreateMockVISSIMServer creates a mock VISSIM traffic simulation server.
func CreateMockVISSIMServer(registry *Registry) *Server {
	config := MockMCPConfig("VISSIM", "simulation")
	config.Description = "VISSIM traffic microsimulation - mock implementation"

	server := NewServer(config, registry)

	// Register: run_microsimulation
	server.RegisterHandler("run_microsimulation", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(700 * time.Millisecond)

		intersection, _ := input["intersection"].(string)
		duration := 3600.0
		if v, ok := input["duration_seconds"]; ok {
			if vf, ok := toFloat64(v); ok {
				duration = vf
			}
		}

		return map[string]interface{}{
			"intersection":        intersection,
			"simulation_duration": duration,
			"total_vehicles":      rand.Intn(2000) + 500,
			"average_delay":       rand.Float64() * 60,
			"los":                 "C", // Level of Service
			"queue_length_max":    rand.Intn(50) + 10,
		}, nil
	})

	// Register: calibrate_parameters
	server.RegisterHandler("calibrate_parameters", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(500 * time.Millisecond)

		return map[string]interface{}{
			"car_following": map[string]interface{}{
				"headway":       2.5,
				"reaction_time": 1.0,
			},
			"lane_change": map[string]interface{}{
				"gap_acceptance": 3.0,
				"safety_factor":  0.8,
			},
			"r_squared": 0.92,
		}, nil
	})

	return server
}

// toFloat64 attempts to convert a value to float64.
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}