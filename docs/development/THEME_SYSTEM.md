# Theme System Implementation

## Overview
Implemented a comprehensive theme system for the Tempest HomeKit web console with 7 different visual styles that users can switch between via a dropdown in the footer.

## Themes Available

### 1. **Default (Purple Gradient)** - Original design with purple gradient background
- High contrast, modern look
- Colors: Purple/violet gradient (#667eea → #764ba2)

### 2. **Ocean Blue**
- Calming blue tones inspired by the ocean
- Professional and clean
- Colors: Deep blue to cyan (#2E3192 → #1BFFFF)

### 3. **Sunset Orange**
- Warm, vibrant sunset colors
- Energetic and eye-catching
- Colors: Pink to red (#FA8BFF → #FD1D1D)

### 4. **Forest Green**
- Natural green tones inspired by forests
- Easy on the eyes, nature-themed
- Colors: Teal to sage green (#134E5E → #71B280)

### 5. **Midnight Dark**
- Dark mode with deep blues
- Reduces eye strain in low light
- Dark cards (#1a2a3a) with light text (#e0e0e0)
- Special chart color adjustments for readability

### 6. **Arctic Light**
- Cool, light theme with soft blues
- Minimalist and clean
- Colors: Light blue gradient (#E0EAFC → #CFDEF3)
- Dark text for high contrast

### 7. **Autumn Earth**
- Warm earth tones inspired by autumn
- Comfortable, natural palette
- Colors: Beige to sage (#DAD299 → #B0DAB9)

## Technical Implementation

### Files Created/Modified

1. **`pkg/web/static/themes.css`** (NEW)
 - Contains all theme definitions using CSS variables
 - 7 theme variations using `[data-theme="name"]` selectors
 - Theme selector styling

2. **`pkg/web/static/styles.css`** (MODIFIED)
 - Added CSS variables for themeable properties
 - Updated all color references to use variables
 - Maintains backward compatibility

3. **`pkg/web/server.go`** (MODIFIED)
 - Added theme selector dropdown in footer
 - Added themes.css to static file serving
 - Added themes.css link in HTML head

4. **`pkg/web/static/script.js`** (MODIFIED)
 - Theme switching JavaScript (70+ lines)
 - Saves theme preference to localStorage
 - Updates chart colors for dark themes
 - Applies theme on page load

5. **`pkg/web/static/chart.html`** (MODIFIED)
 - Added theme support for popout charts
 - Syncs theme with main dashboard
 - Updates chart grid/text colors for dark mode

## Features

### User Experience
-  Theme dropdown in footer with descriptive names
-  Theme preference saved to browser localStorage
-  Instant theme switching without page reload
-  Chart colors automatically adjust for dark themes
- 🪟 Popout charts inherit theme from main dashboard

### Technical Features
- CSS Variables for easy theme management
- No inline styles - all theming via external CSS
- Dark mode chart adjustments (grid & text colors)
- Backward compatible with existing code
- Clean separation of concerns

## CSS Variables Used

Each theme defines these variables:
```css
--bg-gradient-start /* Background gradient start color */
--bg-gradient-end /* Background gradient end color */
--card-bg /* Card background */
--card-text /* Main text color */
--card-text-light /* Secondary text color */
--card-title /* Card title color */
--header-text /* Header text color */
--footer-text /* Footer text color */
--status-bg /* Status banner background */
--status-text /* Status banner text */
--shadow-color /* Card shadow color */
--shadow-hover /* Card hover shadow */
--link-color /* Link and accent color */
--chart-grid /* Chart grid line color */
```

## Usage

1. Open web dashboard at `http://localhost:8080`
2. Scroll to footer
3. Click " Theme:" dropdown
4. Select desired theme
5. Theme applies instantly and is saved

## Dark Mode Support

The **Midnight** theme includes special handling:
- Light text on dark cards
- Adjusted chart grid colors (white with low opacity)
- Modified tick label colors for readability
- Darker shadows for depth

## Future Enhancements

Possible additions:
- Custom theme creator
- Import/export custom themes
- Per-card theme overrides
- System preference detection (prefers-color-scheme)
- Theme preview thumbnails
- More theme variations

## Testing Checklist

- [x] All 7 themes apply correctly
- [x] Theme persists after page reload
- [x] Popout charts use same theme
- [x] Dark mode charts readable
- [x] Dropdown styled properly in all themes
- [x] No console errors
- [x] Build successful
- [x] Works in main dashboard
- [x] Works in popout charts

## Code Quality

- No hardcoded colors (all use variables)
- Clean separation (themes.css separate from styles.css)
- Maintainable (add new themes easily)
- Performant (CSS variables, no JavaScript color manipulation except charts)
- Accessible (good contrast ratios in all themes)
