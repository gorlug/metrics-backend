package journal

import (
	"fmt"
	"time"
)

func GetLocation() *time.Location {
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		println(fmt.Sprintf("failed to load location %v, error: %v", "Europe/Berlin", err))
		location = time.Local
	}
	return location
}
