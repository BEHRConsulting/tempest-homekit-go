package alarm

import (
	"fmt"
	"strconv"
	"strings"

	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/weather"
)

// Evaluator evaluates alarm conditions against weather observations
type Evaluator struct{}

// NewEvaluator creates a new alarm evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate checks if an alarm condition is met given weather data
// alarm parameter is optional and only needed for change-detection operators
func (e *Evaluator) Evaluate(condition string, obs *weather.Observation) (bool, error) {
	return e.EvaluateWithAlarm(condition, obs, nil)
}

// EvaluateWithAlarm checks if an alarm condition is met given weather data and alarm state
func (e *Evaluator) EvaluateWithAlarm(condition string, obs *weather.Observation, alarm *Alarm) (bool, error) {
	// Parse and evaluate the condition
	// Supports: >, <, >=, <=, ==, !=, &&, ||, *field, >field, <field
	// Examples:
	//   "temperature > 85"
	//   "humidity > 80 && temperature > 35"
	//   "lux > 10000 && lux < 50000"
	//   "rain_rate > 0"
	//   "*lightning_count" (triggers on any change)
	//   ">rain_rate" (triggers when rain increases)
	//   "<lightning_distance" (triggers when lightning gets closer)

	condition = strings.TrimSpace(condition)

	// Debug log the evaluation attempt
	logger.Debug("Evaluating condition: %s (temp=%.1f, humidity=%.0f, pressure=%.2f)",
		condition, obs.AirTemperature, obs.RelativeHumidity, obs.StationPressure)

	// Handle compound conditions with && and ||
	if strings.Contains(condition, "&&") {
		parts := strings.Split(condition, "&&")
		for _, part := range parts {
			result, err := e.evaluateSimpleWithAlarm(strings.TrimSpace(part), obs, alarm)
			if err != nil {
				logger.Debug("Evaluation error for part '%s': %v", part, err)
				return false, err
			}
			if !result {
				logger.Debug("AND condition failed on part: %s", part)
				return false, nil
			}
		}
		logger.Debug("AND condition passed: %s", condition)
		return true, nil
	}

	if strings.Contains(condition, "||") {
		parts := strings.Split(condition, "||")
		for _, part := range parts {
			result, err := e.evaluateSimpleWithAlarm(strings.TrimSpace(part), obs, alarm)
			if err != nil {
				logger.Debug("Evaluation error for part '%s': %v", part, err)
				return false, err
			}
			if result {
				logger.Debug("OR condition passed on part: %s", part)
				return true, nil
			}
		}
		logger.Debug("OR condition failed: %s", condition)
		return false, nil
	}

	// Simple condition
	return e.evaluateSimpleWithAlarm(condition, obs, alarm)
}

// evaluateSimple evaluates a simple comparison (no logical operators)
// Kept for backward compatibility
func (e *Evaluator) evaluateSimple(condition string, obs *weather.Observation) (bool, error) {
	return e.evaluateSimpleWithAlarm(condition, obs, nil)
}

// evaluateSimpleWithAlarm evaluates a simple comparison with optional alarm state
func (e *Evaluator) evaluateSimpleWithAlarm(condition string, obs *weather.Observation, alarm *Alarm) (bool, error) {
	// Check for unary change-detection operators first
	if len(condition) > 0 {
		firstChar := condition[0]
		if firstChar == '*' || firstChar == '>' || firstChar == '<' {
			// This is a unary operator for change detection
			if alarm == nil {
				return false, fmt.Errorf("change-detection operator %c requires alarm context", firstChar)
			}
			return e.evaluateChangeDetection(condition, obs, alarm)
		}
	}

	// Parse the condition: "field operator value"
	operators := []string{">=", "<=", "!=", "==", ">", "<"}

	var field, operator, valueStr string
	for _, op := range operators {
		if idx := strings.Index(condition, op); idx > 0 {
			field = strings.TrimSpace(condition[:idx])
			operator = op
			valueStr = strings.TrimSpace(condition[idx+len(op):])
			break
		}
	}

	if operator == "" {
		return false, fmt.Errorf("invalid condition format: %s (expected 'field operator value')", condition)
	}

	// Get the field value from observation
	fieldValue, err := e.getFieldValue(field, obs)
	if err != nil {
		return false, err
	}

	// Parse the comparison value with unit conversion support
	compareValue, err := e.parseValueWithUnits(valueStr, field)
	if err != nil {
		return false, fmt.Errorf("invalid comparison value %s: %w", valueStr, err)
	}

	// Perform comparison
	return e.compare(fieldValue, operator, compareValue), nil
}

// getFieldValue extracts a field value from the weather observation
func (e *Evaluator) getFieldValue(field string, obs *weather.Observation) (float64, error) {
	field = strings.ToLower(strings.ReplaceAll(field, " ", "_"))

	switch field {
	case "temperature", "temp":
		return obs.AirTemperature, nil
	case "humidity":
		return float64(obs.RelativeHumidity), nil
	case "pressure":
		return obs.StationPressure, nil
	case "wind_speed", "wind":
		return obs.WindAvg, nil
	case "wind_gust":
		return obs.WindGust, nil
	case "wind_direction":
		return float64(obs.WindDirection), nil
	case "lux", "light":
		return obs.Illuminance, nil
	case "uv", "uv_index":
		return float64(obs.UV), nil
	case "rain_rate", "rain_accumulated":
		return obs.RainAccumulated, nil
	case "rain_daily", "rain_accumulation":
		return obs.RainAccumulated, nil
	case "lightning_count":
		return float64(obs.LightningStrikeCount), nil
	case "lightning_distance":
		return obs.LightningStrikeAvg, nil
	case "precipitation_type":
		return float64(obs.PrecipitationType), nil
	default:
		return 0, fmt.Errorf("unknown field: %s", field)
	}
}

