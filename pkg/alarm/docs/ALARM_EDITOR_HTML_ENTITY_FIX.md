# Alarm Editor Template Fix - HTML Entity Escaping

## Issue
The alarm editor page was not loading, showing only a blank page with:
```html
<html><head><style>...</style></head><body></body></html>
```

Browser console showed the HTML was essentially empty, and the server was returning `500 Internal Server Error`.

## Root Cause
The HTML template contained unescaped `<` and `>` characters in the help text for the condition field:

```html
<small>...Change detection: *field (any change), >field (increase), <field (decrease)...</small>
```

The Go `html/template` parser interpreted `<field` as an HTML tag, causing a template parsing error:
```
html/template:index: "<" in attribute name: "closer)</small>\n"
```

## Solution
Escaped all `<` and `>` characters in the help text using HTML entities:
- `<` → `&lt;`
- `>` → `&gt;`

### Changed Line
**File:** `pkg/alarm/editor/html.go`

**Before:**
```html
<small>Click sensor names above to insert into condition. Supports units: 80F or 26.7C (temp), 25mph or 11.2m/s (wind). Change detection: *field (any change), >field (increase), <field (decrease). Examples: temperature > 85F, *lightning_count (any strike), >rain_rate (rain increasing), <lightning_distance (lightning closer)</small>
```

**After:**
```html
<small>Click sensor names above to insert into condition. Supports units: 80F or 26.7C (temp), 25mph or 11.2m/s (wind). Change detection: *field (any change), &gt;field (increase), &lt;field (decrease). Examples: temperature &gt; 85F, *lightning_count (any strike), &gt;rain_rate (rain increasing), &lt;lightning_distance (lightning closer)</small>
```

## Why This Happened
When we added the change-detection operator documentation, we included examples like `>field` and `<field` in the help text. These characters have special meaning in HTML:
- `<` starts an HTML tag
- `>` ends an HTML tag

The Go `html/template` package's strict parsing detected this as invalid HTML structure.

## Verification

### Before Fix
```bash
$ curl -I http://localhost:8081/
HTTP/1.1 500 Internal Server Error
Content-Length: 22

$ cat /tmp/alarm-editor.log
ERROR: Failed to execute template: html/template:index: "<" in attribute name: "closer)</small>
```

### After Fix
```bash
$ curl -I http://localhost:8081/
HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8

$ curl -s http://localhost:8081/ | wc -l
148

$ curl -I http://localhost:8081/alarm-editor/static/styles.css
HTTP/1.1 200 OK
Content-Type: text/css
Content-Length: 9547

$ curl -I http://localhost:8081/alarm-editor/static/script.js
HTTP/1.1 200 OK
Content-Type: application/javascript
Content-Length: 14424
```

Page loads correctly
CSS loads correctly (9.5 KB)
JavaScript loads correctly (14.4 KB)
All static assets accessible

## Browser Display
The help text now correctly displays as:
```
Change detection: *field (any change), >field (increase), <field (decrease)
```

The HTML entities are automatically rendered as the actual characters by the browser, so users see the intended symbols.

## Additional Fix
Also added error handling to the `handleIndex` function to log template execution errors:

```go
if err := tmpl.Execute(w, data); err != nil {
 logger.Error("Failed to execute template: %v", err)
 http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
```

This ensures future template errors are logged and don't silently fail.

## Testing
To test the alarm editor:
```bash
# Build the project
go build

# Run the alarm editor
./tempest-homekit-go --alarms-edit tempest-alarms.json

# Open in browser
open http://localhost:8081
```

Expected result: - Page loads with full UI
- Help text displays correctly: `>field (increase), <field (decrease)`
- All styling and JavaScript functionality works

## Lessons Learned

1. **HTML Entity Escaping Required**
 - Always escape `<`, `>`, `&`, `"`, and `'` in HTML content
 - Use `&lt;`, `&gt;`, `&amp;`, `&quot;`, `&apos;` respectively

2. **Template Error Handling**
 - Always check template execution errors
 - Log errors to help with debugging
 - Return proper HTTP error codes

3. **Test After Documentation Changes**
 - Documentation changes can introduce syntax errors
 - Test the actual rendered page, not just the build
 - Use curl to verify HTTP responses

4. **Go Template Strictness**
 - `html/template` is strict about HTML syntax
 - Provides security against XSS attacks
 - Prevents accidental HTML injection

## Related Files
- `pkg/alarm/editor/html.go` - Fixed HTML template with entity escaping
- `pkg/alarm/editor/server.go` - Added error handling to handleIndex
- `CHANGE_DETECTION_OPERATORS.md` - Documentation for the operators

## Status
**FIXED** - Alarm editor now loads correctly with properly escaped HTML entities in the help text.
