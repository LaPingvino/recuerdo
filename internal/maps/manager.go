// Package maps provides map management functionality for topography lessons
package maps

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// MapPlace represents a place on a map with coordinates
type MapPlace struct {
	X     int      `json:"x"`
	Y     int      `json:"y"`
	Names []string `json:"names"`
}

// BaseMap represents a geographical map with associated places
type BaseMap struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	ImagePath        string                 `json:"imagePath"`
	Places           []MapPlace             `json:"places"`
	Width            int                    `json:"width"`
	Height           int                    `json:"height"`
	Description      string                 `json:"description"`
	IsEmbedded       bool                   `json:"isEmbedded"`
	CoordinateSystem CoordinateSystemConfig `json:"coordinateSystem"`
}

// CoordinateSystemConfig defines how coordinates map to real-world locations
type CoordinateSystemConfig struct {
	// Geographic bounds for coordinate conversion
	MinLatitude  float64 `json:"minLatitude"`
	MaxLatitude  float64 `json:"maxLatitude"`
	MinLongitude float64 `json:"minLongitude"`
	MaxLongitude float64 `json:"maxLongitude"`

	// Map pixel dimensions (can override Width/Height from BaseMap)
	PixelWidth  int `json:"pixelWidth,omitempty"`
	PixelHeight int `json:"pixelHeight,omitempty"`

	// Plus Code configuration
	PlusCodePrecision int `json:"plusCodePrecision"` // Number of digits after decimal
}

// MapManager handles loading and managing base maps
type MapManager struct {
	basePath           string
	maps               map[string]*BaseMap
	externalMapsConfig string       // Path to external maps configuration
	tileManager        *TileManager // Tile-based maps manager
}

// NewMapManager creates a new MapManager instance
func NewMapManager(basePath string) *MapManager {
	mm := &MapManager{
		basePath:           basePath,
		maps:               make(map[string]*BaseMap),
		externalMapsConfig: filepath.Join(basePath, "config", "external_maps.json"),
		tileManager:        NewTileManager(basePath),
	}
	return mm
}

// NewMapManagerWithEmbedded creates a new MapManager instance with embedded filesystem
func NewMapManagerWithEmbedded(basePath string, embeddedFS embed.FS) *MapManager {
	mm := &MapManager{
		basePath:           basePath,
		maps:               make(map[string]*BaseMap),
		externalMapsConfig: filepath.Join(basePath, "config", "external_maps.json"),
		tileManager:        NewTileManager(basePath),
	}
	return mm
}

// LoadAvailableMaps loads both embedded baseline maps and external maps
func (mm *MapManager) LoadAvailableMaps() error {
	// Load from data directory first
	if err := mm.loadFromDataDirectory(); err != nil {
		fmt.Printf("Warning: Failed to load data directory maps: %v\n", err)
	}

	// Then load external maps if configuration exists
	if err := mm.loadExternalMaps(); err != nil {
		fmt.Printf("Warning: Failed to load external maps: %v\n", err)
	}

	// Load tile maps configuration
	if err := mm.tileManager.LoadTileMapConfigs(); err != nil {
		fmt.Printf("Warning: Failed to load tile maps: %v\n", err)
	}

	// Additional info
	if len(mm.maps) > 0 {
		fmt.Printf("Loaded %d total maps\n", len(mm.maps))
	}

	return nil
}

// loadMap loads a single map from its directory
func (mm *MapManager) loadMap(mapID, mapPath string) (*BaseMap, error) {
	placesPath := filepath.Join(mapPath, "resources", "places.json")
	imagePath := filepath.Join(mapPath, "resources", "map.gif")

	// Load places data
	places, err := mm.loadPlaces(placesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load places: %v", err)
	}

	// Check if image exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("map image not found: %s", imagePath)
	}

	// Create map name from ID
	name := strings.Title(mapID)
	if mapID == "latinamerica" {
		name = "Latin America"
	} else if mapID == "usa" {
		name = "United States"
	}

	baseMap := &BaseMap{
		ID:          mapID,
		Name:        name,
		ImagePath:   imagePath,
		Places:      places,
		Width:       800, // Default dimensions - could be read from image
		Height:      600,
		Description: fmt.Sprintf("%s geographical map with %d places", name, len(places)),
	}

	return baseMap, nil
}

