package generator

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"tempest-homekit-go/pkg/types"
)

// TestDeterministicHistoricalGeneration verifies that with a fixed RNG seed
// and a fixed CurrentTime, GenerateHistoricalData produces the same output
// across multiple generator instances.
func TestDeterministicHistoricalGeneration(t *testing.T) {
	seed := int64(42)
	now := time.Date(2025, 11, 22, 12, 0, 0, 0, time.UTC)

	makeObs := func() []*types.Observation {
		rng := rand.New(rand.NewSource(seed))
		wg := &WeatherGenerator{
			Location:    Location{Name: "DetTest", Latitude: 40.0, Elevation: 100},
			Season:      Summer,
			rng:         rng,
			CurrentTime: now,
		}
		wg.initializeBaseValues()
		// set some known totals to ensure restoration logic exercised
		wg.cumulativeRain = 3.14
		wg.dailyRainTotal = 0.5
		return wg.GenerateHistoricalData(12)
	}

	obs1 := makeObs()
	obs2 := makeObs()

	if len(obs1) != len(obs2) {
		t.Fatalf("length mismatch: %d vs %d", len(obs1), len(obs2))
	}

	for i := range obs1 {
		if !reflect.DeepEqual(obs1[i], obs2[i]) {
			t.Fatalf("observation %d differs:\n%+v\nvs\n%+v", i, obs1[i], obs2[i])
		}
	}
}
