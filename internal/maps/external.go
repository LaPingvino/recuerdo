package maps

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ExternalMapConfig represents configuration for user-defined maps
type ExternalMapConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ImagePath   string `json:"imagePath"`  // Can be absolute or relative to config file
	PlacesPath  string `json:"placesPath"` // Can be absolute or relative to config file
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Description string `json:"description"`

	// Coordinate system configuration
	CoordinateSystem CoordinateSystemConfig `json:"coordinateSystem"`

	// Optional metadata
	Author   string            `json:"author,omitempty"`
	Version  string            `json:"version,omitempty"`
	License  string            `json:"license,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ExternalMapsConfiguration holds all external map configurations
type ExternalMapsConfiguration struct {
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Maps        []ExternalMapConfig `json:"maps"`
}

// loadExternalMaps loads maps from external configuration file
func (mm *MapManager) loadExternalMaps() error {
	if _, err := os.Stat(mm.externalMapsConfig); os.IsNotExist(err) {
		return nil // No external config file, that's fine
	}

	file, err := os.Open(mm.externalMapsConfig)
	if err != nil {
		return fmt.Errorf("failed to open external maps config: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read external maps config: %v", err)
	}

	var config ExternalMapsConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse external maps config: %v", err)
	}

	configDir := filepath.Dir(mm.externalMapsConfig)

	for _, mapConfig := range config.Maps {
		baseMap, err := mm.loadExternalMap(mapConfig, configDir)
		if err != nil {
			fmt.Printf("Warning: Failed to load external map %s: %v\n", mapConfig.ID, err)
			continue
		}

		// Check for ID conflicts with embedded maps
		if existingMap, exists := mm.maps[mapConfig.ID]; exists {
			if existingMap.IsEmbedded {
				fmt.Printf("Warning: External map %s conflicts with embedded map, using external version\n", mapConfig.ID)
			} else {
				fmt.Printf("Warning: Duplicate external map ID %s, using first occurrence\n", mapConfig.ID)
				continue
			}
		}

		mm.maps[mapConfig.ID] = baseMap
		fmt.Printf("Loaded external map: %s (%d places)\n", baseMap.Name, len(baseMap.Places))
	}

	return nil
}

// loadExternalMap loads a single external map from its configuration
func (mm *MapManager) loadExternalMap(config ExternalMapConfig, configDir string) (*BaseMap, error) {
	// Resolve paths relative to config directory if not absolute
	imagePath := config.ImagePath
	if !filepath.IsAbs(imagePath) {
		imagePath = filepath.Join(configDir, imagePath)
	}

	placesPath := config.PlacesPath
	if !filepath.IsAbs(placesPath) {
		placesPath = filepath.Join(configDir, placesPath)
	}

	// Load places data
	places, err := mm.loadPlacesFromFile(placesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load places from %s: %v", placesPath, err)
	}

	// Check if image exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("map image not found: %s", imagePath)
	}

	baseMap := &BaseMap{
		ID:               config.ID,
		Name:             config.Name,
		ImagePath:        imagePath,
		Places:           places,
		Width:            config.Width,
		Height:           config.Height,
		Description:      config.Description,
		IsEmbedded:       false,
		CoordinateSystem: config.CoordinateSystem,
	}

	return baseMap, nil
}

// loadPlacesFromFile loads places from a JSON file (helper function)
func (mm *MapManager) loadPlacesFromFile(placesPath string) ([]MapPlace, error) {
	file, err := os.Open(placesPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var places []MapPlace
	if err := json.Unmarshal(data, &places); err != nil {
		return nil, err
	}

	return places, nil
}

// CreateExternalMapConfig creates a new external map configuration file
func (mm *MapManager) CreateExternalMapConfig() error {
	// Ensure config directory exists
	configDir := filepath.Dir(mm.externalMapsConfig)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Create a sample configuration
	sampleConfig := ExternalMapsConfiguration{
		Version:     "1.0",
		Description: "External maps configuration for Recuerdo topography lessons",
		Maps: []ExternalMapConfig{
			{
				ID:          "custom_region",
				Name:        "Custom Region",
				ImagePath:   "maps/custom_region.png",
				PlacesPath:  "maps/custom_region_places.json",
				Width:       800,
				Height:      600,
				Description: "A custom regional map for specialized learning",
				Author:      "Map Creator",
				Version:     "1.0",
				License:     "CC BY-SA 4.0",
				Tags:        []string{"custom", "regional", "example"},
				CoordinateSystem: CoordinateSystemConfig{
					MinLatitude:       40.0,
					MaxLatitude:       50.0,
					MinLongitude:      -10.0,
					MaxLongitude:      10.0,
					PlusCodePrecision: 3,
				},
				Metadata: map[string]string{
					"source":     "Custom survey data",
					"projection": "Web Mercator",
					"datum":      "WGS84",
					"created":    "2024",
				},
			},
		},
	}

	// Marshal to JSON with pretty printing
	data, err := json.MarshalIndent(sampleConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write to file
	if err := os.WriteFile(mm.externalMapsConfig, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetExternalMapsConfigPath returns the path to the external maps configuration
func (mm *MapManager) GetExternalMapsConfigPath() string {
	return mm.externalMapsConfig
}

// ValidateExternalMapConfig validates an external map configuration
func ValidateExternalMapConfig(config ExternalMapConfig) error {
	if config.ID == "" {
		return fmt.Errorf("map ID cannot be empty")
	}

	if config.Name == "" {
		return fmt.Errorf("map name cannot be empty")
	}

	if config.ImagePath == "" {
		return fmt.Errorf("image path cannot be empty")
	}

	if config.PlacesPath == "" {
		return fmt.Errorf("places path cannot be empty")
	}

	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("map dimensions must be positive")
	}

	// Validate coordinate system
	cs := config.CoordinateSystem
	if cs.MinLatitude >= cs.MaxLatitude {
		return fmt.Errorf("invalid latitude range: min (%.2f) must be less than max (%.2f)",
			cs.MinLatitude, cs.MaxLatitude)
	}

	if cs.MinLongitude >= cs.MaxLongitude {
		return fmt.Errorf("invalid longitude range: min (%.2f) must be less than max (%.2f)",
			cs.MinLongitude, cs.MaxLongitude)
	}

	if cs.MinLatitude < -90.0 || cs.MaxLatitude > 90.0 {
		return fmt.Errorf("latitude values must be between -90 and 90 degrees")
	}

	if cs.MinLongitude < -180.0 || cs.MaxLongitude > 180.0 {
		return fmt.Errorf("longitude values must be between -180 and 180 degrees")
	}

	return nil
}

// ListExternalMaps returns information about configured external maps
func (mm *MapManager) ListExternalMaps() ([]ExternalMapConfig, error) {
	if _, err := os.Stat(mm.externalMapsConfig); os.IsNotExist(err) {
		return nil, nil // No config file exists
	}

	file, err := os.Open(mm.externalMapsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open external maps config: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read external maps config: %v", err)
	}

	var config ExternalMapsConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse external maps config: %v", err)
	}

	return config.Maps, nil
}

// AddExternalMap adds a new external map to the configuration
func (mm *MapManager) AddExternalMap(mapConfig ExternalMapConfig) error {
	// Validate the configuration
	if err := ValidateExternalMapConfig(mapConfig); err != nil {
		return fmt.Errorf("invalid map configuration: %v", err)
	}

	// Load existing configuration or create new one
	var config ExternalMapsConfiguration
	if _, err := os.Stat(mm.externalMapsConfig); os.IsNotExist(err) {
		config = ExternalMapsConfiguration{
			Version:     "1.0",
			Description: "External maps configuration for Recuerdo topography lessons",
			Maps:        []ExternalMapConfig{},
		}
	} else {
		file, err := os.Open(mm.externalMapsConfig)
		if err != nil {
			return fmt.Errorf("failed to open existing config: %v", err)
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read existing config: %v", err)
		}

		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing config: %v", err)
		}
	}

	// Check for duplicate IDs
	for _, existing := range config.Maps {
		if existing.ID == mapConfig.ID {
			return fmt.Errorf("map with ID '%s' already exists", mapConfig.ID)
		}
	}

	// Add the new map
	config.Maps = append(config.Maps, mapConfig)

	// Save back to file
	return mm.saveExternalMapsConfig(config)
}

// RemoveExternalMap removes an external map from the configuration
func (mm *MapManager) RemoveExternalMap(mapID string) error {
	// Load existing configuration
	file, err := os.Open(mm.externalMapsConfig)
	if err != nil {
		return fmt.Errorf("failed to open config: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	var config ExternalMapsConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// Find and remove the map
	found := false
	for i, mapConfig := range config.Maps {
		if mapConfig.ID == mapID {
			config.Maps = append(config.Maps[:i], config.Maps[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("map with ID '%s' not found", mapID)
	}

	// Save back to file
	return mm.saveExternalMapsConfig(config)
}

// saveExternalMapsConfig saves the configuration to file
func (mm *MapManager) saveExternalMapsConfig(config ExternalMapsConfiguration) error {
	// Ensure config directory exists
	configDir := filepath.Dir(mm.externalMapsConfig)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Marshal to JSON with pretty printing
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write to file
	if err := os.WriteFile(mm.externalMapsConfig, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