// loadPlaces loads places from a JSON file
func (mm *MapManager) loadPlaces(placesPath string) ([]MapPlace, error) {
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

// GetAvailableMaps returns a list of available base maps
func (mm *MapManager) GetAvailableMaps() []*BaseMap {
	var maps []*BaseMap
	for _, baseMap := range mm.maps {
		maps = append(maps, baseMap)
	}
	return maps
}

// GetMap returns a specific map by ID
func (mm *MapManager) GetMap(mapID string) (*BaseMap, error) {
	baseMap, exists := mm.maps[mapID]
	if !exists {
		return nil, fmt.Errorf("map not found: %s", mapID)
	}
	return baseMap, nil
}

// FindNearestPlace finds the closest place to given coordinates
func (mm *MapManager) FindNearestPlace(mapID string, x, y int, threshold int) (*MapPlace, error) {
	baseMap, err := mm.GetMap(mapID)
	if err != nil {
		return nil, err
	}

	var nearest *MapPlace
	minDistance := float64(threshold + 1)

	for i := range baseMap.Places {
		place := &baseMap.Places[i]
		dx := float64(x - place.X)
		dy := float64(y - place.Y)
		distance := dx*dx + dy*dy // Using squared distance for performance

		if distance < minDistance {
			minDistance = distance
			nearest = place
		}
	}

	if minDistance > float64(threshold*threshold) {
		return nil, fmt.Errorf("no place found within threshold")
	}

	return nearest, nil
}

// PlusCode utilities for coordinate conversion
// Plus Codes (Open Location Code) provide a geocoding system

// CoordinateToPlusCode converts map coordinates to Plus Code using the map's coordinate system
func CoordinateToPlusCode(x, y int, baseMap *BaseMap) string {
	// Use the new coordinate system if available
	if baseMap.CoordinateSystem.MinLatitude != 0 || baseMap.CoordinateSystem.MaxLatitude != 0 {
		latitude, longitude := ConvertCoordinateWithSystem(x, y, baseMap)
		precision := baseMap.CoordinateSystem.PlusCodePrecision
		if precision <= 0 {
			precision = 2
		}
		format := fmt.Sprintf("%%.%df+%%.%df", precision, precision)
		return fmt.Sprintf(format, latitude, longitude)
	}

	// Fallback to old method for backward compatibility
	return legacyCoordinateToPlusCode(x, y, baseMap)
}

// legacyCoordinateToPlusCode provides backward compatibility
func legacyCoordinateToPlusCode(x, y int, baseMap *BaseMap) string {
	// Normalize coordinates to 0-1 range
	normalizedX := float64(x) / float64(baseMap.Width)
	normalizedY := float64(y) / float64(baseMap.Height)

	// Convert to approximate lat/lng based on map (very rough approximation)
	var baseLat, baseLng, latRange, lngRange float64

	switch baseMap.ID {
	case "world":
		baseLat, baseLng = -90.0, -180.0
		latRange, lngRange = 180.0, 360.0
	case "europe":
		baseLat, baseLng = 35.0, -25.0
		latRange, lngRange = 35.0, 65.0
	case "usa":
		baseLat, baseLng = 25.0, -125.0
		latRange, lngRange = 25.0, 50.0
	case "africa":
		baseLat, baseLng = -35.0, -20.0
		latRange, lngRange = 70.0, 75.0
	case "asia":
		baseLat, baseLng = 10.0, 60.0
		latRange, lngRange = 70.0, 120.0
	case "latinamerica":
		baseLat, baseLng = -55.0, -120.0
		latRange, lngRange = 90.0, 80.0
	default:
		baseLat, baseLng = 0.0, 0.0
		latRange, lngRange = 180.0, 360.0
	}

	lat := baseLat + normalizedY*latRange
	lng := baseLng + normalizedX*lngRange

	return fmt.Sprintf("%.2f+%.2f", lat, lng)
}

// PlusCodeToCoordinate converts a Plus Code-like string back to map coordinates
func PlusCodeToCoordinate(plusCode string, baseMap *BaseMap) (int, int, error) {
	// Use the new coordinate system if available
	if baseMap.CoordinateSystem.MinLatitude != 0 || baseMap.CoordinateSystem.MaxLatitude != 0 {
		var latitude, longitude float64
		n, err := fmt.Sscanf(plusCode, "%f+%f", &latitude, &longitude)
		if err != nil || n != 2 {
			return 0, 0, fmt.Errorf("invalid Plus Code format: %s", plusCode)
		}
		return ConvertGeographicToCoordinate(latitude, longitude, baseMap)
	}

	// Fallback to old method for backward compatibility
	return legacyPlusCodeToCoordinate(plusCode, baseMap)
}

// legacyPlusCodeToCoordinate provides backward compatibility
func legacyPlusCodeToCoordinate(plusCode string, baseMap *BaseMap) (int, int, error) {
	// Parse the simplified Plus Code format
	parts := strings.Split(plusCode, "+")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid Plus Code format")
	}

	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid latitude in Plus Code")
	}

	lng, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid longitude in Plus Code")
	}

	// Convert back to map coordinates (reverse of the above process)
	var baseLat, baseLng, latRange, lngRange float64

	switch baseMap.ID {
	case "world":
		baseLat, baseLng = -90.0, -180.0
		latRange, lngRange = 180.0, 360.0
	case "europe":
		baseLat, baseLng = 35.0, -25.0
		latRange, lngRange = 35.0, 65.0
	case "usa":
		baseLat, baseLng = 25.0, -125.0
		latRange, lngRange = 25.0, 50.0
	case "africa":
		baseLat, baseLng = -35.0, -20.0
		latRange, lngRange = 70.0, 75.0
	case "asia":
		baseLat, baseLng = 10.0, 60.0
		latRange, lngRange = 70.0, 120.0
	case "latinamerica":
		baseLat, baseLng = -55.0, -120.0
		latRange, lngRange = 90.0, 80.0
	default:
		baseLat, baseLng = 0.0, 0.0
		latRange, lngRange = 180.0, 360.0
	}

	normalizedY := (lat - baseLat) / latRange
	normalizedX := (lng - baseLng) / lngRange

	// Clamp to valid ranges
	if normalizedX < 0 {
		normalizedX = 0
	}
	if normalizedX > 1 {
		normalizedX = 1
	}
	if normalizedY < 0 {
		normalizedY = 0
	}
	if normalizedY > 1 {
		normalizedY = 1
	}

	x := int(normalizedX * float64(baseMap.Width))
	y := int(normalizedY * float64(baseMap.Height))

	return x, y, nil
}

