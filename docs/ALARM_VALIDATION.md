# Alarm Condition Validation Feature

## Overview
This document describes the new validation and paraphrase functionality added to the alarm editor.

## Changes Made

### 1. Fixed Save Button
**Problem:** The Save button in the alarm editor modal was not working properly.

**Solution:** Changed the button from `onclick="handleSubmit()"` to `type="submit"` to properly trigger the form submission event listener.

**Files Changed:**
- `pkg/alarm/editor/html.go` - Changed Save button type from button to submit

### 2. Added Condition Validation Button
**Feature:** Added a "✓ Validate Condition" button next to the condition textarea.

**Behavior:**
- Validates the condition syntax against the evaluator
- Shows a color-coded result message:
  - ✓ Green background for valid conditions
  - ✗ Red background for invalid conditions
  - ⚠ Yellow background for warnings
- Displays a human-readable paraphrase of valid conditions

**Files Changed:**
- `pkg/alarm/editor/html.go` - Added validate button and result div
- `pkg/alarm/editor/static/script.js` - Added `validateCondition()` function

### 3. Automatic Validation on Save
**Feature:** Conditions are automatically validated when the Save button is clicked.

**Behavior:**
- Validates the condition before submitting the form
- Prevents saving if the condition is invalid
- Shows an error notification if validation fails
- Only proceeds with save if validation succeeds

**Files Changed:**
- `pkg/alarm/editor/static/script.js` - Updated `handleSubmit()` to call `validateCondition()`

### 4. Condition Paraphrase Engine
**Feature:** Added intelligent paraphrasing of alarm conditions into human-readable text.

**Examples:**
- `temperature > 85F` → "When temperature exceeds 85°F"
- `humidity >= 80` → "When humidity is at least 80"
- `*lightning_count` → "When lightning strike count changes (any value)"
- `>rain_rate` → "When rain rate increases"
- `<lightning_distance` → "When lightning distance decreases"
- `temperature > 85F && humidity > 80` → "When temperature exceeds 85°F AND humidity exceeds 80"
- `lux > 50000 || uv > 8` → "When light level exceeds 50000 OR UV index exceeds 8"

**Features:**
- Converts field names to human-readable format (e.g., `wind_speed` → "wind speed")
- Handles all comparison operators (>, <, >=, <=, ==, !=)
- Supports change detection operators (*, >, <)
- Formats units properly (°F, °C, mph, m/s)
- Handles compound conditions (AND, OR)
- Works with all weather fields

**Files Changed:**
- `pkg/alarm/evaluator.go` - Added `Paraphrase()`, `paraphraseSimple()`, `formatFieldName()`, and `formatValue()` methods

### 5. Updated Validation API Endpoint
**Feature:** Enhanced the `/api/validate` endpoint to return condition paraphrase.

**API Response:**
```json
{
  "valid": true,
  "paraphrase": "When temperature exceeds 85°F"
}
```

Or for invalid conditions:
```json
{
  "valid": false,
  "error": "unknown field: fake_field"
}
```

**Files Changed:**
- `pkg/alarm/editor/server.go` - Updated `handleValidate()` to include paraphrase

### 6. Comprehensive Test Coverage
**Feature:** Added comprehensive tests for the new paraphrase functionality.

**Test Coverage:**
- Simple comparisons with all operators
- Temperature with °F and °C units
- Wind speed with mph and m/s units
- Change detection operators (*, >, <)
- Compound conditions (AND, OR)
- Edge cases (empty conditions, unknown fields)
- Field name formatting
- Value formatting

**Files Added:**
- `pkg/alarm/evaluator_paraphrase_test.go` - 18 test cases covering all scenarios

**Files Updated:**
- `pkg/alarm/editor/server_test.go` - Updated to verify paraphrase in validate endpoint

## Test Results

All tests pass successfully:
```
go test -cover ./pkg/alarm/
ok      tempest-homekit-go/pkg/alarm    5.291s  coverage: 68.4% of statements
```

Coverage increased from 66.5% to 68.4% (+1.9%).

## User Experience

### Before:
1. User enters condition
2. Clicks Save
3. **Save button doesn't work** (bug)
4. No feedback on whether condition is valid
5. Only discovers syntax errors after saving to file

### After:
1. User enters condition
2. User clicks "✓ Validate Condition" button (optional)
   - Sees immediate validation result
   - Sees human-readable explanation of what the condition means
3. User clicks Save
   - **Save button now works properly** (fixed)
   - Condition is automatically validated
   - Clear error message if condition is invalid
   - Only saves if condition is valid

## Examples

### Example 1: Simple Temperature Alert
**Condition:** `temperature > 85F`
**Paraphrase:** "When temperature exceeds 85°F"

### Example 2: Heat Index Alert
**Condition:** `temperature > 35C && humidity > 80`
**Paraphrase:** "When temperature exceeds 35°C AND humidity exceeds 80"

### Example 3: Lightning Detection
**Condition:** `*lightning_count`
**Paraphrase:** "When lightning strike count changes (any value)"

### Example 4: Lightning Proximity Alert
**Condition:** `<lightning_distance`
**Paraphrase:** "When lightning distance decreases"

### Example 5: Bright or High UV
**Condition:** `lux > 50000 || uv > 8`
**Paraphrase:** "When light level exceeds 50000 OR UV index exceeds 8"

## Technical Details

### Paraphrase Algorithm
1. Parse condition into parts (splitting on && or ||)
2. For each part:
   - Detect unary operators (*, >, <) for change detection
   - Detect binary operators (>=, <=, !=, ==, >, <) for comparisons
   - Extract field name and value
   - Format field name (e.g., "wind_speed" → "wind speed")
   - Format value with units (e.g., "85F" → "85°F")
   - Convert operator to human text (e.g., ">" → "exceeds")
3. Combine parts with AND/OR connectors

### Supported Fields
- temperature/temp
- humidity
- pressure
- wind_speed/wind
- wind_gust
- wind_direction
- lux/light
- uv/uv_index
- rain_rate
- rain_daily
- lightning_count
- lightning_distance
- precipitation_type

### Supported Operators
- `>` - exceeds
- `<` - is below
- `>=` - is at least
- `<=` - is at most
- `==` - is
- `!=` - is not
- `*field` - changes (any value)
- `>field` - increases
- `<field` - decreases
- `&&` - AND
- `||` - OR

### Supported Units
- Temperature: F, C (displayed as °F, °C)
- Wind Speed: mph, m/s (displayed with space)
- All other values: displayed as-is

## Future Enhancements

Potential improvements:
1. Add validation suggestions (e.g., "Did you mean 'temperature' instead of 'temp'?")
2. Add real-time validation as user types (with debounce)
3. Add syntax highlighting in condition textarea
4. Add autocomplete for field names
5. Add visual condition builder (drag-and-drop)
6. Add more detailed error messages with line/column numbers
7. Add validation against current weather data to show if condition would trigger now