// parseValueWithUnits parses a value string with optional unit suffix and converts to base units
// Supports:
//   - Temperature: 80F or 80f -> Celsius, 32C or 32c -> Celsius (no conversion)
//   - Wind: 25mph -> m/s, 10m/s or 10 -> m/s (no conversion)
func (e *Evaluator) parseValueWithUnits(valueStr string, field string) (float64, error) {
	valueStr = strings.TrimSpace(valueStr)
	field = strings.ToLower(field)

	// Check for temperature fields (stored in Celsius)
	if field == "temperature" || field == "temp" {
		// Check for Fahrenheit suffix
		if strings.HasSuffix(strings.ToLower(valueStr), "f") {
			valueStr = strings.TrimSuffix(strings.TrimSuffix(valueStr, "f"), "F")
			fahrenheit, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
			if err != nil {
				return 0, err
			}
			// Convert Fahrenheit to Celsius: (F - 32) * 5/9
			celsius := (fahrenheit - 32.0) * 5.0 / 9.0
			return celsius, nil
		}
		// Check for explicit Celsius suffix (optional, already in Celsius)
		if strings.HasSuffix(strings.ToLower(valueStr), "c") {
			valueStr = strings.TrimSuffix(strings.TrimSuffix(valueStr, "c"), "C")
		}
	}

	// Check for wind speed fields (stored in m/s)
	if field == "wind_speed" || field == "wind" || field == "wind_gust" {
		// Check for mph suffix
		if strings.HasSuffix(strings.ToLower(valueStr), "mph") {
			valueStr = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(valueStr, "mph"), "MPH"), "Mph")
			mph, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
			if err != nil {
				return 0, err
			}
			// Convert mph to m/s: mph * 0.44704
			ms := mph * 0.44704
			return ms, nil
		}
		// Check for explicit m/s suffix (optional, already in m/s)
		if strings.HasSuffix(strings.ToLower(valueStr), "m/s") {
			valueStr = strings.TrimSuffix(valueStr, "m/s")
			valueStr = strings.TrimSuffix(valueStr, "M/S")
		} else if strings.HasSuffix(strings.ToLower(valueStr), "ms") {
			valueStr = strings.TrimSuffix(strings.TrimSuffix(valueStr, "ms"), "MS")
		}
	}

	// Parse as plain number
	return strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
}

// evaluateChangeDetection evaluates unary change-detection operators
// Supports:
//
//	*field - triggers on ANY change in field value
//	>field - triggers when field value increases from previous
//	<field - triggers when field value decreases from previous
func (e *Evaluator) evaluateChangeDetection(condition string, obs *weather.Observation, alarm *Alarm) (bool, error) {
	if len(condition) < 2 {
		return false, fmt.Errorf("invalid change-detection condition: %s", condition)
	}

	operator := condition[0]
	fieldName := strings.TrimSpace(condition[1:])

	// Get current field value
	currentValue, err := e.getFieldValue(fieldName, obs)
	if err != nil {
		return false, err
	}

	// Get previous value
	previousValue, hasPrevious := alarm.GetPreviousValue(fieldName)

	// Store current value for next evaluation
	defer alarm.SetPreviousValue(fieldName, currentValue)

	// If no previous value exists, don't trigger (need baseline)
	if !hasPrevious {
		logger.Debug("No previous value for %s, establishing baseline: %.2f", fieldName, currentValue)
		return false, nil
	}

	// Evaluate based on operator
	switch operator {
	case '*':
		// Any change
		changed := currentValue != previousValue
		if changed {
			logger.Debug("Change detected in %s: %.2f -> %.2f", fieldName, previousValue, currentValue)
		}
		return changed, nil

	case '>':
		// Increase from previous
		increased := currentValue > previousValue
		if increased {
			logger.Debug("Increase detected in %s: %.2f -> %.2f", fieldName, previousValue, currentValue)
		}
		return increased, nil

	case '<':
		// Decrease from previous
		decreased := currentValue < previousValue
		if decreased {
			logger.Debug("Decrease detected in %s: %.2f -> %.2f", fieldName, previousValue, currentValue)
		}
		return decreased, nil

	default:
		return false, fmt.Errorf("unknown change-detection operator: %c", operator)
	}
}

// compare performs the actual comparison
func (e *Evaluator) compare(a float64, operator string, b float64) bool {
	switch operator {
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	case "==":
		return a == b
	case "!=":
		return a != b
	default:
		return false
	}
}

// GetAvailableFields returns a list of supported field names for conditions
func (e *Evaluator) GetAvailableFields() []string {
	return []string{
		"temperature", "temp",
		"humidity",
		"pressure",
		"wind_speed", "wind",
		"wind_gust",
		"wind_direction",
		"lux", "light",
		"uv", "uv_index",
		"rain_rate",
		"rain_daily",
		"lightning_count",
		"lightning_distance",
		"precipitation_type",
	}
}
