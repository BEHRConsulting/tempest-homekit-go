# web/static/ Directory

This directory contains all frontend assets for the Tempest HomeKit Go web dashboard. The assets provide a modern, interactive web interface for real-time weather monitoring.

## Files

### `script.js`
**Main JavaScript Application (~800+ lines)**

**Core Functionality:**
- **Real-time Data Updates**: Fetches weather data every 10 seconds via REST API
- **Interactive Dashboard**: Updates all weather cards with live data
- **Unit Conversion System**: Toggle between metric and imperial units
- **Chart Integration**: Chart.js for historical data visualization
- **Tooltip Management**: Interactive information tooltips for all sensors
- **Event Handling**: Proper click handling and event propagation
- **Local Storage**: Persistent user preferences for units

**Key Features:**
```javascript
// Real-time data fetching
function fetchWeatherData() {
    fetch('/api/weather')
        .then(response => response.json())
        .then(data => updateDisplay(data));
}

// Unit conversion system
function toggleUnit(sensorType) {
    // Toggle between metric/imperial
    // Save preference to localStorage
    // Update display immediately
}

// Chart management
function updateCharts(weatherData) {
    // Update Chart.js charts with new data
    // Handle historical data points
    // Maintain chart performance
}
```

**Tooltip System:**
- Information tooltips for all weather sensors
- Interactive reference tables (lux, pressure, UV index, rain intensity)
- Consistent positioning and styling
- Click-outside-to-close functionality

### `styles.css`
**Modern CSS Styling**

**Design Features:**
- **Responsive Grid**: Adapts from 3-column (desktop) to 1-column (mobile)
- **Weather Theme**: Blue gradients and weather-appropriate colors
- **Card-based Layout**: Hover effects and smooth transitions
- **Modern Typography**: System font stack for optimal readability
- **Dark Tooltips**: Professional tooltip styling with contrast
- **Mobile Optimization**: Touch-friendly interface elements

**Key Styling:**
```css
/* Responsive grid system */
.weather-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 20px;
}

/* Weather card styling */
.weather-card {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 15px;
    padding: 20px;
    transition: transform 0.3s ease;
}

/* Tooltip system */
.tooltip {
    position: absolute;
    background: rgba(0, 0, 0, 0.9);
    color: white;
    border-radius: 6px;
    z-index: 1000;
}
```

### `date-fns.min.js`
**Date Manipulation Library (v2.30.0)**

**Purpose:** External library for Chart.js date/time handling
- **Chart Axes**: Formats time-based chart axes
- **Date Parsing**: Handles weather data timestamps
- **Timezone Support**: Converts UTC timestamps to local time
- **Chart.js Integration**: Powers the date adapter for time-series charts

**Usage in Charts:**
```javascript
// Chart.js configuration with date-fns adapter
const chartConfig = {
    scales: {
        x: {
            type: 'time',
            adapters: {
                date: {
                    library: dateFns
                }
            }
        }
    }
};
```

## Frontend Architecture

### External JavaScript Pattern
The dashboard follows a clean separation of concerns:
- **HTML**: Semantic markup without inline scripts
- **CSS**: All styling in external stylesheet
- **JavaScript**: Complete functionality in external script
- **No Framework**: Vanilla JavaScript for simplicity and performance

### Chart.js Integration
Interactive data visualization using Chart.js v4.4.0:

**Chart Types:**
- **Temperature**: Line chart with gradient fills
- **Humidity**: Area chart with comfort zones
- **Wind**: Combination chart (speed + direction)
- **Rain**: Bar chart with accumulation data
- **Pressure**: Line chart with trend indicators
- **UV/Light**: Time-series charts with color coding

**Chart Features:**
- Real-time data updates
- Zoom and pan capabilities
- Responsive design
- Smooth animations
- Average trend lines
- Custom styling

### Event Management System
Sophisticated event handling for interactive elements:
- **Click Handlers**: Unit conversion, tooltip toggles
- **Event Propagation**: Prevents conflicts between nested elements
- **Debounced Updates**: Efficient real-time data handling
- **Error Recovery**: Graceful handling of failed API calls

## Static Asset Management

### Cache-Busting Strategy
Static files are served with timestamps to prevent browser caching issues:
```
script.js?t=1694808000
styles.css?t=1694808000
date-fns.min.js?t=1694808000
```

### File Serving
The Go web server automatically serves these static assets:
- **Automatic Detection**: Server detects files in static/ directory
- **MIME Types**: Proper content-type headers for each file type
- **Error Handling**: 404 responses for missing assets
- **Performance**: Efficient file serving for production use

## Interactive Features

### Unit Conversion System
Users can click any weather card to toggle units:
- **Temperature**: Celsius ↔ Fahrenheit
- **Wind Speed**: mph ↔ kph
- **Rain**: inches ↔ millimeters
- **Pressure**: mb ↔ inHg
- **Persistence**: Settings saved in browser localStorage

### Information Tooltips
Each sensor has an ℹ️ icon that displays detailed information:
- **Lux Reference**: Illuminance levels and real-world examples
- **Pressure Analysis**: Weather forecasting methodology
- **UV Index**: EPA risk categories and protection recommendations
- **Rain Intensity**: Precipitation classification system
- **Humidity Comfort**: Comfort levels and health effects

### Real-time Updates
The dashboard continuously updates with fresh data:
- **10-Second Interval**: Automatic data refresh
- **Error Handling**: Graceful degradation during API failures
- **Connection Status**: Visual indicators for server connectivity
- **Performance**: Efficient updates without full page reloads

## Mobile Responsiveness

### Adaptive Layout
The interface adapts to different screen sizes:
- **Desktop (>1200px)**: 3-column grid with full features
- **Tablet (768px-1200px)**: 2-column grid with optimized spacing
- **Mobile (<768px)**: Single-column layout with simplified interface

### Touch Optimization
- **Touch Targets**: Minimum 44px touch areas for mobile devices
- **Hover States**: Appropriate hover effects that work on touch devices
- **Scrolling**: Smooth scrolling and proper viewport handling

## Development

### File Modification
When modifying static assets:
1. Edit the relevant file (`script.js`, `styles.css`)
2. Restart the Go application to pick up changes
3. Hard refresh browser (Ctrl+Shift+R) to bypass cache
4. Test across different devices and screen sizes

### Adding New Assets
To add new static assets:
1. Place files in the `pkg/web/static/` directory
2. Reference them in HTML templates with proper paths
3. Update the web server if special handling is needed
4. Test cache-busting functionality

### Performance Optimization
- **Minification**: Consider minifying CSS and JavaScript for production
- **Compression**: Enable gzip compression in the Go server
- **Caching**: Implement proper cache headers for static assets
- **CDN**: Consider CDN hosting for external libraries in production

## Dependencies

### External Libraries
- **Chart.js v4.4.0**: Interactive charts and data visualization
- **date-fns v2.30.0**: Date manipulation and formatting

### Browser Compatibility
- **Modern Browsers**: Chrome, Firefox, Safari, Edge (latest versions)
- **Mobile Browsers**: iOS Safari, Android Chrome
- **JavaScript**: ES6+ features used (const, let, arrow functions, fetch API)
- **CSS**: Flexbox and Grid layout, CSS custom properties

### No Build Process
The frontend uses vanilla technologies without a build process:
- No npm/webpack/babel required
- Direct browser execution
- Simplified deployment
- Reduced complexity for maintenance