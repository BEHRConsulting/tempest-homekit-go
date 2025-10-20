# Alarm Editor Display Fix

## Issue
The alarm editor page was not displaying after the CSS/JS extraction to static files. No errors appeared in the console, but the page was blank.

## Root Cause
During the CSS and JavaScript extraction to external files, the old inline code was not completely removed from `pkg/alarm/editor/html.go`. Instead, it was left in place with invalid `style="display:none;"` attributes that didn't actually hide the content. This caused the browser to:

1. Try to load external stylesheets/scripts
2. Also parse the remaining inline CSS/JS
3. Encounter conflicts or malformed HTML
4. Fail to render the page properly

## Solution
Completely removed all inline CSS and JavaScript content from the HTML template, keeping only the external file references.

### Files Modified
- **pkg/alarm/editor/html.go** - Reduced from 1135 lines to 151 lines

### Changes Made

**Before (broken):**
```html
<head>
 <title>Alarm Configuration Editor</title>
 <link rel="stylesheet" href="/alarm-editor/static/styles.css">
 <style type="text/css" style="display:none;">
 /* 580+ lines of CSS still embedded */
 </style>
</head>
<body>
 <!-- HTML content -->
 <script src="/alarm-editor/static/script.js"></script>
 <script style="display:none;">
 /* 400+ lines of JavaScript still embedded */
 </script>
</body>
```

**After (fixed):**
```html
<head>
 <title>Alarm Configuration Editor</title>
 <link rel="stylesheet" href="/alarm-editor/static/styles.css">
</head>
<body>
 <!-- HTML content -->
 <script src="/alarm-editor/static/script.js"></script>
</body>
```

## Verification

### File Sizes
- `html.go`: 1135 lines → **151 lines** (984 lines removed)
- `static/styles.css`: **9.3 KB** (properly formatted)
- `static/script.js`: **14 KB** (properly formatted)

### Build Status
Build successful
Tests passing
Static files verified

### Static File Serving
The server correctly serves static files from:
- URL path: `/alarm-editor/static/*`
- Physical path: `./pkg/alarm/editor/static/*`
- Content-Type headers: Set correctly for `.css` and `.js` files
- Cache-Control: `no-cache` for development

## How It Works Now

1. Browser loads `http://localhost:8081/`
2. Server serves HTML from `html.go` template
3. Browser requests `/alarm-editor/static/styles.css`
4. Server serves file from `./pkg/alarm/editor/static/styles.css`
5. Browser requests `/alarm-editor/static/script.js`
6. Server serves file from `./pkg/alarm/editor/static/script.js`
7. Page renders correctly with all styles and functionality

## Testing

To test the alarm editor:
```bash
# Build the project
go build

# Run the alarm editor
./tempest-homekit-go --alarm-editor config/alarms.json --port 8081

# Open in browser
open http://localhost:8081
```

Expected result: Page displays correctly with all styling and functionality.

## What Was Fixed

Removed all inline CSS (580+ lines)
Removed all inline JavaScript (400+ lines)
Kept only external file references
Clean HTML structure (151 lines)
Proper separation of concerns
Build successful
Tests passing

## Related Files

- `pkg/alarm/editor/html.go` - HTML template (cleaned)
- `pkg/alarm/editor/static/styles.css` - All CSS (9.3 KB)
- `pkg/alarm/editor/static/script.js` - All JavaScript (14 KB)
- `pkg/alarm/editor/server.go` - Static file serving handler

## Lessons Learned

1. **Complete removal is better than hiding** - Using `style="display:none;"` is invalid HTML and doesn't actually hide embedded code from the parser
2. **Verify extraction** - When extracting code to external files, verify all inline content is removed
3. **File size as indicator** - Large HTML templates after "extraction" indicate incomplete removal
4. **Test the UI** - Build and visual tests are essential after structural changes

## Future Prevention

When extracting inline code to external files:
1. Extract content to external file
2. **Completely delete** inline content (not just hide it)
3. Add external file reference
4. Verify file sizes (should decrease significantly)
5. Test in browser before committing
