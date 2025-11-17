package main

import (
	"fmt"
	"log"

	"github.com/LaPingvino/recuerdo/internal/maps"
)

func main() {
	fmt.Println("Testing map loading...")

	// Initialize map manager with embedded maps support
	embeddedFS := GetEmbeddedMapsFS()
	mapManager := maps.NewMapManagerWithEmbedded("./", embeddedFS)

	// Load available maps
	err := mapManager.LoadAvailableMaps()
	if err != nil {
		log.Fatalf("Failed to load maps: %v", err)
	}

	// Get available maps
	availableMaps := mapManager.GetAvailableMaps()
	fmt.Printf("Found %d available maps:\n", len(availableMaps))

	for _, baseMap := range availableMaps {
		fmt.Printf("- %s (ID: %s)\n", baseMap.Name, baseMap.ID)
		fmt.Printf("  Image: %s\n", baseMap.ImagePath)
		fmt.Printf("  Places: %d\n", len(baseMap.Places))
		fmt.Printf("  Description: %s\n", baseMap.Description)

		// Test Plus Code functionality
		if len(baseMap.Places) > 0 {
			place := baseMap.Places[0]
			plusCode := maps.CoordinateToPlusCode(place.X, place.Y, baseMap)
			fmt.Printf("  Sample place: %s at (%d, %d) -> Plus Code: %s\n",
				place.Names[0], place.X, place.Y, plusCode)

			// Test conversion back
			backX, backY, err := maps.PlusCodeToCoordinate(plusCode, baseMap)
			if err != nil {
				fmt.Printf("    Error converting back: %v\n", err)
			} else {
				fmt.Printf("    Converted back: (%d, %d)\n", backX, backY)
			}
		}
		fmt.Println()
	}

	// Test coordinate scaling
	fmt.Println("Testing coordinate scaling:")
	scaledX, scaledY := maps.ScaleCoordinates(400, 300, 800, 600, 400, 300)
	fmt.Printf("Scaling (400, 300) from 800x600 to 400x300: (%d, %d)\n", scaledX, scaledY)

	fmt.Println("Map testing complete!")
}
