# Alarm Editor Message Configuration - Implementation Summary

## Overview

Enhanced the alarm editor with comprehensive message configuration capabilities, allowing users to customize notification messages for each delivery method with an easy-to-use variable dropdown selector.

## Changes Made

### 1. HTML Template (pkg/alarm/editor/html.go)

**Added**:
- Default message section with checkbox toggle
- Variable dropdown selector (17 available variables)
- Custom message sections for each delivery method:
 - Console message with variable dropdown
 - Syslog message with variable dropdown
 - Event Log message with variable dropdown
 - Email configuration (To, Subject, Body) with variable dropdown
 - SMS configuration (Phone numbers, Message) with variable dropdown
- Dynamic show/hide based on selected delivery methods
- `onchange` handlers for checkbox interactions

**Key Features**:
- Default message template: `ALARM: {{alarm_name}}\nStation: {{station}}\nTime: {{timestamp}}`
- All variable dropdowns include helpful descriptions (e.g., "{{temperature_f}} - Temperature °F")
- Email section has separate fields for recipients, subject, and body
- SMS section has separate fields for phone numbers and message

### 2. JavaScript (pkg/alarm/editor/static/script.js)

**New Functions**:

1. `insertVariable(textareaId, alternateId)`
 - Inserts selected variable at cursor position
 - Handles email dual-target (subject or body based on focus)
 - Resets dropdown after insertion
 - Maintains cursor position

2. `toggleMessageSections()`
 - Shows/hides custom message sections based on delivery method checkboxes
 - Called on delivery method change
 - Dynamic UI responsiveness

3. `toggleCustomMessages()`
 - Toggles between default message and custom messages per channel
 - Shows/hides appropriate sections
 - Called on "Use Default Message" checkbox change

**Updated Functions**:

1. `showCreateModal()`
 - Resets all message fields to defaults
 - Sets default message template
 - Clears custom message fields
 - Calls toggle functions to set initial UI state

2. `editAlarm(name)`
 - Loads messages from existing channel configurations
 - Populates console, syslog, eventlog, email, SMS fields
 - Detects if custom messages are used (different per channel)
 - Auto-selects default vs custom message mode
 - Preserves email recipients and SMS phone numbers

3. `handleSubmit(e)`
 - Reads message configuration based on default/custom toggle
 - Builds channels array with appropriate templates
 - Uses default message for all channels when "Use Default" is checked
 - Uses individual messages when custom mode is enabled
 - Parses comma-separated email addresses and phone numbers
 - Provides fallback values for email/SMS if not specified

### 3. CSS Styling (pkg/alarm/editor/static/styles.css)

**New Styles**:

- `.message-section` - Container for message configuration
- `.message-input-section` - Individual message input areas
- `.message-header` - Header with label and variable dropdown
- `.variable-dropdown` - Styled dropdown for variable selection
- `#customMessageSections` - Container for per-channel messages
- Input and textarea focus states with purple accent (`#667eea`)
- Responsive layouts for message sections

**Visual Design**:
- Light gray background for message section container
- White background for individual input areas
- Border styling consistent with form theme
- Hover effects on variable dropdown
- Focus states with purple accent and subtle shadow

### 4. Documentation

**Created**:
- `ALARM_EDITOR_MESSAGES.md` - Comprehensive 500+ line documentation
 - Feature overview
 - All 17 available variables with descriptions
 - User interface workflow
 - 4 detailed examples with actual use cases
 - Technical details on storage and expansion
 - Best practices for message length, formatting, variable usage
 - Troubleshooting guide
 - Related documentation links

**Updated**:
- `README.md` - Added reference to new documentation in "Additional Documentation" section

## Available Template Variables

### Alarm & Station Metadata
- `{{alarm_name}}` - Name of the triggered alarm
- `{{station}}` - Weather station name
- `{{timestamp}}` - Current time (YYYY-MM-DD HH:MM:SS TZ)

### Temperature
- `{{temperature}}` - Temperature in Celsius
- `{{temperature_f}}` - Temperature in Fahrenheit
- `{{temperature_c}}` - Temperature in Celsius (explicit)

### Atmospheric
- `{{humidity}}` - Relative humidity percentage
- `{{pressure}}` - Station pressure in millibars

### Wind
- `{{wind_speed}}` - Average wind speed (m/s)
- `{{wind_gust}}` - Peak wind gust (m/s)
- `{{wind_direction}}` - Wind direction in degrees

### Light & UV
- `{{lux}}` - Light intensity in lux
- `{{uv}}` - UV index

