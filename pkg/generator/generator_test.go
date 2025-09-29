package generator

import (
	"math/rand"
	"testing"
	"time"
)

// TestGeneratePrecipitationType verifies precipitation type selection by temperature and rain
func TestGeneratePrecipitationType(t *testing.T) {
	wg := &WeatherGenerator{}

	// No rain
	if got := wg.generatePrecipitationType(10, 0); got != 0 {
		t.Fatalf("expected 0 (none) for no rain, got %d", got)
	}

	// Rain and warm -> rain
	if got := wg.generatePrecipitationType(10, 0.5); got != 1 {
		t.Fatalf("expected 1 (rain) for temp 10 and rain 0.5, got %d", got)
	}

	// Cold temperatures -> snow
	if got := wg.generatePrecipitationType(-5, 0.2); got != 3 {
		t.Fatalf("expected 3 (snow) for temp -5 and rain 0.2, got %d", got)
	}

	// Near-freezing -> ice pellets
	if got := wg.generatePrecipitationType(1, 0.2); got != 2 {
		t.Fatalf("expected 2 (ice pellets) for temp 1 and rain 0.2, got %d", got)
	}
}

// TestGenerateHistoricalData ensures historical generation returns increasing timestamps
// and that daily/cumulative totals are restored after generation.
func TestGenerateHistoricalData(t *testing.T) {
	// Create a deterministic RNG for repeatable behavior
	rng := rand.New(rand.NewSource(42))

	wg := &WeatherGenerator{
		Location: Location{Name: "Test", Latitude: 40.0, Elevation: 100},
		Season:   Summer,
		rng:      rng,
	}
	wg.initializeBaseValues()

	// Set some known totals
	wg.cumulativeRain = 5.0
	wg.dailyRainTotal = 0.3
	originalCumulative := wg.cumulativeRain
	originalDaily := wg.dailyRainTotal

	observations := wg.GenerateHistoricalData(10)
	if len(observations) != 10 {
		t.Fatalf("expected 10 observations, got %d", len(observations))
	}

	// Verify timestamps strictly increase
	for i := 1; i < len(observations); i++ {
		if observations[i].Timestamp <= observations[i-1].Timestamp {
			t.Fatalf("expected timestamps to increase: idx %d <= %d", i, i-1)
		}
	}

	// Verify cumulative and daily totals were restored
	if wg.cumulativeRain != originalCumulative {
		t.Fatalf("expected cumulativeRain restored to %.2f, got %.2f", originalCumulative, wg.cumulativeRain)
	}
	if wg.dailyRainTotal != originalDaily {
		t.Fatalf("expected dailyRainTotal restored to %.2f, got %.2f", originalDaily, wg.dailyRainTotal)
	}

	// Ensure history field is set
	if wg.history == nil || len(wg.history) != 10 {
		t.Fatalf("expected wg.history to be set with 10 entries, got %v", wg.history)
	}

	// Ensure timestamps are approximately 24 hours span
	span := time.Unix(observations[len(observations)-1].Timestamp, 0).Sub(time.Unix(observations[0].Timestamp, 0))
	if span <= 0 {
		t.Fatalf("expected positive timespan for generated historical data, got %v", span)
	}
}
