# testhelpers

This package contains shared helpers used by UI and headless tests.

## SeedObservationsWithOptions

`SeedObservationsWithOptions` generates synthetic `weather.Observation` data for UI tests. It is driven by a small `SeedOptions` struct so tests can produce deterministic and realistic-looking datasets.

Example usage:

```go
base := time.Now().Truncate(time.Second)
chartTypes := []string{"temperature", "humidity", "uv"}

opts := testhelpers.SeedOptions{
    Points:      5,
    Season:      "summer",      // "summer", "winter", or "neutral"
    Location:    "coastal",     // "coastal", "desert", "mountain", or "default"
    TimeSpacing: 5 * time.Minute, // spacing between generated points
    TempNoise:   0.5,             // small gaussian noise magnitude for temperature
    UVNoise:     2,               // small integer noise applied to UV values
    RandSeed:    base.Unix(),     // deterministic seed for reproducible tests
}

var got []weather.Observation
updater := func(o *weather.Observation) { got = append(got, *o) }

testhelpers.SeedObservationsWithOptions(t, updater, chartTypes, base, opts)
```

Notes
- `SeedObservations` remains available as a thin compatibility wrapper that constructs a default `SeedOptions` and calls `SeedObservationsWithOptions`.
- Use `RandSeed` when you need reproducible outputs across test runs.
- Location and season biases are intentionally small; tweak `TempNoise`/`UVNoise` or extend profiles if you need stronger variation for visual tests.

More examples
----------------

Multiple profiles (coastal then desert) in the same test

```go
base := time.Now().Truncate(time.Second)
chartTypes := []string{"temperature", "humidity", "uv"}

var got []weather.Observation
updater := func(o *weather.Observation) { got = append(got, *o) }

// coastal, winter
coastalOpts := testhelpers.SeedOptions{
    Points:      4,
    Season:      "winter",
    Location:    "coastal",
    TimeSpacing: 5 * time.Minute,
    TempNoise:   0.2,
    UVNoise:     1,
    RandSeed:    base.Unix(),
}
testhelpers.SeedObservationsWithOptions(t, updater, chartTypes, base, coastalOpts)

// desert, summer (a separate batch of observations for the same charts)
desertOpts := testhelpers.SeedOptions{
    Points:      4,
    Season:      "summer",
    Location:    "desert",
    TimeSpacing: 5 * time.Minute,
    TempNoise:   0.5,
    UVNoise:     2,
    RandSeed:    base.Add(1 * time.Hour).Unix(),
}
testhelpers.SeedObservationsWithOptions(t, updater, chartTypes, base.Add(-6*time.Hour), desertOpts)
```

Combining options for visual stress tests

If you need more variation to stress rendering (e.g., for long sparkline charts), increase `Points`, `TempNoise`, and `UVNoise`, or run multiple profiles back-to-back. Keep `RandSeed` deterministic when you want identical runs in CI.

Available locations and seasons
-------------------------------

The seeder provides a few small built-in location profiles and seasonal biases. These are intentionally conservative; treat them as starting points you can extend.

| Location  | Temperature bias (°C) | UV bias (integer) | Notes |
|-----------|------------------------:|------------------:|-------|
| default   | 0.0                    | 0                 | No special bias |
| coastal   | -1.5                   | -1                | Mildly cooler, slightly lower UV due to marine influence |
| desert    | 3.0                    | 2                 | Warmer and higher UV |
| mountain  | -3.0                   | -2                | Cooler and lower UV |

| Season    | Temperature bias (°C) | UV bias (integer) | Notes |
|-----------|------------------------:|------------------:|-------|
| summer    | +5.0                   | +2                | Warmer temps and higher UV |
| winter    | -5.0                   | 0                 | Colder temps, low UV |
| neutral   | 0.0                    | 0                 | No seasonal bias |

Extending profiles
-------------------

If you need site-specific behavior (e.g., seasonal amplitude, humidity effects, or special rain patterns), consider adding a small helper that maps a `Location` or `SeedOptions` to a custom per-chart scaling function. The provided `SeedObservationsWithOptions` is intentionally small so it's easy to extend in tests without pulling in heavy fixtures.

If you'd like, I can add an example of a custom profile hook and a tiny helper to register additional location profiles.