### Precipitation
- `{{rain_rate}}` - Current rain rate (mm)
- `{{rain_daily}}` - Daily rain accumulation (mm)

### Lightning
- `{{lightning_count}}` - Number of lightning strikes
- `{{lightning_distance}}` - Average lightning strike distance (km)

## User Experience Flow

### Creating New Alarm
1. Click "New Alarm"
2. Enter alarm details (name, condition)
3. Select delivery methods (console, syslog, email, SMS, etc.)
4. Choose message strategy:
 - **Option A**: Leave "Use Default Message" checked → Single message for all channels
 - **Option B**: Uncheck "Use Default Message" → Customize each channel independently
5. Use variable dropdown to insert weather/alarm data
6. For email: Enter recipients, subject, body
7. For SMS: Enter phone numbers, keep message short
8. Save alarm

### Editing Existing Alarm
1. Click "Edit" on alarm card
2. Existing messages load into appropriate fields
3. If all channels have same message → Default mode
4. If channels differ → Custom mode
5. Modify messages as needed
6. Variable dropdown available for easy insertion
7. Save changes

### Using Variable Dropdown
1. Click in message textarea
2. Click variable dropdown
3. Select variable from categorized list
4. Variable inserted at cursor position
5. Continue typing or select more variables

## Technical Implementation

### Message Storage Format

**Simple Channels (Console/Syslog/EventLog)**:
```json
{
 "type": "console",
 "template": "ALARM: {{alarm_name}}\nTemp: {{temperature_f}}°F"
}
```

**Email Channel**:
```json
{
 "type": "email",
 "email": {
 "to": ["user@example.com", "admin@example.com"],
 "subject": "Warning: Weather Alert: {{alarm_name}}",
 "body": "Alert at {{timestamp}}\nTemp: {{temperature_f}}°F\nHumidity: {{humidity}}%"
 }
}
```

**SMS Channel**:
```json
{
 "type": "sms",
 "sms": {
 "to": ["+15551234567", "+15559876543"],
 "message": "{{alarm_name}}: {{temperature_f}}°F @ {{timestamp}}"
 }
}
```

### Variable Expansion

Variables are expanded at runtime by `expandTemplate()` function in `pkg/alarm/notifiers.go`:
- Reads current observation data
- Replaces all `{{variable}}` placeholders with actual values
- Formats numbers appropriately (temperature to 1 decimal, humidity to integer, etc.)
- Ensures timestamp is formatted as readable string

### Backward Compatibility

Existing alarms without custom messages continue to work:
- Default template is generated if none exists
- Loading old alarms shows default message mode
- Saving preserves existing message structure

## Testing Checklist

Build successful (no compilation errors)
HTML structure valid
JavaScript functions defined
CSS styling applied
Documentation complete

### Manual Testing Required
- [ ] Create new alarm with default message
- [ ] Create new alarm with custom messages per channel
- [ ] Edit existing alarm and verify message loading
- [ ] Test variable dropdown insertion in all textareas
- [ ] Verify email with multiple recipients
- [ ] Verify SMS with multiple phone numbers
- [ ] Test message save/load cycle
- [ ] Verify alarm triggering uses correct message
- [ ] Check mobile responsiveness

## Benefits

1. **User-Friendly**: Variable dropdown eliminates need to memorize template syntax
2. **Flexible**: Support both simple (default) and advanced (custom) message configurations
3. **Consistent**: Same variable system across all delivery methods
4. **Maintainable**: Clear separation between default and custom messages
5. **Professional**: Comprehensive documentation with examples
6. **Reliable**: Backward compatible with existing alarm configurations

## Future Enhancements (Possible)

- [ ] Message preview with sample data substitution
- [ ] Message templates library (predefined common messages)
- [ ] HTML email support (currently plain text only)
- [ ] Character counter for SMS messages
- [ ] Copy message between channels
- [ ] Import/export message templates
- [ ] Conditional message content based on sensor values
- [ ] Rich text editor for email bodies

## Files Modified

1. `/pkg/alarm/editor/html.go` - Added message configuration UI
2. `/pkg/alarm/editor/static/script.js` - Added message handling logic
3. `/pkg/alarm/editor/static/styles.css` - Added message section styling
4. `/README.md` - Added documentation reference
5. `/ALARM_EDITOR_MESSAGES.md` - Created comprehensive documentation

## Version Information

- **Implementation Date**: October 10, 2025
- **Go Version**: 1.24.2+
- **Feature Version**: 1.0
- **Status**: Complete and ready for testing