// ValidatePlusCode checks if a Plus Code string is valid
func ValidatePlusCode(plusCode string) bool {
	parts := strings.Split(plusCode, "+")
	if len(parts) != 2 {
		return false
	}

	_, err1 := strconv.ParseFloat(parts[0], 64)
	_, err2 := strconv.ParseFloat(parts[1], 64)

	return err1 == nil && err2 == nil
}

// GetMapBounds returns the coordinate bounds for a given map
func (mm *MapManager) GetMapBounds(mapID string) (width, height int, err error) {
	baseMap, err := mm.GetMap(mapID)
	if err != nil {
		return 0, 0, err
	}
	return baseMap.Width, baseMap.Height, nil
}

// ScaleCoordinates scales coordinates from one map size to another
func ScaleCoordinates(x, y, fromWidth, fromHeight, toWidth, toHeight int) (int, int) {
	scaledX := (x * toWidth) / fromWidth
	scaledY := (y * toHeight) / fromHeight
	return scaledX, scaledY
}

// ConvertCoordinateWithSystem converts map coordinates using the map's coordinate system
func ConvertCoordinateWithSystem(x, y int, baseMap *BaseMap) (latitude, longitude float64) {
	cs := baseMap.CoordinateSystem

	// Use map-specific dimensions if available, otherwise use base map dimensions
	width := cs.PixelWidth
	height := cs.PixelHeight
	if width == 0 {
		width = baseMap.Width
	}
	if height == 0 {
		height = baseMap.Height
	}

	// Convert pixel coordinates to normalized coordinates (0-1)
	normalizedX := float64(x) / float64(width)
	normalizedY := float64(y) / float64(height)

	// Convert to geographic coordinates
	longitude = cs.MinLongitude + normalizedX*(cs.MaxLongitude-cs.MinLongitude)
	latitude = cs.MaxLatitude - normalizedY*(cs.MaxLatitude-cs.MinLatitude) // Y is inverted in images

	return latitude, longitude
}

