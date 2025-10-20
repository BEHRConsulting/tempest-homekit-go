# Change Detection Visual Guide

## Operator Symbols

```
*field → Any Change (asterisk = wildcard = any)
>field → Increase (arrow pointing up)
<field → Decrease (arrow pointing down)
```

## How It Works - Visual Timeline

### Example 1: Lightning Detection (`*lightning_count`)

```
Time Value Previous Result Reason
---- ----- -------- ------ ------
T1 0 none  NO Establishing baseline
T2 0 0  NO No change (0 → 0)
T3 1 0 YES Changed! (0 → 1)
T4 1 1  NO No change (1 → 1)
T5 3 1 YES Changed! (1 → 3)
T6 3 3  NO No change (3 → 3)
T7 2 3 YES Changed! (3 → 2)
```

### Example 2: Rain Increasing (`>rain_rate`)

```
Time Value Previous Result Reason
---- ----- -------- ------ ------
T1 0.0 none  NO Establishing baseline
T2 0.0 0.0  NO No change (0.0 → 0.0)
T3 0.5 0.0 YES Increased! (0.0 → 0.5)
T4 2.0 0.5 YES Increased! (0.5 → 2.0)
T5 2.0 2.0  NO No change (2.0 → 2.0)
T6 1.0 2.0  NO Decreased (2.0 → 1.0) ← NOT increase
T7 3.0 1.0 YES Increased! (1.0 → 3.0)
```

### Example 3: Lightning Getting Closer (`<lightning_distance`)

```
Time Value Previous Result Reason
---- ----- -------- ------ ------
T1 50 none  NO Establishing baseline
T2 50 50  NO No change (50 → 50)
T3 30 50 YES Decreased! (50 → 30) ← Closer!
T4 10 30 YES Decreased! (30 → 10) ← Much closer!
T5 2 10 YES Decreased! (10 → 2) ← Very close!
T6 5 2  NO Increased (2 → 5) ← NOT decrease
T7 3 5 YES Decreased! (5 → 3) ← Closer again
```

## Compound Conditions - Visual Logic

### AND Example: `*lightning_count && lightning_distance < 10`

```
Condition 1: *lightning_count (any change)
Condition 2: lightning_distance < 10 (threshold)

Time Strike Count Distance C1 C2 Result
---- ------------ -------- -- -- ------
T1 0 50    NO (baseline + far)
T2 1 50   NO (strike but too far)
T3 1 50    NO (no change + far)
T4 2 8 YES (strike AND close!)
```

### OR Example: `>rain_rate || >wind_gust`

```
Condition 1: >rain_rate (rain increasing)
Condition 2: >wind_gust (wind increasing)

Time Rain Wind C1 C2 Result
---- ---- ---- -- -- ------
T1 0.0 10.0    NO (baseline)
T2 1.0 10.0  YES (rain up)
T3 1.0 15.0  YES (wind up)
T4 2.0 18.0 YES (both up!)
T5 2.0 18.0    NO (no change)
```

## Real-World Scenarios - Visual

### Scenario 1: Storm Approach

```
️ T0 (Baseline)
 pressure: 1015 mb
 wind: 5 m/s
 rain: 0 mm/hr
 ↓
️ T1 (No Alert Yet)
 pressure: 1015 mb → no change
 wind: 5 m/s → no change
 rain: 0 mm/hr → no change
 ↓
 T2 (Condition: <pressure && >wind_speed)
 pressure: 1010 mb (falling)
 wind: 8 m/s (rising)
 rain: 0 mm/hr
  Alert: ALERT: "Storm approaching!"
 ↓
️ T3 (Condition: >rain_rate)
 pressure: 1005 mb (still falling)
 wind: 12 m/s (still rising)
 rain: 2.0 mm/hr (started!)
  Alert: ALERT: "Rain starting!"
```

### Scenario 2: Lightning Safety

