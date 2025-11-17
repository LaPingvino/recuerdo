// Package maps provides map tile functionality for dynamic map loading
package maps

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TileMapConfig represents configuration for a tile-based map
type TileMapConfig struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TileURL     string   `json:"tileUrl"`     // e.g., "https://cartodb-basemaps-{s}.global.ssl.fastly.net/light_nolabels/{z}/{x}/{y}.png"
	Subdomains  []string `json:"subdomains"`  // e.g., ["a", "b", "c", "d"]
	MinZoom     int      `json:"minZoom"`     // Minimum zoom level
	MaxZoom     int      `json:"maxZoom"`     // Maximum zoom level
	DefaultZoom int      `json:"defaultZoom"` // Default zoom level for lessons

	// Geographic bounds for the tileset
	BoundingBox struct {
		North float64 `json:"north"`
		South float64 `json:"south"`
		East  float64 `json:"east"`
		West  float64 `json:"west"`
	} `json:"boundingBox"`

	// Tile configuration
	TileSize    int    `json:"tileSize"`    // Usually 256 or 512
	Attribution string `json:"attribution"` // Copyright/attribution text

	// Caching settings
	CacheEnabled bool `json:"cacheEnabled"`
	CacheTTL     int  `json:"cacheTtl"` // Cache time-to-live in hours

	// Optional metadata
	Author   string            `json:"author,omitempty"`
	License  string            `json:"license,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// TileCoordinate represents a specific map tile
type TileCoordinate struct {
	Z int // Zoom level
	X int // Tile X coordinate
	Y int // Tile Y coordinate
}

// TileMap represents a tile-based map with caching
type TileMap struct {
	Config     TileMapConfig
	Cache      *TileCache
	httpClient *http.Client
}

// TileCache handles caching of downloaded map tiles
type TileCache struct {
	basePath string
	mutex    sync.RWMutex
	stats    struct {
		hits   int64
		misses int64
		errors int64
	}
}

// TileManager manages tile-based maps
type TileManager struct {
	basePath   string
	configPath string
	tileMaps   map[string]*TileMap
	cache      *TileCache
	mutex      sync.RWMutex
}

// NewTileManager creates a new tile manager
func NewTileManager(basePath string) *TileManager {
	cacheDir := filepath.Join(basePath, "cache", "tiles")
	os.MkdirAll(cacheDir, 0755)

	return &TileManager{
		basePath:   basePath,
		configPath: filepath.Join(basePath, "config", "tile_maps.json"),
		tileMaps:   make(map[string]*TileMap),
		cache:      NewTileCache(cacheDir),
	}
}

// NewTileCache creates a new tile cache
func NewTileCache(basePath string) *TileCache {
	return &TileCache{
		basePath: basePath,
	}
}

// LoadTileMapConfigs loads tile map configurations from file
func (tm *TileManager) LoadTileMapConfigs() error {
	// Create default config if it doesn't exist
	if _, err := os.Stat(tm.configPath); os.IsNotExist(err) {
		if err := tm.CreateDefaultTileConfig(); err != nil {
			return fmt.Errorf("failed to create default tile config: %v", err)
		}
	}

	file, err := os.Open(tm.configPath)
	if err != nil {
		return fmt.Errorf("failed to open tile config: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read tile config: %v", err)
	}

	var configs struct {
		Version  string          `json:"version"`
		TileMaps []TileMapConfig `json:"tileMaps"`
	}

	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse tile config: %v", err)
	}

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	for _, config := range configs.TileMaps {
		tileMap := &TileMap{
			Config: config,
			Cache:  tm.cache,
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}

		tm.tileMaps[config.ID] = tileMap
		fmt.Printf("Loaded tile map: %s (%s)\n", config.Name, config.Description)
	}

	return nil
}

// CreateDefaultTileConfig creates a default tile maps configuration
func (tm *TileManager) CreateDefaultTileConfig() error {
	configDir := filepath.Dir(tm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	defaultConfig := struct {
		Version     string          `json:"version"`
		Description string          `json:"description"`
		TileMaps    []TileMapConfig `json:"tileMaps"`
	}{
		Version:     "1.0",
		Description: "Tile-based maps configuration for Recuerdo topography lessons",
		TileMaps: []TileMapConfig{
			{
				ID:          "cartodb_light_nolabels",
				Name:        "CartoDB Light (No Labels)",
				Description: "Clean light map without labels, perfect for topography training",
				TileURL:     "https://cartodb-basemaps-{s}.global.ssl.fastly.net/light_nolabels/{z}/{x}/{y}.png",
				Subdomains:  []string{"a", "b", "c", "d"},
				MinZoom:     0,
				MaxZoom:     18,
				DefaultZoom: 6,
				BoundingBox: struct {
					North float64 `json:"north"`
					South float64 `json:"south"`
					East  float64 `json:"east"`
					West  float64 `json:"west"`
				}{
					North: 85.0511,
					South: -85.0511,
					East:  180.0,
					West:  -180.0,
				},
				TileSize:     256,
				Attribution:  "© OpenStreetMap contributors, © CartoDB",
				CacheEnabled: true,
				CacheTTL:     168, // 1 week
				License:      "ODbL",
				Tags:         []string{"light", "clean", "training", "global"},
			},
			{
				ID:          "cartodb_positron",
				Name:        "CartoDB Positron",
				Description: "Very light map with minimal features, excellent contrast for place names",
				TileURL:     "https://cartodb-basemaps-{s}.global.ssl.fastly.net/light_all/{z}/{x}/{y}.png",
				Subdomains:  []string{"a", "b", "c", "d"},
				MinZoom:     0,
				MaxZoom:     18,
				DefaultZoom: 6,
				BoundingBox: struct {
					North float64 `json:"north"`
					South float64 `json:"south"`
					East  float64 `json:"east"`
					West  float64 `json:"west"`
				}{
					North: 85.0511,
					South: -85.0511,
					East:  180.0,
					West:  -180.0,
				},
				TileSize:     256,
				Attribution:  "© OpenStreetMap contributors, © CartoDB",
				CacheEnabled: true,
				CacheTTL:     168,
				License:      "ODbL",
				Tags:         []string{"light", "minimal", "contrast", "global"},
			},
			{
				ID:          "openstreetmap_standard",
				Name:        "OpenStreetMap Standard",
				Description: "Standard OpenStreetMap tiles with full detail",
				TileURL:     "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
				Subdomains:  []string{"a", "b", "c"},
				MinZoom:     0,
				MaxZoom:     19,
				DefaultZoom: 8,
				BoundingBox: struct {
					North float64 `json:"north"`
					South float64 `json:"south"`
					East  float64 `json:"east"`
					West  float64 `json:"west"`
				}{
					North: 85.0511,
					South: -85.0511,
					East:  180.0,
					West:  -180.0,
				},
				TileSize:     256,
				Attribution:  "© OpenStreetMap contributors",
				CacheEnabled: true,
				CacheTTL:     72, // 3 days
				License:      "ODbL",
				Tags:         []string{"standard", "detailed", "osm", "global"},
			},
		},
	}

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tm.configPath, data, 0644)
}

// GetTileMap returns a tile map by ID
func (tm *TileManager) GetTileMap(id string) (*TileMap, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tileMap, exists := tm.tileMaps[id]
	if !exists {
		return nil, fmt.Errorf("tile map not found: %s", id)
	}

	return tileMap, nil
}

// GetAvailableTileMaps returns all available tile maps
func (tm *TileManager) GetAvailableTileMaps() []*TileMap {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var tileMaps []*TileMap
	for _, tileMap := range tm.tileMaps {
		tileMaps = append(tileMaps, tileMap)
	}

	return tileMaps
}

// DownloadTile downloads a specific tile and returns the image data
func (t *TileMap) DownloadTile(coord TileCoordinate) ([]byte, error) {
	// Check cache first
	if t.Config.CacheEnabled {
		if data, err := t.Cache.GetTile(t.Config.ID, coord); err == nil {
			return data, nil
		}
	}

	// Build tile URL
	url := t.buildTileURL(coord)

	// Download tile
	resp, err := t.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download tile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tile server returned status %d for %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tile data: %v", err)
	}

	// Cache the tile
	if t.Config.CacheEnabled {
		if err := t.Cache.StoreTile(t.Config.ID, coord, data); err != nil {
			fmt.Printf("Warning: failed to cache tile: %v\n", err)
		}
	}

	return data, nil
}

// buildTileURL constructs the full URL for a tile
func (t *TileMap) buildTileURL(coord TileCoordinate) string {
	url := t.Config.TileURL

	// Replace template variables
	url = strings.ReplaceAll(url, "{z}", strconv.Itoa(coord.Z))
	url = strings.ReplaceAll(url, "{x}", strconv.Itoa(coord.X))
	url = strings.ReplaceAll(url, "{y}", strconv.Itoa(coord.Y))

	// Replace subdomain if available
	if len(t.Config.Subdomains) > 0 {
		subdomain := t.Config.Subdomains[coord.X%len(t.Config.Subdomains)]
		url = strings.ReplaceAll(url, "{s}", subdomain)
	}

	return url
}

// GetTilesForBounds returns tile coordinates for a geographic bounding box
func (t *TileMap) GetTilesForBounds(north, south, east, west float64, zoom int) []TileCoordinate {
	var tiles []TileCoordinate

	// Convert geographic bounds to tile coordinates
	// Get all four corners to ensure we cover the full bounding box
	nwX, nwY := deg2tile(north, west, zoom) // North-west corner
	neX, neY := deg2tile(north, east, zoom) // North-east corner
	swX, swY := deg2tile(south, west, zoom) // South-west corner
	seX, seY := deg2tile(south, east, zoom) // South-east corner

	// Find the actual bounds from all corners
	minX := nwX
	maxX := nwX
	minY := nwY
	maxY := nwY

	// Check all corners to find true min/max
	corners := [][2]int{{nwX, nwY}, {neX, neY}, {swX, swY}, {seX, seY}}
	for _, corner := range corners {
		x, y := corner[0], corner[1]
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	// Ensure bounds are within tile limits
	tileLimit := int(math.Pow(2, float64(zoom)))
	if minX < 0 {
		minX = 0
	}
	if maxX >= tileLimit {
		maxX = tileLimit - 1
	}
	if minY < 0 {
		minY = 0
	}
	if maxY >= tileLimit {
		maxY = tileLimit - 1
	}

	// Generate tile coordinates
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			tiles = append(tiles, TileCoordinate{Z: zoom, X: x, Y: y})
		}
	}

	return tiles
}

// GetTile retrieves a tile from cache
func (tc *TileCache) GetTile(mapID string, coord TileCoordinate) ([]byte, error) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	tilePath := tc.getTilePath(mapID, coord)

	// Check if file exists and is not expired
	stat, err := os.Stat(tilePath)
	if err != nil {
		tc.stats.misses++
		return nil, fmt.Errorf("tile not in cache")
	}

	// Check TTL (simple implementation - could be more sophisticated)
	if time.Since(stat.ModTime()).Hours() > 168 { // 1 week default
		tc.stats.misses++
		os.Remove(tilePath) // Remove expired tile
		return nil, fmt.Errorf("tile expired")
	}

	data, err := os.ReadFile(tilePath)
	if err != nil {
		tc.stats.errors++
		return nil, fmt.Errorf("failed to read cached tile: %v", err)
	}

	tc.stats.hits++
	return data, nil
}

// StoreTile stores a tile in cache
func (tc *TileCache) StoreTile(mapID string, coord TileCoordinate, data []byte) error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	tilePath := tc.getTilePath(mapID, coord)
	tileDir := filepath.Dir(tilePath)

	// Ensure directory exists
	if err := os.MkdirAll(tileDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Write tile data
	if err := os.WriteFile(tilePath, data, 0644); err != nil {
		tc.stats.errors++
		return fmt.Errorf("failed to write tile to cache: %v", err)
	}

	return nil
}

// getTilePath returns the file system path for a tile
func (tc *TileCache) getTilePath(mapID string, coord TileCoordinate) string {
	return filepath.Join(tc.basePath, mapID, strconv.Itoa(coord.Z), strconv.Itoa(coord.X), fmt.Sprintf("%d.png", coord.Y))
}

// GetCacheStats returns cache statistics
func (tc *TileCache) GetCacheStats() (hits, misses, errors int64) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	return tc.stats.hits, tc.stats.misses, tc.stats.errors
}

// ClearCache removes all cached tiles for a specific map
func (tc *TileCache) ClearCache(mapID string) error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	cachePath := filepath.Join(tc.basePath, mapID)
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil // Nothing to clear
	}

	return os.RemoveAll(cachePath)
}

// deg2tile converts geographic coordinates to tile coordinates
func deg2tile(lat, lng float64, zoom int) (int, int) {
	latRad := lat * math.Pi / 180.0
	n := math.Pow(2.0, float64(zoom))
	x := int((lng + 180.0) / 360.0 * n)
	y := int((1.0 - math.Asinh(math.Tan(latRad))/math.Pi) / 2.0 * n)
	return x, y
}

// tile2deg converts tile coordinates to geographic coordinates
func tile2deg(x, y, zoom int) (float64, float64) {
	n := math.Pow(2.0, float64(zoom))
	lng := float64(x)/n*360.0 - 180.0
	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n)))
	lat := latRad * 180.0 / math.Pi
	return lat, lng
}

// ValidateTileMapConfig validates a tile map configuration
func ValidateTileMapConfig(config TileMapConfig) error {
	if config.ID == "" {
		return fmt.Errorf("tile map ID cannot be empty")
	}

	if config.Name == "" {
		return fmt.Errorf("tile map name cannot be empty")
	}

	if config.TileURL == "" {
		return fmt.Errorf("tile URL cannot be empty")
	}

	if !strings.Contains(config.TileURL, "{z}") || !strings.Contains(config.TileURL, "{x}") || !strings.Contains(config.TileURL, "{y}") {
		return fmt.Errorf("tile URL must contain {z}, {x}, and {y} placeholders")
	}

	if config.MinZoom < 0 || config.MaxZoom > 20 || config.MinZoom > config.MaxZoom {
		return fmt.Errorf("invalid zoom range: min=%d, max=%d", config.MinZoom, config.MaxZoom)
	}

	if config.DefaultZoom < config.MinZoom || config.DefaultZoom > config.MaxZoom {
		return fmt.Errorf("default zoom %d must be within zoom range [%d, %d]", config.DefaultZoom, config.MinZoom, config.MaxZoom)
	}

	if config.TileSize <= 0 {
		config.TileSize = 256 // Default tile size
	}

	return nil
}

// CreateTileMap creates a new tile map from geographic bounds
func (tm *TileManager) CreateTileMap(config TileMapConfig, north, south, east, west float64, zoom int) (*BaseMap, error) {
	// Validate configuration
	if err := ValidateTileMapConfig(config); err != nil {
		return nil, fmt.Errorf("invalid tile map config: %v", err)
	}

	// Get tile map
	tileMap, exists := tm.tileMaps[config.ID]
	if !exists {
		return nil, fmt.Errorf("tile map not configured: %s", config.ID)
	}

	// Get tiles for bounds
	tiles := tileMap.GetTilesForBounds(north, south, east, west, zoom)
	if len(tiles) == 0 {
		return nil, fmt.Errorf("no tiles found for bounds")
	}

	if len(tiles) > 100 {
		return nil, fmt.Errorf("too many tiles required (%d), reduce area or zoom level", len(tiles))
	}

	// Create composite map from tiles (this would need image stitching implementation)
	// For now, we'll create a BaseMap structure with tile information
	baseMap := &BaseMap{
		ID:          fmt.Sprintf("%s_z%d", config.ID, zoom),
		Name:        fmt.Sprintf("%s (Zoom %d)", config.Name, zoom),
		Description: fmt.Sprintf("Tile-based map from %s at zoom level %d", config.Name, zoom),
		Width:       len(tiles) * config.TileSize, // Simplified - would need proper calculation
		Height:      config.TileSize,              // Simplified - would need proper calculation
		IsEmbedded:  false,
		CoordinateSystem: CoordinateSystemConfig{
			MinLatitude:       south,
			MaxLatitude:       north,
			MinLongitude:      west,
			MaxLongitude:      east,
			PlusCodePrecision: 4,
		},
		Places: []MapPlace{}, // Would be populated based on user input or external data
	}

	return baseMap, nil
}

// DownloadTilesForRegion downloads all tiles for a specific region and zoom level
func (tm *TileManager) DownloadTilesForRegion(mapID string, north, south, east, west float64, zoom int) error {
	tileMap, err := tm.GetTileMap(mapID)
	if err != nil {
		return err
	}

	tiles := tileMap.GetTilesForBounds(north, south, east, west, zoom)

	fmt.Printf("Downloading %d tiles for region...\n", len(tiles))

	// Download tiles concurrently (with rate limiting)
	semaphore := make(chan struct{}, 5) // Limit to 5 concurrent downloads
	var wg sync.WaitGroup

	for i, tile := range tiles {
		wg.Add(1)
		go func(tile TileCoordinate, index int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			if _, err := tileMap.DownloadTile(tile); err != nil {
				fmt.Printf("Failed to download tile %d/%d (%d,%d,%d): %v\n", index+1, len(tiles), tile.Z, tile.X, tile.Y, err)
			} else if (index+1)%10 == 0 {
				fmt.Printf("Downloaded %d/%d tiles\n", index+1, len(tiles))
			}
		}(tile, i)
	}

	wg.Wait()
	fmt.Printf("Download complete for %s\n", mapID)

	return nil
}

// GetCacheStats returns cache statistics for the tile manager
func (tm *TileManager) GetCacheStats() (hits, misses, errors int64) {
	return tm.cache.GetCacheStats()
}