// ConvertGeographicToCoordinate converts geographic coordinates to map pixel coordinates
func ConvertGeographicToCoordinate(latitude, longitude float64, baseMap *BaseMap) (x, y int, err error) {
	cs := baseMap.CoordinateSystem

	// Validate bounds
	if latitude < cs.MinLatitude || latitude > cs.MaxLatitude {
		return 0, 0, fmt.Errorf("latitude %.2f is out of bounds [%.2f, %.2f]", latitude, cs.MinLatitude, cs.MaxLatitude)
	}
	if longitude < cs.MinLongitude || longitude > cs.MaxLongitude {
		return 0, 0, fmt.Errorf("longitude %.2f is out of bounds [%.2f, %.2f]", longitude, cs.MinLongitude, cs.MaxLongitude)
	}

	// Use map-specific dimensions if available
	width := cs.PixelWidth
	height := cs.PixelHeight
	if width == 0 {
		width = baseMap.Width
	}
	if height == 0 {
		height = baseMap.Height
	}

	// Convert to normalized coordinates
	normalizedX := (longitude - cs.MinLongitude) / (cs.MaxLongitude - cs.MinLongitude)
	normalizedY := (cs.MaxLatitude - latitude) / (cs.MaxLatitude - cs.MinLatitude) // Y is inverted

	// Convert to pixel coordinates
	x = int(normalizedX * float64(width))
	y = int(normalizedY * float64(height))

	// Clamp to valid ranges
	if x < 0 {
		x = 0
	}
	if x >= width {
		x = width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= height {
		y = height - 1
	}

	return x, y, nil
}

// LoadEmbeddedMaps loads baseline maps from embedded filesystem (disabled for now)
func (mm *MapManager) LoadEmbeddedMaps() error {
	return fmt.Errorf("embedded maps disabled - using data directory fallback")
}

// GetTileManager returns the tile manager for tile-based maps
func (mm *MapManager) GetTileManager() *TileManager {
	return mm.tileManager
}

// CreateTileBasedMap creates a BaseMap from a tile map configuration
func (mm *MapManager) CreateTileBasedMap(tileMapID string, north, south, east, west float64, zoom int) (*BaseMap, error) {
	if mm.tileManager == nil {
		return nil, fmt.Errorf("tile manager not initialized")
	}

	tileMap, err := mm.tileManager.GetTileMap(tileMapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tile map: %v", err)
	}

	// Create a tile-based map configuration
	baseMap, err := mm.tileManager.CreateTileMap(tileMap.Config, north, south, east, west, zoom)
	if err != nil {
		return nil, fmt.Errorf("failed to create tile map: %v", err)
	}

	// Store in maps collection with a unique ID
	mapID := fmt.Sprintf("tile_%s_z%d", tileMapID, zoom)
	baseMap.ID = mapID
	baseMap.ImagePath = fmt.Sprintf("tile://%s/%d/%.4f,%.4f,%.4f,%.4f", tileMapID, zoom, north, south, east, west)

	mm.maps[mapID] = baseMap

	return baseMap, nil
}

// DownloadTilesForRegion downloads tiles for a region and caches them
func (mm *MapManager) DownloadTilesForRegion(tileMapID string, north, south, east, west float64, zoom int) error {
	if mm.tileManager == nil {
		return fmt.Errorf("tile manager not initialized")
	}

	return mm.tileManager.DownloadTilesForRegion(tileMapID, north, south, east, west, zoom)
}

// GetAvailableTileMaps returns all available tile map configurations
func (mm *MapManager) GetAvailableTileMaps() []*TileMap {
	if mm.tileManager == nil {
		return nil
	}

	return mm.tileManager.GetAvailableTileMaps()
}

// loadEmbeddedMap loads a single map from embedded resources (disabled)
func (mm *MapManager) loadEmbeddedMap(id, name, imagePath, placesPath string, width, height int, description string) (*BaseMap, error) {
	return nil, fmt.Errorf("embedded maps disabled")
}

// GetEmbeddedMapData returns the binary data for an embedded map image (disabled)
func (mm *MapManager) GetEmbeddedMapData(mapID string) ([]byte, error) {
	return nil, fmt.Errorf("embedded maps disabled")
}

// loadFromDataDirectory provides fallback loading from data directory
func (mm *MapManager) loadFromDataDirectory() error {
	mapDirs := []string{"africa", "asia", "europe", "latinamerica", "usa", "world"}

	for _, mapID := range mapDirs {
		mapPath := filepath.Join(mm.basePath, "data/maps", mapID)

		// Check if map directory exists
		if _, err := os.Stat(mapPath); os.IsNotExist(err) {
			continue
		}

		baseMap, err := mm.loadMap(mapID, mapPath)
		if err != nil {
			fmt.Printf("Warning: Failed to load map %s: %v\n", mapID, err)
			continue
		}

		mm.maps[mapID] = baseMap
	}

	return nil
}