```
️ T0 (Clear)
 lightning_count: 0
 lightning_distance: --
 ↓
 T1 (Condition: *lightning_count)
 lightning_count: 1 (changed from 0)
 lightning_distance: 25 km
  Alert: ALERT: "Lightning detected at 25km"
 ↓
 T2 (No alert - same count)
 lightning_count: 1 (no change)
 lightning_distance: 25 km
 ↓
 T3 (Condition: <lightning_distance && lightning_distance < 20)
 lightning_count: 1
 lightning_distance: 15 km (closer! and < 20)
  Alert: ALERT: "Lightning approaching - now 15km!"
 ↓
 T4 (Condition: *lightning_count)
 lightning_count: 3 (changed from 1)
 lightning_distance: 10 km
  Alert: ALERT: "More lightning - now 3 strikes at 10km!"
```

## Operator Decision Tree

```
 Start
 |
 ↓
 Does value need tracking?
 / \
 NO YES
 | |
 ↓ ↓
 Use regular Use change
 comparison detection
 (temp > 30) |
 ↓
 What type of change?
 / | \
 Any Increase Decrease
 | | |
 ↓ ↓ ↓
 *field >field <field
 | | |
 ↓ ↓ ↓
 Triggers on Triggers Triggers
 ANY change when UP when DOWN
```

## State Management - Visual

```
Alarm Object
┌─────────────────────────────────────┐
│ Name: "Lightning Monitor" │
│ Condition: "*lightning_count" │
│ Enabled: true │
│ Cooldown: 60s │
│ ┌─────────────────────────────────┐ │
│ │ previousValue (map) │ │
│ │ ┌───────────────────────────┐ │ │
│ │ │ "lightning_count" → 1.0 │ │ │
│ │ │ "rain_rate" → 0.5 │ │ │
│ │ │ "wind_gust" → 10.0 │ │ │
│ │ └───────────────────────────┘ │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘

Each field tracked independently!
```

## Performance Flow

```
New Weather Observation Received
 ↓
 For Each Alarm:
 ↓
 Parse Condition
 ↓
 Unary Operator? ───NO───→ Regular Comparison
 │ ↓
 YES Return Result
 ↓
 Get Current Value
 ↓
 Get Previous Value (map lookup) ← O(1) operation
 ↓
 Compare Values
 (*, >, <)
 ↓
 Store Current as Previous ← O(1) operation
 ↓
 Return Result
```

## Memory Layout

```
Single Alarm with Change Detection:
┌────────────────────────────────┐
│ Alarm struct: ~200 bytes │
│ - name, condition, etc. │
│ │
│ previousValue map: │
│ - 1 field: ~40 bytes │
│ - 2 fields: ~80 bytes │
│ - 3 fields: ~120 bytes │
└────────────────────────────────┘

Typical overhead: 40-120 bytes per alarm
```

## Legend

```
YES - Condition triggered, alarm fires
 NO - Condition not met, no alarm
→ - Value transition (before → after)
Alert: - Alert/notification sent
 - Lightning event
️ - Rain event
 - Falling value
 - Rising value
```

## Quick Operator Reference Card

```
┌─────────────────────────────────────────────────────────┐
│ CHANGE DETECTION OPERATORS │
├─────────────────────────────────────────────────────────┤
│ │
│ *field ANY CHANGE Example: *lightning_count │
│ ═══════════ │
│ Triggers on any value change │
│ Use for: Events, strikes, state changes │
│ │
│ >field INCREASE Example: >rain_rate │
│ ═══════════ │
│ Triggers when value goes UP │
│ Use for: Intensifying conditions │
│ │
│ <field DECREASE Example: <lightning_distance │
│ ═══════════ │
│ Triggers when value goes DOWN │
│ Use for: Approaching threats │
│ │
├─────────────────────────────────────────────────────────┤
│ COMBINE WITH: │
│ • Thresholds: >rain_rate && rain_rate > 5 │
│ • AND logic: *lightning_count && distance < 10 │
│ • OR logic: >rain_rate || >wind_gust │
└─────────────────────────────────────────────────────────┘
```
