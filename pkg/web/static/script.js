// Debug logging configuration
let DEBUG_MODE = false;
const logLevels = {
    DEBUG: 0,
    INFO: 1, 
    WARN: 2,
    ERROR: 3
};

// Check log level from URL parameter or localStorage
const urlParams = new URLSearchParams(window.location.search);
if (urlParams.get('loglevel') === 'debug' || localStorage.getItem('loglevel') === 'debug') {
    DEBUG_MODE = true;
    console.log('🐛 DEBUG MODE ENABLED - Will log calculated values, API calls and responses');
}

// Enhanced debug logger
function debugLog(level, message, data = null) {
    if (!DEBUG_MODE && level < logLevels.INFO) return;
    
    const levelNames = ['DEBUG', 'INFO', 'WARN', 'ERROR'];
    const emoji = ['🐛', 'ℹ️', '⚠️', '❌'];
    const timestamp = new Date().toISOString().split('T')[1].split('.')[0];
    
    console.log(`${emoji[level]} [${timestamp}] ${levelNames[level]}: ${message}`);
    if (data !== null) {
        console.log('   Data:', data);
    }
}

let units = {
    temperature: localStorage.getItem('temperature-unit') || 'celsius',
    wind: localStorage.getItem('wind-unit') || 'mph',
    rain: localStorage.getItem('rain-unit') || 'inches',
    pressure: localStorage.getItem('pressure-unit') || 'mb'
};

// Friendly unit label helper (keeps formatting consistent with chart.html)
function prettyUnitLabel(key, val) {
    const map = {
        temperature: { celsius: '°C', fahrenheit: '°F' },
        temp: { celsius: '°C', fahrenheit: '°F' },
        wind: { mph: 'mph', kmh: 'km/h', kph: 'km/h', mps: 'm/s' },
        rain: { inches: 'in', mm: 'mm' },
        pressure: { inHg: 'inHg', mb: 'mb', hpa: 'hPa' },
        uv: { uv: 'UV' },
        illuminance: { lux: 'lux' }
    };
    const km = map[key] || map[key.toLowerCase()] || {};
    return (km[val] || val || '').toString();
}

// Load units configuration from server
async function loadUnitsConfig() {
    try {
        const response = await fetch('/api/units');
        const serverUnits = await response.json();
        
        // Map server units to client units
        if (serverUnits.units === 'imperial') {
            units.temperature = 'fahrenheit';
            units.wind = 'mph';
            units.rain = 'inches';
        } else if (serverUnits.units === 'metric') {
            units.temperature = 'celsius';
            units.wind = 'kmh';
            units.rain = 'mm';
        } else if (serverUnits.units === 'sae') {
            units.temperature = 'fahrenheit';
            units.wind = 'mph';
            units.rain = 'inches';
        }
        
        // Set pressure units
        units.pressure = serverUnits.unitsPressure;
        
        debugLog(logLevels.DEBUG, 'Loaded units config from server', serverUnits);
        debugLog(logLevels.DEBUG, 'Mapped client units', units);
        
        // Update localStorage for persistence
        localStorage.setItem('temperature-unit', units.temperature);
        localStorage.setItem('wind-unit', units.wind);
        localStorage.setItem('rain-unit', units.rain);
        localStorage.setItem('pressure-unit', units.pressure);
        
        return true;
    } catch (error) {
        debugLog(logLevels.ERROR, 'Failed to load units config from server, using defaults', error);
        return false;
    }
}

let weatherData = null;
let forecastData = null; // Store current forecast data for unit conversions
let statusData = null; // Store current status data for unit conversions
const charts = {};

// Provide a global openChartPopout so click handlers can call it even if
// forceChartColors() or other initialization hasn't finished. This mirrors
// the per-dataset metadata encoding used by the internal helper.
window.openChartPopout = window.openChartPopout || function(type, field, title, color) {
    try {
        const chartObj = charts[type];
        const datasetsMeta = [];
        if (chartObj && chartObj.data && Array.isArray(chartObj.data.datasets)) {
            chartObj.data.datasets.forEach(ds => {
                const meta = {};
                if (ds.label) meta.label = ds.label;
                if (ds.borderColor) meta.borderColor = ds.borderColor;
                if (ds.backgroundColor) meta.backgroundColor = ds.backgroundColor;
                if (ds.borderDash) meta.borderDash = ds.borderDash;
                if (ds.borderWidth !== undefined) meta.borderWidth = ds.borderWidth;
                if (ds.fill !== undefined) meta.fill = ds.fill;
                if (ds.pointRadius !== undefined) meta.pointRadius = ds.pointRadius;
                if (ds.tension !== undefined) meta.tension = ds.tension;
                if (ds.pointStyle !== undefined) meta.pointStyle = ds.pointStyle;
                if (ds.showLine !== undefined) meta.showLine = ds.showLine;
                if (ds.stepped !== undefined) meta.stepped = ds.stepped;
                if (ds.order !== undefined) meta.order = ds.order;
                if (ds.spanGaps !== undefined) meta.spanGaps = ds.spanGaps;
                if (String(ds.label).toLowerCase().includes('average')) meta.role = 'average';
                if (String(ds.label).toLowerCase().includes('trend')) meta.role = 'trend';
                if (String(ds.label).toLowerCase().includes('today') || String(ds.label).toLowerCase().includes('total')) meta.role = 'total';
                datasetsMeta.push(meta);
            });
        }

        const cfg = { type: type, field: field, title: title, color: color, units: units, datasets: datasetsMeta };
        const encoded = encodeURIComponent(JSON.stringify(cfg));
        const url = '/chart/' + type + '?config=' + encoded;
        window.open(url, '_blank');
    } catch (e) {
        debugLog(logLevels.ERROR, 'Global openChartPopout failed', e);
    }
};

// Expose minimal debug hooks for automated tests / headless browsers.
// These do not alter application behavior and simply return current in-memory state.
window.getWeatherData = function() { return weatherData; };
window.getCharts = function() { return charts; };
const maxDataPoints = 1000; // As specified in requirements

// Ensure a given dataset index exists on a chart. Creates a minimal dataset if missing.
// Guard counter to avoid infinite retry loops when initializing charts fails repeatedly
let __chartInitAttempts = 0;
// Flag to indicate charts have been successfully initialized
let __chartsInitialized = false;
// Flag to indicate vendor Chart.js is currently being loaded
let __chartVendorLoading = false;
// Whether we've attempted a global Chart.destroy() sweep already
let __didGlobalChartDestroy = false;
// Flags to indicate initial data fetch completion
let __weatherFetched = false;
let __statusFetched = false;
// Public readiness flag for tests to wait on
window.__dashboardReady = window.__dashboardReady || false;

function trySetDashboardReady() {
    try {
            if (window.__dashboardReady) return;
            try {
                // require charts initialized, first fetches done, and temperature chart has >=2 data points
                const tempHasData = charts && charts.temperature && charts.temperature.data && charts.temperature.data.datasets && charts.temperature.data.datasets[0] && charts.temperature.data.datasets[0].data && charts.temperature.data.datasets[0].data.length >= 2;

                if (__chartsInitialized && __weatherFetched && __statusFetched && tempHasData) {
                    window.__dashboardReady = true;
                    debugLog(logLevels.INFO, 'Dashboard ready - window.__dashboardReady set to true (includes temperature dataset length check)');
                }
            } catch (e) {
                // safe fallback: don't block if charts object is malformed
                console.warn('trySetDashboardReady: error checking dataset lengths', e);
                if (__chartsInitialized && __weatherFetched && __statusFetched) {
                    window.__dashboardReady = true;
                    debugLog(logLevels.INFO, 'Dashboard ready - window.__dashboardReady set to true (fallback)');
                }
        }
    } catch (e) {
        debugLog(logLevels.WARN, 'trySetDashboardReady encountered error', e);
    }
}

function destroyAllCharts() {
    try {
        if (typeof Chart === 'undefined') return;

        // Destroy any chart instances tracked by Chart.js itself
        try {
            // Chart.instances may be an object or Map-like; handle both
            const instances = Chart.instances ? Object.values(Chart.instances) : [];
            instances.forEach(inst => {
                try {
                    if (inst && typeof inst.destroy === 'function') {
                        inst.destroy();
                    }
                } catch (e) {
                    debugLog(logLevels.WARN, 'Failed to destroy Chart instance during global sweep', e);
                }
            });
        } catch (e) {
            debugLog(logLevels.WARN, 'Error enumerating Chart.instances', e);
        }

        // Also try Chart.getChart on known canvases as a fallback
        const canvasIds = ['temperature-chart','humidity-chart','wind-chart','rain-chart','pressure-chart','light-chart','uv-chart'];
        canvasIds.forEach(id => {
            const el = document.getElementById(id);
            if (!el) return;
            try {
                const existing = (typeof Chart.getChart === 'function') ? Chart.getChart(el) : null;
                if (existing && typeof existing.destroy === 'function') existing.destroy();
            } catch (e) {
                debugLog(logLevels.WARN, `Error destroying chart via getChart for ${id}`, e);
            }
        });

        // Clear our charts map
        Object.keys(charts).forEach(k => delete charts[k]);
        __didGlobalChartDestroy = true;
        debugLog(logLevels.INFO, 'Global Chart.js sweep completed - destroyed existing charts');
    } catch (e) {
        debugLog(logLevels.WARN, 'destroyAllCharts encountered an error', e);
    }
}
function ensureDataset(chart, index) {
    if (!chart || !chart.data) return;
    if (!chart.data.datasets) chart.data.datasets = [];
    if (!chart.data.datasets[index]) {
        // Create placeholder datasets up to the requested index
        for (let i = chart.data.datasets.length; i <= index; i++) {
            chart.data.datasets[i] = { data: [] };
        }
    }
}

function initCharts() {
    debugLog(logLevels.DEBUG, 'Initializing all charts with configuration');
    
    // If Chart.js has already created Chart instances on these canvases,
    // destroy them first to avoid "Canvas is already in use" errors when
    // re-initializing (headless tests or dynamic reload scenarios).
    try {
        if (typeof Chart !== 'undefined' && Chart.getChart) {
            // Map canvas IDs to chart keys so we can null out the charts object
            const mapping = {
                'temperature-chart': 'temperature',
                'humidity-chart': 'humidity',
                'wind-chart': 'wind',
                'rain-chart': 'rain',
                'pressure-chart': 'pressure',
                'light-chart': 'light',
                'uv-chart': 'uv'
            };

            Object.keys(mapping).forEach(id => {
                const el = document.getElementById(id);
                if (el) {
                    try {
                        const existing = Chart.getChart(el);
                        if (existing && typeof existing.destroy === 'function') {
                            debugLog(logLevels.DEBUG, `Destroying existing Chart instance for canvas: ${id}`);
                            existing.destroy();
                        }
                    } catch (e) {
                        debugLog(logLevels.WARN, `Failed to destroy existing Chart for ${id}`, e);
                    }

                    // Ensure we remove any lingering reference in the charts map
                    const key = mapping[id];
                    if (charts[key]) {
                        try { delete charts[key]; } catch(_) { charts[key] = null; }
                    }
                }
            });
        }
    } catch (e) {
        debugLog(logLevels.WARN, 'Error while cleaning up existing Chart instances', e);
    }

    // Set Chart.js default locale to ensure 24-hour format
    Chart.defaults.font.family = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';
    
    const ctxTemp = document.getElementById('temperature-chart').getContext('2d');
    const ctxHumidity = document.getElementById('humidity-chart').getContext('2d');
    const ctxWind = document.getElementById('wind-chart').getContext('2d');
    const ctxRain = document.getElementById('rain-chart').getContext('2d');
    const ctxPressure = document.getElementById('pressure-chart').getContext('2d');
    const ctxLight = document.getElementById('light-chart').getContext('2d');
    const uvElement = document.getElementById('uv-chart');
    const ctxUV = uvElement ? uvElement.getContext('2d') : null;

    const chartConfig = {
        type: 'line',
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(20,20,20,0.9)',
                    titleColor: '#fff',
                    bodyColor: '#fff',
                    padding: 8,
                    cornerRadius: 6,
                    callbacks: {
                        title: function(context) {
                            const date = new Date(context[0].parsed.x);
                            return date.toLocaleDateString('en-US', { month: 'short', day: '2-digit' }) + ', ' + date.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', hour12: false });
                        }
                    }
                }
            },
            scales: {
                x: {
                    display: true,
                    type: 'time',
                    time: {
                        displayFormats: { minute: 'HH:mm', hour: 'HH:mm', day: 'MMM dd' },
                        tooltipFormat: 'MMM dd, HH:mm'
                    },
                    grid: { display: true, color: 'rgba(0,0,0,0.06)' },
                    ticks: { maxTicksLimit: 6, color: '#666', font: { size: 11 }, callback: function(value){ return new Date(value).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', hour12: false }); } },
                    title: { display: false }
                },
                y: {
                    display: true,
                    grid: { display: true, color: 'rgba(0,0,0,0.06)' },
                    ticks: { maxTicksLimit: 5, color: '#444', font: { size: 12 }, callback: function(value){ return value.toFixed(1); } },
                    title: { display: true, text: 'Value', color: '#444', font: { size: 12, weight: '600' } }
                }
            },
            elements: {
                point: { radius: 3, hoverRadius: 5, backgroundColor: '#fff', borderWidth: 1 },
                line: { borderWidth: 2.5, borderJoinStyle: 'round', tension: 0.35 }
            },
            interaction: { intersect: false, mode: 'index' }
        }
    };

    // If a chart already exists in memory, do not recreate it (prevents reuse errors)
    if (charts.temperature && typeof charts.temperature.update === 'function') {
        debugLog(logLevels.INFO, 'Temperature chart already exists in memory; skipping creation');
    } else {
        charts.temperature = new Chart(ctxTemp, {
        ...chartConfig,
        data: {
            datasets: [{
                data: [],
                borderColor: '#ff6384',
                backgroundColor: 'rgba(255, 99, 132, 0.1)',
                fill: false,
                tension: 0.4,
                spanGaps: false,
                label: 'Temperature'
            }, {
                data: [],
                borderColor: '#00cc66',
                backgroundColor: 'rgba(0, 204, 102, 0.2)',
                borderDash: [5, 5],
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                tension: 0,
                label: 'Average'
            }]
        }
    });
        // clicking on a chart should open a pop-out detailed chart page
        document.getElementById('temperature-chart').addEventListener('click', function(){
            openChartPopout('temperature', 'temperature', 'Temperature', charts.temperature.data.datasets[0].borderColor);
        });
    }
    
    // Force the colors after creation
    charts.temperature.data.datasets[1].borderColor = '#00cc66';
    charts.temperature.data.datasets[1].backgroundColor = 'rgba(0, 204, 102, 0.2)';
    
    debugLog(logLevels.INFO, 'Temperature chart created with colors:', {
        dataColor: charts.temperature.data.datasets[0].borderColor,
        avgColor: charts.temperature.data.datasets[1].borderColor
    });

    if (charts.humidity && typeof charts.humidity.update === 'function') {
        debugLog(logLevels.INFO, 'Humidity chart already exists in memory; skipping creation');
    } else {
        charts.humidity = new Chart(ctxHumidity, {
        ...chartConfig,
        data: {
            datasets: [{
                data: [],
                borderColor: 'rgba(54, 162, 235, 0.8)',
                backgroundColor: 'rgba(54, 162, 235, 0.1)',
                fill: false,
                tension: 0.4,
                spanGaps: false,
                label: 'Humidity'
            }, {
                data: [],
                borderColor: '#ff8533',
                backgroundColor: 'rgba(255, 133, 51, 0.2)',
                borderDash: [5, 5],
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                tension: 0,
                label: 'Average'
            }]
        }
    });
        document.getElementById('humidity-chart').addEventListener('click', function(){
            openChartPopout('humidity', 'humidity', 'Humidity', charts.humidity.data.datasets[0].borderColor);
        });
    }
    
    // Force the colors after creation
    charts.humidity.data.datasets[1].borderColor = '#ff8533';
    charts.humidity.data.datasets[1].backgroundColor = 'rgba(255, 133, 51, 0.2)';
    
    debugLog(logLevels.INFO, 'Humidity chart created with colors:', {
        dataColor: charts.humidity.data.datasets[0].borderColor,
        avgColor: charts.humidity.data.datasets[1].borderColor
    });

    if (charts.wind && typeof charts.wind.update === 'function') {
        debugLog(logLevels.INFO, 'Wind chart already exists in memory; skipping creation');
    } else {
        charts.wind = new Chart(ctxWind, {
        ...chartConfig,
        data: {
            datasets: [{
                data: [],
                borderColor: 'rgba(75, 192, 192, 0.8)',
                backgroundColor: 'rgba(75, 192, 192, 0.1)',
                fill: false,
                tension: 0.4,
                spanGaps: false,
                label: 'Wind'
            }, {
                data: [],
                borderColor: '#ff4d4d',
                backgroundColor: 'rgba(255, 77, 77, 0.2)',
                borderDash: [5, 5],
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                tension: 0,
                label: 'Average'
            }]
        }
    });
        document.getElementById('wind-chart').addEventListener('click', function(){
            openChartPopout('wind', 'windSpeed', 'Wind Speed', charts.wind.data.datasets[0].borderColor);
        });
    }

    if (charts.rain && typeof charts.rain.update === 'function') {
        debugLog(logLevels.INFO, 'Rain chart already exists in memory; skipping creation');
    } else {
        charts.rain = new Chart(ctxRain, {
        ...chartConfig,
        data: {
            datasets: [{
                data: [],
                borderColor: '#9966ff',
                backgroundColor: 'rgba(153, 102, 255, 0.1)',
                fill: false,
                tension: 0.4,
                spanGaps: false,
                label: 'Rain (incremental)',
                pointRadius: 2,
                pointHoverRadius: 4,
                order: 3  // Render data points at bottom layer
            }, {
                data: [],
                borderColor: '#66ff66',
                backgroundColor: 'rgba(102, 255, 102, 0.2)',
                borderDash: [5, 5],
                borderWidth: 3,
                fill: false,
                pointRadius: 0,
                tension: 0,
                label: 'Average',
                order: 2  // Render above data points
            }, {
                data: [],
                borderColor: '#ff6b35',
                backgroundColor: 'rgba(255, 107, 53, 0.1)',
                borderDash: [3, 3],
                borderWidth: 4,
                fill: false,
                pointRadius: 0,
                tension: 0,
                label: 'Today Total',
                order: 1  // Render on top of everything
            }]
        },
        options: {
            ...chartConfig.options,
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                ...chartConfig.options.plugins,
                tooltip: {
                    ...chartConfig.options.plugins.tooltip,
                    filter: function(tooltipItem) {
                        // Always show all datasets in rain chart tooltips
                        return true;
                    }
                }
            }
        }
    });
        document.getElementById('rain-chart').addEventListener('click', function(){
            openChartPopout('rain', 'rainAccum', 'Rain', charts.rain.data.datasets[0].borderColor);
        });
    }

    if (charts.pressure && typeof charts.pressure.update === 'function') {
        debugLog(logLevels.INFO, 'Pressure chart already exists in memory; skipping creation');
    } else {
        charts.pressure = new Chart(ctxPressure, {
        ...chartConfig,
        data: {
            datasets: [{
                data: [],
                borderColor: 'rgba(255, 159, 64, 0.8)',
                backgroundColor: 'rgba(255, 159, 64, 0.1)',
                fill: false,
                tension: 0.4,
                spanGaps: false,
                label: 'Pressure'
            }, {
                data: [],
                borderColor: '#4080ff',
                backgroundColor: 'rgba(64, 128, 255, 0.2)',
                borderDash: [5, 5],
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                tension: 0,
                label: 'Average'
            }, {
                data: [],
                borderColor: '#ff6384',
                backgroundColor: 'rgba(255, 99, 132, 0.1)',
                borderDash: [2, 2],
                borderWidth: 1.5,
                fill: false,
                pointRadius: 0,
                tension: 0,
                label: 'Trend'
            }]
        }
    });
        document.getElementById('pressure-chart').addEventListener('click', function(){
            openChartPopout('pressure', 'pressure', 'Pressure', charts.pressure.data.datasets[0].borderColor);
        });
    }

    if (charts.light && typeof charts.light.update === 'function') {
        debugLog(logLevels.INFO, 'Light chart already exists in memory; skipping creation');
    } else {
        charts.light = new Chart(ctxLight, {
        ...chartConfig,
        data: {
            datasets: [{
                data: [],
                borderColor: 'rgba(255, 205, 86, 0.8)',
                backgroundColor: 'rgba(255, 205, 86, 0.1)',
                fill: false,
                tension: 0.4,
                spanGaps: false,
                label: 'Light'
            }]
        }
    });
        document.getElementById('light-chart').addEventListener('click', function(){
            openChartPopout('light', 'illuminance', 'Illuminance', charts.light.data.datasets[0].borderColor);
        });
    }

    // Check if UV chart element exists before creating chart
    if (ctxUV) {
        if (charts.uv && typeof charts.uv.update === 'function') {
            debugLog(logLevels.INFO, 'UV chart already exists in memory; skipping creation');
        } else {
            charts.uv = new Chart(ctxUV, {
            ...chartConfig,
            data: {
                datasets: [{
                    data: [],
                    borderColor: 'rgba(153, 102, 255, 0.8)',
                    backgroundColor: 'rgba(153, 102, 255, 0.1)',
                    fill: false,
                    tension: 0.4,
                    spanGaps: false,
                    label: 'UV Index'
                }]
            }
        });
            debugLog(logLevels.DEBUG, 'UV chart created successfully');
        }
    } else {
        debugLog(logLevels.DEBUG, 'UV chart element not found, skipping UV chart creation');
    }
    
    // Force all chart colors after creation
    forceChartColors();
}

function forceChartColors() {
    debugLog(logLevels.INFO, '🎨 Forcing chart colors to complementary pairs');
    
    // Temperature: Red data → Green average
    if (charts.temperature) {
        ensureDataset(charts.temperature, 1);
        charts.temperature.data.datasets[1].borderColor = '#00cc66';
        charts.temperature.data.datasets[1].backgroundColor = 'rgba(0, 204, 102, 0.2)';
        charts.temperature.update('none');
    }
    
    // Humidity: Blue data → Orange average  
    if (charts.humidity) {
        ensureDataset(charts.humidity, 1);
        charts.humidity.data.datasets[1].borderColor = '#ff8533';
        charts.humidity.data.datasets[1].backgroundColor = 'rgba(255, 133, 51, 0.2)';
        charts.humidity.update('none');
    }
    
    // Wind: Teal data → Bright Red average (more visible)
    if (charts.wind) {
        ensureDataset(charts.wind, 1);
        charts.wind.data.datasets[1].borderColor = '#FF0000';
        charts.wind.data.datasets[1].backgroundColor = 'rgba(255, 0, 0, 0.3)';
        charts.wind.data.datasets[1].borderWidth = 3;
        charts.wind.update('none');
        
        debugLog(logLevels.INFO, '🌬️ Wind chart colors applied:', {
            dataColor: charts.wind.data.datasets[0].borderColor,
            avgColor: charts.wind.data.datasets[1].borderColor,
            avgDataPoints: charts.wind.data.datasets[1].data.length
        });
        // Attach UV chart click handler if the uv canvas exists
        const uvCanvasEl = document.getElementById('uv-chart');
        if (uvCanvasEl) {
            uvCanvasEl.addEventListener('click', function(){
                openChartPopout('uv', 'uv', 'UV Index', charts.uv.data.datasets[0].borderColor);
            });
        }
    }

    // helper to open popout chart page with encoded configuration
    function openChartPopout(type, field, title, color) {
        try {
            // Build a compact per-dataset metadata payload so the popout can mirror
            // the small-card chart visuals exactly (colors, dashes, widths, fill, etc).
            const chartObj = charts[type];
            const datasetsMeta = [];
            if (chartObj && chartObj.data && Array.isArray(chartObj.data.datasets)) {
                chartObj.data.datasets.forEach(ds => {
                    const meta = {};
                    if (ds.label) meta.label = ds.label;
                    if (ds.borderColor) meta.borderColor = ds.borderColor;
                    if (ds.backgroundColor) meta.backgroundColor = ds.backgroundColor;
                    if (ds.borderDash) meta.borderDash = ds.borderDash;
                    if (ds.borderWidth !== undefined) meta.borderWidth = ds.borderWidth;
                    if (ds.fill !== undefined) meta.fill = ds.fill;
                    if (ds.pointRadius !== undefined) meta.pointRadius = ds.pointRadius;
                    if (ds.tension !== undefined) meta.tension = ds.tension;
                    if (ds.pointStyle !== undefined) meta.pointStyle = ds.pointStyle;
                    if (ds.showLine !== undefined) meta.showLine = ds.showLine;
                    if (ds.stepped !== undefined) meta.stepped = ds.stepped;
                    if (ds.order !== undefined) meta.order = ds.order;
                    if (ds.spanGaps !== undefined) meta.spanGaps = ds.spanGaps;
                    // mark a simple role hint for common lines (Average/Trend/Today Total)
                    if (String(ds.label).toLowerCase().includes('average')) meta.role = 'average';
                    if (String(ds.label).toLowerCase().includes('trend')) meta.role = 'trend';
                    if (String(ds.label).toLowerCase().includes('today') || String(ds.label).toLowerCase().includes('total')) meta.role = 'total';
                    datasetsMeta.push(meta);
                });
            }

            // Build an incomingUnits hint to inform the popout about the units the server/raw data used.
            // Best-effort: if statusData or weatherData expose unit hints, prefer those; otherwise fall back to current `units` (safe default).
            let incomingUnits = null;
            try {
                if (statusData && statusData.unitHints) {
                    incomingUnits = statusData.unitHints;
                } else if (weatherData && weatherData.unitHints) {
                    incomingUnits = weatherData.unitHints;
                } else {
                    // fallback: assume incoming units match the current UI units
                    incomingUnits = Object.assign({}, units);
                }
            } catch (e) {
                incomingUnits = Object.assign({}, units);
            }

            const cfg = { type: type, field: field, title: title, color: color, units: units, incomingUnits: incomingUnits, datasets: datasetsMeta };
            const encoded = encodeURIComponent(JSON.stringify(cfg));
            const url = '/chart/' + type + '?config=' + encoded;
            window.open(url, '_blank');
        } catch(e) {
            debugLog(logLevels.ERROR, 'Failed to open chart popout', e);
        }
    }
    
    // Rain: Purple data → Yellow-green average → Orange 24h total
    if (charts.rain) {
        ensureDataset(charts.rain, 1);
        ensureDataset(charts.rain, 2);
        charts.rain.data.datasets[1].borderColor = '#66ff66';
        charts.rain.data.datasets[1].backgroundColor = 'rgba(102, 255, 102, 0.2)';
        charts.rain.data.datasets[2].borderColor = '#ff6b35';
        charts.rain.data.datasets[2].backgroundColor = 'rgba(255, 107, 53, 0.1)';
        charts.rain.update('none');
    }
    
    // Pressure: Orange data → Blue average
    if (charts.pressure) {
        ensureDataset(charts.pressure, 1);
        ensureDataset(charts.pressure, 2);
        charts.pressure.data.datasets[1].borderColor = '#4080ff';
        charts.pressure.data.datasets[1].backgroundColor = 'rgba(64, 128, 255, 0.2)';
        charts.pressure.update('none');
    }
    
    // Light: Only has main data, no average line needed
    // (Light naturally goes to zero at night)
    
    // UV: Only has main data, no average line needed  
    // (UV naturally goes to zero at night)
    
    debugLog(logLevels.INFO, '✅ Chart colors forced - complementary pairs applied');
}

function updateElevationDisplay() {
    const tempestElevation = document.getElementById('tempest-elevation');
    
    if (!statusData || !statusData.elevation || !tempestElevation) {
        if (tempestElevation) tempestElevation.textContent = '--';
        return;
    }
    
    if (units.temperature === 'fahrenheit') {
        // If using imperial units, show elevation in feet
        const elevationFt = statusData.elevation * 3.28084;
        tempestElevation.textContent = `${Math.round(elevationFt)} ft`;
    } else {
        // If using metric units, show elevation in meters
        tempestElevation.textContent = `${Math.round(statusData.elevation)} m`;
    }
}

function updateUnits() {
    // Use prettyUnitLabel to ensure consistency with popout formatting
    const tempEl = document.getElementById('temperature-unit');
    if (tempEl) tempEl.textContent = prettyUnitLabel('temperature', units.temperature);
    const windEl = document.getElementById('wind-unit');
    if (windEl) windEl.textContent = prettyUnitLabel('wind', units.wind);
    
    // Special handling for rain-unit to preserve the info icon
    const rainUnitElement = document.getElementById('rain-unit');
    const rainInfoIcon = rainUnitElement.querySelector('.info-icon');
    const newRainUnitText = units.rain === 'inches' ? 'in ' : 'mm ';
    
    if (rainInfoIcon) {
        rainUnitElement.innerHTML = newRainUnitText + rainInfoIcon.outerHTML;
        // Re-attach click event listener to the rain info icon
        const newRainInfoIcon = rainUnitElement.querySelector('.info-icon');
        if (newRainInfoIcon) {
            newRainInfoIcon.addEventListener('click', function(event) {
                event.stopPropagation();
                toggleRainTooltip(event);
            });
        }
    }
    
    // Special handling for pressure-unit to preserve the info icon
    const pressureUnitElement = document.getElementById('pressure-unit');
    const infoIcon = pressureUnitElement.querySelector('.info-icon');
    const newUnitText = units.pressure === 'mb' ? 'mb ' : 'inHg ';
    
    console.log('🔧 updateUnits() - Pressure unit update:', {
        pressureUnitElement: !!pressureUnitElement,
        infoIcon: !!infoIcon,
        currentInnerHTML: pressureUnitElement ? pressureUnitElement.innerHTML : 'N/A',
        newUnitText: newUnitText,
        infoIconOuterHTML: infoIcon ? infoIcon.outerHTML : 'N/A'
    });
    
    if (infoIcon) {
        // Preserve the info icon by only updating the text node
        pressureUnitElement.innerHTML = newUnitText + infoIcon.outerHTML;
        console.log('🔧 updateUnits() - After setting innerHTML:', pressureUnitElement.innerHTML);
        
        // Re-attach the event listener to prevent click bubbling
        const newInfoIcon = pressureUnitElement.querySelector('.info-icon');
        if (newInfoIcon) {
            newInfoIcon.addEventListener('click', function(e) {
                e.stopPropagation();
                togglePressureTooltip(e);
            });
            console.log('🔧 updateUnits() - Re-attached click event listener to info icon');
        }
    } else {
        // Fallback if no info icon found
        pressureUnitElement.textContent = newUnitText.trim();
        console.log('🔧 updateUnits() - Info icon not found, used textContent fallback');
    }
    
    // Update elevation display with new units
    updateElevationDisplay();
    
    // Update chart labels with new units
    updateChartLabels();
}

function toggleUnit(sensor) {
    console.log('🚀 toggleUnit() called with sensor:', sensor);
    const oldUnit = units[sensor];
    
    if (sensor === 'temperature') {
        units.temperature = units.temperature === 'celsius' ? 'fahrenheit' : 'celsius';
        localStorage.setItem('temperature-unit', units.temperature);
    } else if (sensor === 'wind') {
        units.wind = units.wind === 'mph' ? 'kph' : 'mph';
        localStorage.setItem('wind-unit', units.wind);
    } else if (sensor === 'rain') {
        units.rain = units.rain === 'inches' ? 'mm' : 'inches';
        localStorage.setItem('rain-unit', units.rain);
    } else if (sensor === 'pressure') {
        units.pressure = units.pressure === 'mb' ? 'inHg' : 'mb';
        localStorage.setItem('pressure-unit', units.pressure);
        console.log('🔄 toggleUnit(pressure) - Unit changed:', {
            oldUnit: oldUnit,
            newUnit: units.pressure
        });
    }
    
    debugLog(logLevels.DEBUG, `Unit toggle for ${sensor}`, {
        oldUnit: oldUnit,
        newUnit: units[sensor],
        allUnits: units
    });
    
    console.log('🔄 toggleUnit() - About to call functions:', {
        sensor: sensor,
        sequence: ['updateUnits()', 'updateDisplay()', 'recalculateAverages()']
    });
    
    updateUnits();
    console.log('🔄 toggleUnit() - updateUnits() completed');
    
    updateDisplay();
    console.log('🔄 toggleUnit() - updateDisplay() completed');
    
    refreshForecastDisplay(); // Update forecast display with new units
    console.log('🔄 toggleUnit() - refreshForecastDisplay() completed');
    
    recalculateAverages(sensor);
    console.log('🔄 toggleUnit() - recalculateAverages() completed');
    console.log('🔄 toggleUnit() - All functions completed');
}

function updateChartLabels() {
    // Update temperature chart label (use prettyUnitLabel for consistent formatting)
    if (charts.temperature && charts.temperature.options && charts.temperature.options.scales) {
        charts.temperature.options.scales.y.title = {
            display: true,
            text: `Temperature (${prettyUnitLabel('temperature', units.temperature)})`
        };
    }
    
    // Update wind chart label
    if (charts.wind && charts.wind.options && charts.wind.options.scales) {
        let windUnit = 'm/s';
        if (units.wind === 'mph') {
            windUnit = 'mph';
        } else if (units.wind === 'kmh') {
            windUnit = 'km/h';
        }
        charts.wind.options.scales.y.title = {
            display: true,
            text: `Wind Speed (${prettyUnitLabel('wind', units.wind)})`
        };
    }
    
    // Update rain chart label
    if (charts.rain && charts.rain.options && charts.rain.options.scales) {
        const rainUnit = units.rain === 'inches' ? 'in' : 'mm';
        charts.rain.options.scales.y.title = {
            display: true,
            text: `Rainfall (${prettyUnitLabel('rain', units.rain)})`
        };
    }
    
    // Update pressure chart label
    if (charts.pressure && charts.pressure.options && charts.pressure.options.scales) {
        const pressureUnit = units.pressure === 'inHg' ? 'inHg' : 'mb';
        charts.pressure.options.scales.y.title = {
            display: true,
            text: `Pressure (${prettyUnitLabel('pressure', units.pressure)})`
        };
    }
    
    // Update all charts
    Object.values(charts).forEach(chart => {
        if (chart && typeof chart.update === 'function') {
            chart.update('none');
        }
    });
    
    debugLog(logLevels.DEBUG, 'Chart labels updated with new units', units);
}

function degreesToDirection(degrees) {
    const directions = ['N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE', 'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'];
    const index = Math.round(degrees / 22.5) % 16;
    return directions[index];
}

function updateArrow(direction) {
    const arrows = {
        'N': '↑', 'NNE': '↗', 'NE': '↗', 'ENE': '↗',
        'E': '→', 'ESE': '↘', 'SE': '↘', 'SSE': '↘',
        'S': '↓', 'SSW': '↙', 'SW': '↙', 'WSW': '↙',
        'W': '←', 'WNW': '↖', 'NW': '↖', 'NNW': '↖'
    };
    return arrows[direction] || '↑';
}

function celsiusToFahrenheit(celsius) {
    return (celsius * 9/5) + 32;
}

function fahrenheitToCelsius(fahrenheit) {
    return (fahrenheit - 32) * 5/9;
}

function mphToKph(mph) {
    return mph * 1.60934;
}

function kphToMph(kph) {
    return kph / 1.60934;
}

function inchesToMm(inches) {
    return inches * 25.4;
}

function mmToInches(mm) {
    return mm / 25.4;
}

function mbToInHg(mb) {
    return mb * 0.02953;
}

function inHgToMb(inHg) {
    return inHg / 0.02953;
}

function calculateHeatIndex(tempC, humidity) {
    // Convert temperature to Fahrenheit for calculation
    const tempF = (tempC * 9/5) + 32;
    
    // If conditions don't warrant heat index calculation, return the temperature
    if (tempF < 80 || humidity < 40) {
        debugLog(logLevels.DEBUG, 'Heat index conditions not met, using actual temperature', {
            tempF: tempF,
            humidity: humidity,
            result: tempC
        });
        return tempC; // Return original temperature in Celsius
    }
    
    // Heat Index calculation using the NOAA formula
    const c1 = -42.379, c2 = 2.04901523, c3 = 10.14333127, c4 = -0.22475541;
    const c5 = -0.00683783, c6 = -0.05481717, c7 = 0.00122874, c8 = 0.00085282, c9 = -0.00000199;
    
    // Calculate heat index in Fahrenheit
    const heatIndexF = c1 + (c2 * tempF) + (c3 * humidity) + (c4 * tempF * humidity) +
                     (c5 * tempF * tempF) + (c6 * humidity * humidity) +
                     (c7 * tempF * tempF * humidity) + (c8 * tempF * humidity * humidity) +
                     (c9 * tempF * tempF * humidity * humidity);
    
    // Convert back to Celsius
    const heatIndexC = (heatIndexF - 32) * 5/9;
    
    debugLog(logLevels.DEBUG, 'Heat index calculated', {
        tempC: tempC,
        tempF: tempF,
        humidity: humidity,
        heatIndexF: heatIndexF,
        heatIndexC: heatIndexC
    });
    
    return heatIndexC;
}

function getPrecipitationTypeDescription(precipType) {
    switch (precipType) {
        case 0: return 'None';
        case 1: return 'Rain';
        case 2: return 'Hail';
        case 3: return 'Rain + Hail';
        default: return 'Unknown';
    }
}

function kmToMiles(km) {
    return km / 1.60934;
}

function milesToKm(miles) {
    return miles * 1.60934;
}

function calculateAverage(data) {
    if (data.length === 0) return 0;
    const sum = data.reduce((acc, point) => acc + point.y, 0);
    const average = sum / data.length;
    
    debugLog(logLevels.DEBUG, `Average calculated for ${data.length} data points`, {
        sum: sum,
        count: data.length,
        average: average
    });
    
    return average;
}

// Calculate linear regression trend line
function calculateTrendLine(data) {
    if (!data || data.length < 2) return [];
    
    // Convert timestamps to numerical values for regression
    const points = data.map((point, index) => ({
        x: index, // Use index as x for linear progression
        y: point.y,
        timestamp: point.x
    }));
    
    const n = points.length;
    const sumX = points.reduce((sum, point) => sum + point.x, 0);
    const sumY = points.reduce((sum, point) => sum + point.y, 0);
    const sumXY = points.reduce((sum, point) => sum + (point.x * point.y), 0);
    const sumXX = points.reduce((sum, point) => sum + (point.x * point.x), 0);
    
    // Calculate slope (m) and intercept (b) for y = mx + b
    const slope = (n * sumXY - sumX * sumY) / (n * sumXX - sumX * sumX);
    const intercept = (sumY - slope * sumX) / n;
    
    // Generate trend line points using original timestamps
    const trendLine = points.map((point, index) => ({
        x: point.timestamp,
        y: slope * index + intercept
    }));
    
    debugLog(logLevels.DEBUG, `Trend line calculated for ${data.length} data points`, {
        slope: slope,
        intercept: intercept,
        firstPoint: trendLine[0],
        lastPoint: trendLine[trendLine.length - 1]
    });
    
    return trendLine;
}

function updateAverageLine(chart, data) {
    // Ensure the second dataset exists (used for moving average). Create a minimal
    // placeholder if it's missing so assignments below won't throw.
    if (!chart.data.datasets[1]) {
        chart.data.datasets[1] = chart.data.datasets[1] || { data: [] };
    }

    if (data.length === 0) {
        chart.data.datasets[1].data = [];
        return;
    }

    // For the pressure chart we want a single, constant average line
    // across the full time range (horizontal line). For all other
    // charts use the moving average implementation.
    // Treat these charts as summary charts where a single global average
    // (horizontal line) is more appropriate than a moving average.
    const constantAvgCharts = ['pressure', 'temperature', 'humidity', 'wind'];
    const datasetLabel = (chart && chart.data && chart.data.datasets && chart.data.datasets[0] && String(chart.data.datasets[0].label).toLowerCase()) || '';
    const chartNameFromLabel = constantAvgCharts.find(name => datasetLabel.includes(name));
    const isConstantAvgChart = chartNameFromLabel || (chart === charts.pressure) || (chart === charts.temperature) || (chart === charts.humidity) || (chart === charts.wind);

    if (isConstantAvgChart) {
        // Compute simple average of all y values
        let sum = 0;
        let count = 0;
        for (let i = 0; i < data.length; i++) {
            if (data[i] && typeof data[i].y === 'number') {
                sum += data[i].y;
                count++;
            }
        }

        if (count === 0) {
            chart.data.datasets[1].data = [];
            return;
        }

        const avg = sum / count;

        // Create a horizontal line spanning start..end timestamps
        const firstX = data[0].x;
        const lastX = data[data.length - 1].x;

        chart.data.datasets[1].data = [
            { x: firstX, y: avg },
            { x: lastX, y: avg }
        ];

        debugLog(logLevels.DEBUG, 'Constant average updated', {
            chartLabel: chart.data.datasets[0].label || chartNameFromLabel || 'Chart',
            dataPoints: data.length,
            average: avg
        });
        return;
    }

    // Calculate a moving average with a window of 10% of total data points (minimum 5, maximum 50)
    const windowSize = Math.max(5, Math.min(50, Math.floor(data.length * 0.1)));
    const movingAverageData = [];

    for (let i = 0; i < data.length; i++) {
        // Calculate the range for the moving window
        const start = Math.max(0, i - Math.floor(windowSize / 2));
        const end = Math.min(data.length - 1, i + Math.floor(windowSize / 2));
        
        // Calculate average for the window
        let sum = 0;
        let count = 0;
        for (let j = start; j <= end; j++) {
            if (data[j] && typeof data[j].y === 'number') {
                sum += data[j].y;
                count++;
            }
        }
        
        if (count > 0) {
            movingAverageData.push({
                x: data[i].x,
                y: sum / count
            });
        }
    }

    // Safely assign the moving average points to the second dataset
    chart.data.datasets[1].data = movingAverageData;
    
    debugLog(logLevels.DEBUG, 'Moving average updated', {
        chartLabel: chart.data.datasets[0].label || 'Unknown',
        dataPoints: data.length,
        windowSize: windowSize,
        averagePoints: movingAverageData.length
    });
}

function update24HourAccumulationLine(chart, rainDailyTotal, units) {
    // For rain chart only - updates the third dataset (index 2) with 24-hour accumulation
    debugLog(logLevels.INFO, 'Updating 24h Rain Line', {
        rainDailyTotal: rainDailyTotal,
        units: units?.rain || 'inches',
        hasDataset2: !!chart.data.datasets[2]
    });
    
    // Also log to console for debugging
    console.log('Rain Daily Total:', rainDailyTotal, 'Type:', typeof rainDailyTotal);
    
    if (!chart.data.datasets[2] || rainDailyTotal === undefined || rainDailyTotal === null) {
        if (chart.data.datasets[2]) {
            chart.data.datasets[2].data = [];
        }
        console.log('WARNING: No rain daily total data available');
        debugLog(logLevels.WARN, '24h Rain Line: No data or dataset - line will be empty');
        return;
    }

    // Convert daily total based on current units
    let convertedDailyTotal = rainDailyTotal;
    if (units && units.rain === 'mm') {
        convertedDailyTotal = inchesToMm(rainDailyTotal);
    }

    // Create a horizontal line at the daily total level across the current time range
    const mainData = chart.data.datasets[0].data;
    if (mainData.length === 0) {
        chart.data.datasets[2].data = [];
        return;
    }

    // Get the time range from main data
    const startTime = mainData[0].x;
    const endTime = mainData[mainData.length - 1].x;

    // Create horizontal line data points
    const accumulationLineData = [
        { x: startTime, y: convertedDailyTotal },
        { x: endTime, y: convertedDailyTotal }
    ];

    chart.data.datasets[2].data = accumulationLineData;
    
    // Adjust Y-axis scale to ensure both incremental rain data and daily total are visible
    // This handles the case where rain data is near 0.0 but daily total is much higher
    if (convertedDailyTotal > 0.001) { // Changed from > 0 to > 0.001 to handle very small values
        const mainDataValues = mainData.map(point => point.y);
        const minDataValue = Math.min(...mainDataValues, 0);
        const maxDataValue = Math.max(...mainDataValues);
        
        // Ensure the scale includes both the data range and the daily total
        const suggestedMin = Math.min(minDataValue, 0);
        const suggestedMax = Math.max(maxDataValue, convertedDailyTotal * 1.1); // Add 10% padding above daily total
        
        chart.options.scales.y.min = suggestedMin;
        chart.options.scales.y.max = suggestedMax;
        
        debugLog(logLevels.DEBUG, 'Rain chart Y-axis adjusted', {
            minData: minDataValue,
            maxData: maxDataValue,
            dailyTotal: convertedDailyTotal,
            scaleMin: suggestedMin,
            scaleMax: suggestedMax
        });
    } else {
        // For very small or zero daily totals, remove Y-axis constraints to allow auto-scaling
        delete chart.options.scales.y.min;
        delete chart.options.scales.y.max;
        
        debugLog(logLevels.DEBUG, 'Rain chart Y-axis reset to auto-scale', {
            dailyTotal: convertedDailyTotal
        });
    }
    
    debugLog(logLevels.DEBUG, '24h accumulation line updated', {
        originalTotal: rainDailyTotal,
        convertedTotal: convertedDailyTotal,
        unit: units?.rain || 'inches',
        dataPoints: accumulationLineData.length,
        lineData: accumulationLineData
    });
}

// Temporary test function to simulate the rain issue (for debugging)
function testRainChartScaling() {
    if (charts.rain && weatherData) {
        // Simulate the scenario: rain data near 0.0 but daily total is 6.724
        const testDailyTotal = 6.724;
        
        debugLog(logLevels.INFO, 'Testing rain chart scaling with simulated data', {
            currentRainData: weatherData.rainAccum,
            currentDailyTotal: weatherData.rainDailyTotal,
            testDailyTotal: testDailyTotal
        });
        
        // Temporarily override the daily total for testing
        const originalDailyTotal = weatherData.rainDailyTotal;
        weatherData.rainDailyTotal = testDailyTotal;
        
        // Update the 24-hour accumulation line with test data
        update24HourAccumulationLine(charts.rain, testDailyTotal, units);
        charts.rain.update();
        
        // Update the display
        const dailyRainElement = document.getElementById('daily-rain-total');
        if (dailyRainElement) {
            const rainUnit = units.rain === 'inches' ? 'in' : 'mm';
            let displayValue = testDailyTotal;
            if (units.rain === 'mm') {
                displayValue = inchesToMm(testDailyTotal);
            }
            dailyRainElement.textContent = displayValue.toFixed(3) + ' ' + rainUnit;
        }
        
        debugLog(logLevels.INFO, 'Rain chart test applied - check the rain card for scaling');
        
        // Restore original value after 10 seconds
        setTimeout(() => {
            weatherData.rainDailyTotal = originalDailyTotal;
            update24HourAccumulationLine(charts.rain, originalDailyTotal, units);
            charts.rain.update();
            
            if (dailyRainElement) {
                const rainUnit = units.rain === 'inches' ? 'in' : 'mm';
                let displayValue = originalDailyTotal;
                if (units.rain === 'mm') {
                    displayValue = inchesToMm(originalDailyTotal);
                }
                dailyRainElement.textContent = displayValue.toFixed(3) + ' ' + rainUnit;
            }
            
            debugLog(logLevels.INFO, 'Rain chart test restored to original values');
        }, 10000);
    }
}

// Make test function available globally
window.testRainChartScaling = testRainChartScaling;

function validateAndSortChartData(chart) {
    // Validate and sort data for the main dataset
    if (chart.data.datasets[0] && chart.data.datasets[0].data) {
        let data = chart.data.datasets[0].data;

        // Filter out invalid data points and normalize timestamps
        data = data.filter(point => {
            if (!point) return false;
            if (typeof point.y !== 'number' || isNaN(point.y)) return false;

            // Accept Date objects, numeric timestamps, or ISO date strings that parse to valid times
            if (point.x instanceof Date) return !isNaN(point.x.getTime());
            if (typeof point.x === 'number') return Number.isFinite(point.x);
            if (typeof point.x === 'string') return !isNaN(Date.parse(point.x));
            return false;
        }).map(point => {
            // Normalize x to a Date object for consistent sorting/rendering
            return { ...point, x: (point.x instanceof Date) ? point.x : new Date(point.x) };
        });

        // Sort by timestamp (ascending)
        data.sort((a, b) => a.x.getTime() - b.x.getTime());

        // Keep all points (no duplicate removal) but ensure chart receives fully-normalized data
        chart.data.datasets[0].data = data;

        debugLog(logLevels.DEBUG, 'Chart data validated, normalized and sorted', {
            points: data.length,
            first: data[0] || null,
            last: data[data.length - 1] || null
        });
    }
}

function updateTrendLine(chart, data) {
    if (data.length < 2) {
        // Clear trend line if insufficient data
        if (chart.data.datasets[2]) {
            chart.data.datasets[2].data = [];
        }
        return;
    }
    
    const trendLineData = calculateTrendLine(data);
    
    // Ensure trend line dataset exists (dataset index 2)
    if (chart.data.datasets[2]) {
        chart.data.datasets[2].data = trendLineData;
    }
    
    debugLog(logLevels.DEBUG, 'Trend line updated', {
        originalDataPoints: data.length,
        trendLinePoints: trendLineData.length,
        slope: trendLineData.length > 1 ? 
            (trendLineData[trendLineData.length - 1].y - trendLineData[0].y) / (trendLineData.length - 1) : 0
    });
}

function getLuxDescription(lux) {
    let description;
    if (lux <= 0.0001) description = "Moonless, overcast night sky (starlight)";
    else if (lux <= 0.002) description = "Moonless clear night sky with airglow";
    else if (lux <= 0.01) description = "Quarter moon on a clear night";
    else if (lux <= 0.3) description = "Full moon on a clear night";
    else if (lux <= 3.4) description = "Dark limit of civil twilight";
    else if (lux <= 50) description = "Public areas with dark surroundings";
    else if (lux <= 80) description = "Family living room lights";
    else if (lux <= 100) description = "Office building hallway/toilet lighting";
    else if (lux <= 150) description = "Very dark overcast day";
    else if (lux <= 400) description = "Train station platforms";
    else if (lux <= 500) description = "Office lighting";
    else if (lux <= 1000) description = "Sunrise or sunset on a clear day";
    else if (lux <= 25000) description = "Overcast day / Full daylight (not direct sun)";
    else if (lux <= 100000) description = "Direct sunlight";
    else description = "Extremely bright conditions";
    
    debugLog(logLevels.DEBUG, 'Lux description calculated', {
        lux: lux,
        description: description
    });
    
    return description;
}

function getHumidityDescription(humidity) {
    let description;
    if (humidity <= 30) description = "Very dry - may cause discomfort";
    else if (humidity <= 40) description = "Dry - comfortable for most people";
    else if (humidity <= 60) description = "Comfortable humidity level";
    else if (humidity <= 70) description = "Slightly humid - still comfortable";
    else if (humidity <= 80) description = "Humid - may feel sticky";
    else description = "Very humid - uncomfortable for most";
    
    debugLog(logLevels.DEBUG, 'Humidity description calculated', {
        humidity: humidity,
        description: description
    });
    
    return description;
}

function getRainDescription(rainAccumMm) {
    // Convert accumulated rainfall to descriptive categories
    // Based on standard meteorological classifications and colorful descriptions
    let description = "";
    
    if (rainAccumMm <= 0) description = "No precipitation ☀️";
    else if (rainAccumMm <= 0.1) description = "Trace - Barely measurable 🌫️";
    else if (rainAccumMm <= 0.5) description = "Very light - Gentle mist 💧";
    else if (rainAccumMm <= 2) description = "Light - Soft drizzle 🌦️";
    else if (rainAccumMm <= 5) description = "Moderate - Steady shower 🌧️";
    else if (rainAccumMm <= 15) description = "Heavy - Strong downpour ⛈️";
    else if (rainAccumMm <= 30) description = "Very heavy - Intense deluge 🌩️";
    else if (rainAccumMm <= 75) description = "Extreme - Torrential rain ⛈️💦";
    else description = "Cats and dogs - Epic deluge! 🐱🐶💧";
    
    debugLog(logLevels.DEBUG, 'Rain description calculated', {
        rainMm: rainAccumMm,
        description: description
    });
    
    return description;
}

function getUVDescription(uvIndex) {
    let description;
    if (uvIndex <= 2) description = "Low - Low danger from the sun's UV rays";
    else if (uvIndex <= 5) description = "Moderate - Moderate risk of harm from unprotected sun exposure";
    else if (uvIndex <= 7) description = "High - High risk of harm, protection needed";
    else if (uvIndex <= 10) description = "Very High - Very high risk, take extra precautions";
    else description = "Extreme - Extreme risk, take all precautions";
    
    debugLog(logLevels.DEBUG, 'UV description calculated', {
        uvIndex: uvIndex,
        description: description
    });
    
    return description;
}

function getPressureDescription(pressure) {
    // Convert to mb if needed
    let pressureMb = pressure;
    if (units.pressure === 'inHg') {
        pressureMb = inHgToMb(pressure);
    }

    let description;
    if (pressureMb < 980) description = "Stormy conditions, heavy precipitation likely";
    else if (pressureMb < 990) description = "Rain expected, possible severe weather";
    else if (pressureMb < 1000) description = "Changeable weather, precipitation possible";
    else if (pressureMb < 1010) description = "Fair weather, improving conditions";
    else if (pressureMb < 1020) description = "Clear and dry weather expected";
    else if (pressureMb < 1030) description = "Very dry, stable high pressure system";
    else description = "Exceptionally dry, very stable conditions";
    
    debugLog(logLevels.DEBUG, 'Pressure description calculated', {
        originalPressure: pressure,
        pressureMb: pressureMb,
        units: units.pressure,
        description: description
    });
    
    return description;
}

function toggleLuxTooltip() {
    const tooltip = document.getElementById('lux-tooltip');
    tooltip.classList.toggle('show');
}

function closeLuxTooltip() {
    const tooltip = document.getElementById('lux-tooltip');
    tooltip.classList.remove('show');
}

function handleLuxTooltipClickOutside(event) {
    const tooltip = document.getElementById('lux-tooltip');
    const context = document.getElementById('lux-context');
    const infoIcon = document.getElementById('lux-info-icon');

    // If tooltip is visible and click is outside the tooltip and info icon
    if (tooltip.classList.contains('show') &&
        !tooltip.contains(event.target) &&
        !infoIcon.contains(event.target)) {
        closeLuxTooltip();
    }
}

function togglePressureTooltip(event) {
    // Stop the click from bubbling up to the parent pressure-unit element
    if (event) {
        event.stopPropagation();
    }
    
    const tooltip = document.getElementById('pressure-tooltip');
    tooltip.classList.toggle('show');
}

function closePressureTooltip() {
    const tooltip = document.getElementById('pressure-tooltip');
    tooltip.classList.remove('show');
}

function handlePressureTooltipClickOutside(event) {
    const tooltip = document.getElementById('pressure-tooltip');
    const context = document.getElementById('pressure-context');
    const infoIcon = document.getElementById('pressure-info-icon');

    // If tooltip is visible and click is outside the tooltip and info icon
    if (tooltip.classList.contains('show') &&
        !tooltip.contains(event.target) &&
        !infoIcon.contains(event.target)) {
        closePressureTooltip();
    }
}

// Tooltip management functions
function toggleLuxTooltip() {
    const tooltip = document.getElementById('lux-tooltip');
    tooltip.classList.toggle('show');
    debugLog(logLevels.DEBUG, 'Lux tooltip toggled', { visible: tooltip.classList.contains('show') });
}

function closeLuxTooltip() {
    const tooltip = document.getElementById('lux-tooltip');
    tooltip.classList.remove('show');
    debugLog(logLevels.DEBUG, 'Lux tooltip closed');
}

function handleLuxTooltipClickOutside(event) {
    const tooltip = document.getElementById('lux-tooltip');
    const infoIcon = document.getElementById('lux-info-icon');
    
    if (tooltip.classList.contains('show') &&
        !tooltip.contains(event.target) &&
        !infoIcon.contains(event.target)) {
        closeLuxTooltip();
    }
}

// Rain tooltip functions
function toggleRainTooltip(event) {
    if (event) {
        event.stopPropagation();
    }
    const tooltip = document.getElementById('rain-tooltip');
    tooltip.classList.toggle('show');
    debugLog(logLevels.DEBUG, 'Rain tooltip toggled', { visible: tooltip.classList.contains('show') });
}

function closeRainTooltip() {
    const tooltip = document.getElementById('rain-tooltip');
    tooltip.classList.remove('show');
    debugLog(logLevels.DEBUG, 'Rain tooltip closed');
}

function handleRainTooltipClickOutside(event) {
    const tooltip = document.getElementById('rain-tooltip');
    const infoIcon = document.getElementById('rain-info-icon');
    
    if (tooltip && tooltip.classList.contains('show') && 
        !tooltip.contains(event.target) && !infoIcon.contains(event.target)) {
        closeRainTooltip();
    }
}

// Humidity tooltip functions
function toggleHumidityTooltip(event) {
    if (event) {
        event.stopPropagation();
    }
    const tooltip = document.getElementById('humidity-tooltip');
    tooltip.classList.toggle('show');
    debugLog(logLevels.DEBUG, 'Humidity tooltip toggled', { visible: tooltip.classList.contains('show') });
}

function closeHumidityTooltip() {
    const tooltip = document.getElementById('humidity-tooltip');
    tooltip.classList.remove('show');
    debugLog(logLevels.DEBUG, 'Humidity tooltip closed');
}

function handleHumidityTooltipClickOutside(event) {
    const tooltip = document.getElementById('humidity-tooltip');
    const infoIcon = document.getElementById('humidity-info-icon');
    
    if (tooltip && tooltip.classList.contains('show') && 
        !tooltip.contains(event.target) && !infoIcon.contains(event.target)) {
        closeHumidityTooltip();
    }
}

function toggleHeatIndexTooltip() {
    const tooltip = document.getElementById('heat-index-tooltip');
    tooltip.classList.toggle('show');
    debugLog(logLevels.DEBUG, 'Heat index tooltip toggled', { visible: tooltip.classList.contains('show') });
}

function closeHeatIndexTooltip() {
    const tooltip = document.getElementById('heat-index-tooltip');
    tooltip.classList.remove('show');
    debugLog(logLevels.DEBUG, 'Heat index tooltip closed');
}

function handleHeatIndexTooltipClickOutside(event) {
    const tooltip = document.getElementById('heat-index-tooltip');
    const infoIcon = document.getElementById('heat-index-info-icon');
    
    if (tooltip.classList.contains('show') &&
        !tooltip.contains(event.target) &&
        !infoIcon.contains(event.target)) {
        closeHeatIndexTooltip();
    }
}

function toggleUVTooltip() {
    const tooltip = document.getElementById('uv-tooltip');
    tooltip.classList.toggle('show');
    debugLog(logLevels.DEBUG, 'UV tooltip toggled', { visible: tooltip.classList.contains('show') });
}

function closeUVTooltip() {
    const tooltip = document.getElementById('uv-tooltip');
    tooltip.classList.remove('show');
    debugLog(logLevels.DEBUG, 'UV tooltip closed');
}

function handleUVTooltipClickOutside(event) {
    const tooltip = document.getElementById('uv-tooltip');
    const infoIcon = document.getElementById('uv-info-icon');
    
    if (tooltip.classList.contains('show') &&
        !tooltip.contains(event.target) &&
        !infoIcon.contains(event.target)) {
        closeUVTooltip();
    }
}

function updateDisplay() {
    if (!weatherData) {
        debugLog(logLevels.WARN, 'updateDisplay called but no weatherData available');
        return;
    }
    
    debugLog(logLevels.DEBUG, 'Updating display with weather data', weatherData);

    // Temperature calculation and display
    document.getElementById('temperature').textContent = formatTemperature(weatherData.temperature);
    debugLog(logLevels.DEBUG, 'Temperature updated', {
        original: weatherData.temperature,
        formatted: formatTemperature(weatherData.temperature),
        unit: units.temperature
    });

    // Humidity and heat index
    document.getElementById('humidity').textContent = weatherData.humidity.toFixed(1);
    document.getElementById('humidity-description').textContent = getHumidityDescription(weatherData.humidity);
    
    // Calculate and display heat index
    const heatIndexC = calculateHeatIndex(weatherData.temperature, weatherData.humidity);
    const heatIndexElement = document.getElementById('heat-index');
    if (heatIndexElement) {
        heatIndexElement.textContent = formatTemperature(heatIndexC);
    }
    debugLog(logLevels.DEBUG, 'Heat index calculated and displayed', {
        heatIndexC: heatIndexC,
        formatted: formatTemperature(heatIndexC),
        unit: units.temperature
    });

    // Wind data
    document.getElementById('wind-speed').textContent = formatWindSpeed(weatherData.windSpeed);

    // Define converted wind variables for logging and display consistency
    let windSpeed = typeof weatherData.windSpeed === 'number' ? weatherData.windSpeed : 0;
    let windGust = typeof weatherData.windGust === 'number' ? weatherData.windGust : 0;
    if (units.wind === 'kph') {
        windSpeed = mphToKph(windSpeed);
        windGust = mphToKph(windGust);
    }

    // Wind gust information
    const windUnit = units.wind === 'kph' ? 'kph' : 'mph';
    if (weatherData.windGust > weatherData.windSpeed) {
        document.getElementById('wind-gust-info').textContent = `Winds gusting to ${formatWindSpeed(weatherData.windGust)}`;
    } else if (weatherData.windGust > 0) {
        document.getElementById('wind-gust-info').textContent = `Gusts up to ${formatWindSpeed(weatherData.windGust)}`;
    } else {
        document.getElementById('wind-gust-info').textContent = 'No gusts detected';
    }

    const direction = degreesToDirection(weatherData.windDirection);
    document.getElementById('wind-direction').textContent = direction + ' (' + weatherData.windDirection.toFixed(0) + '°)';
    document.getElementById('wind-arrow').textContent = updateArrow(direction);
    debugLog(logLevels.DEBUG, 'Wind data updated', {
        originalSpeed: weatherData.windSpeed,
        convertedSpeed: windSpeed,
        originalGust: weatherData.windGust,
        convertedGust: windGust,
        direction: weatherData.windDirection,
        directionText: direction,
        unit: units.wind
    });

    // Rain data
    // Prepare converted values for rain and wind to avoid referencing undefined variables
    // Server provides rain values as inches (incremental). Convert to millimeters for
    // description and formatting functions which expect mm input.
    const rainInInches = typeof weatherData.rainAccum === 'number' ? weatherData.rainAccum : 0;
    const rainMm = inchesToMm(rainInInches);

    // Display current incremental rain (formatRain expects mm input)
    document.getElementById('rain').textContent = formatRain(rainMm);

    // Display daily rain total
    const dailyRainElement = document.getElementById('daily-rain-total');
    const dailyRainInInches = typeof weatherData.rainDailyTotal === 'number' ? weatherData.rainDailyTotal : 0;
    const dailyRainMm = inchesToMm(dailyRainInInches);
    if (dailyRainElement) {
        dailyRainElement.textContent = formatRain(dailyRainMm || 0);
    }

    // Display rain description based on current accumulated rain (in mm)
    const rainDescElement = document.getElementById('rain-description');
    if (rainDescElement) {
        rainDescElement.textContent = getRainDescription(rainMm);
    }

    
    // Precipitation type data
    const precipitationTypeElement = document.getElementById('precipitation-type');
    if (precipitationTypeElement) {
        const precipType = weatherData.precipitationType || 0;
        precipitationTypeElement.textContent = getPrecipitationTypeDescription(precipType);
        
        debugLog(logLevels.DEBUG, 'Updated precipitation type', {
            precipitationType: precipType,
            description: getPrecipitationTypeDescription(precipType)
        });
    }
    
    // Lightning data
    const lightningCountElement = document.getElementById('lightning-count');
    const lightningDistanceElement = document.getElementById('lightning-distance');
    const lightningDistanceUnitElement = document.getElementById('lightning-distance-unit');
    
    if (lightningCountElement) {
        lightningCountElement.textContent = weatherData.lightningStrikeCount || 0;
    }
    
    if (lightningDistanceElement && lightningDistanceUnitElement) {
        let lightningDistance = weatherData.lightningStrikeAvg || 0;
        if (units.rain === 'inches') {
            lightningDistance = kmToMiles(lightningDistance);
            lightningDistanceUnitElement.textContent = 'mi';
        } else {
            lightningDistanceUnitElement.textContent = 'km';
        }
        lightningDistanceElement.textContent = lightningDistance.toFixed(1);
    }
    
    debugLog(logLevels.DEBUG, 'Rain and lightning data updated', {
        originalRain: weatherData.rainAccum,
        convertedRain: rainMm,
        originalDailyRain: weatherData.rainDailyTotal,
        convertedDailyRain: dailyRainMm,
        rainUnit: units.rain,
        rainDescription: getRainDescription(rainMm),
        lightningCount: weatherData.lightningStrikeCount,
        lightningDistance: weatherData.lightningStrikeAvg
    });

    let pressure = weatherData.pressure;
    if (units.pressure === 'inHg') {
        pressure = mbToInHg(pressure);
    }
    document.getElementById('pressure').textContent = formatPressure(weatherData.pressure);
    
    // Use server-provided pressure analysis - AGGRESSIVE DEBUGGING (v3.0)
    const apiCondition = weatherData.pressure_condition;
    const apiTrend = weatherData.pressure_trend;
    const apiForecast = weatherData.weather_forecast;
    
    console.log('� AGGRESSIVE PRESSURE DEBUG:', {
        'Raw API Object': weatherData,
        'Extracted apiCondition': apiCondition,
        'Extracted apiTrend': apiTrend, 
        'Extracted apiForecast': apiForecast,
        'About to set pressure-condition to': apiCondition,
        'About to set pressure-trend to': apiTrend,
        'About to set pressure-forecast to': apiForecast
    });
    
    const conditionElement = document.getElementById('pressure-condition');
    const trendElement = document.getElementById('pressure-trend');
    const forecastElement = document.getElementById('pressure-forecast');
    const seaLevelElement = document.getElementById('pressure-sea-level');
    
    if (conditionElement) conditionElement.textContent = apiCondition || '--';
    if (trendElement) trendElement.textContent = apiTrend || '--';  
    if (forecastElement) forecastElement.textContent = apiForecast || '--';
    
    // Display sea level pressure with unit conversion
    if (seaLevelElement && weatherData.seaLevelPressure) {
        let seaLevelPressure = weatherData.seaLevelPressure;
        let pressureUnit = 'mb';
        
        if (units.pressure === 'inHg') {
            seaLevelPressure = mbToInHg(seaLevelPressure);
            pressureUnit = 'inHg';
        }
        
        seaLevelElement.textContent = `${Math.round(seaLevelPressure)} ${pressureUnit}`;
    } else if (seaLevelElement) {
        seaLevelElement.textContent = '--';
    }
    
    console.log('✅ AFTER SETTING:', {
        'pressure-condition element text': conditionElement ? conditionElement.textContent : 'NOT FOUND',
        'pressure-trend element text': trendElement ? trendElement.textContent : 'NOT FOUND',
        'pressure-forecast element text': forecastElement ? forecastElement.textContent : 'NOT FOUND'
    });

    // Light and UV data
    document.getElementById('illuminance').textContent = weatherData.illuminance.toFixed(0);
    document.getElementById('lux-description').textContent = getLuxDescription(weatherData.illuminance);
    
    const uvElement = document.getElementById('uv-index');
    const uvDescElement = document.getElementById('uv-description');
    if (uvElement) uvElement.textContent = Math.round(weatherData.uv);
    if (uvDescElement) uvDescElement.textContent = getUVDescription(weatherData.uv);
    
    debugLog(logLevels.DEBUG, 'Light and UV data updated', {
        illuminance: weatherData.illuminance,
        luxDescription: getLuxDescription(weatherData.illuminance),
        uv: weatherData.uv,
        uvDescription: getUVDescription(weatherData.uv)
    });

    // Last update timestamp
    const lastUpdateText = new Date(weatherData.lastUpdate).toLocaleString('en-US', {
        year: 'numeric',
        month: '2-digit', 
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
    });
    document.getElementById('last-update').textContent = lastUpdateText;
    
    // Battery level (update from current weather data)
    const tempestBatteryFromWeather = document.getElementById('tempest-battery');
    if (tempestBatteryFromWeather && weatherData.battery !== undefined) {
        if (weatherData.battery === 0) {
            tempestBatteryFromWeather.textContent = 'N/A';
            debugLog(logLevels.DEBUG, '🔋 Battery data not available from API (returned 0)');
        } else {
            tempestBatteryFromWeather.textContent = `${weatherData.battery.toFixed(1)}V`;
            debugLog(logLevels.DEBUG, '🔋 Battery updated from weather data:', `${weatherData.battery.toFixed(1)}V`);
        }
    }
    
    debugLog(logLevels.INFO, 'Display update completed', {
        lastUpdate: weatherData.lastUpdate,
        formattedTime: lastUpdateText
    });
}

function updateCharts() {
    console.log('🚀 DEBUG: updateCharts() called', { weatherData: weatherData });
    
    if (!weatherData) {
        console.warn('⚠️ DEBUG: updateCharts called but no weatherData available');
        debugLog(logLevels.WARN, 'updateCharts called but no weatherData available');
        return;
    }
    
    console.log('🚀 DEBUG: Starting chart updates with data:', weatherData);
    debugLog(logLevels.DEBUG, 'Starting chart updates');
    
    // Use current time for live data updates - this creates real-time progression
    const now = new Date();
    
    // Debug generated weather data freshness
    console.log('📊 CHART UPDATE DEBUG:', {
        currentTime: now.toISOString(),
        weatherLastUpdate: weatherData.lastUpdate,
        temperature: weatherData.temperature,
        timeDiff: now.getTime() - new Date(weatherData.lastUpdate).getTime()
    });

    // Temperature chart (defensive)
    let tempValue = (typeof weatherData.temperature === 'number' && Number.isFinite(weatherData.temperature)) ? weatherData.temperature : 0;
    if (units.temperature === 'fahrenheit') {
        tempValue = celsiusToFahrenheit(tempValue);
    }
    if (charts.temperature && charts.temperature.data && charts.temperature.data.datasets && charts.temperature.data.datasets[0]) {
        charts.temperature.data.datasets[0].data.push({ x: now, y: tempValue });
        if (charts.temperature.data.datasets[0].data.length > maxDataPoints) charts.temperature.data.datasets[0].data.shift();
        const tempAvg = calculateAverage(charts.temperature.data.datasets[0].data);
        validateAndSortChartData(charts.temperature);
        updateAverageLine(charts.temperature, charts.temperature.data.datasets[0].data);
        charts.temperature.options.scales.y.title = {
            display: true,
            text: units.temperature === 'celsius' ? '°C' : '°F'
        };
        try { charts.temperature.update(); } catch (e) { debugLog(logLevels.ERROR, 'Temperature chart update failed', { error: e.message }); }
        debugLog(logLevels.DEBUG, 'Temperature chart updated', { dataPoints: charts.temperature.data.datasets[0].data.length, currentValue: tempValue, average: tempAvg });
    }

    // Humidity chart (defensive)
    const humidityValue = (typeof weatherData.humidity === 'number' && Number.isFinite(weatherData.humidity)) ? weatherData.humidity : 0;
    if (charts.humidity && charts.humidity.data && charts.humidity.data.datasets && charts.humidity.data.datasets[0]) {
        charts.humidity.data.datasets[0].data.push({ x: now, y: humidityValue });
        if (charts.humidity.data.datasets[0].data.length > maxDataPoints) charts.humidity.data.datasets[0].data.shift();
        const humidityAvg = calculateAverage(charts.humidity.data.datasets[0].data);
        updateAverageLine(charts.humidity, charts.humidity.data.datasets[0].data);
        charts.humidity.options.scales.y.title = { display: true, text: '%' };
        try { charts.humidity.update(); } catch (e) { debugLog(logLevels.ERROR, 'Humidity chart update failed', { error: e.message }); }
    }

    // Wind chart (defensive)
    let windValue = (typeof weatherData.windSpeed === 'number' && Number.isFinite(weatherData.windSpeed)) ? weatherData.windSpeed : 0;
    if (units.wind === 'kph') windValue = mphToKph(windValue);
    if (charts.wind && charts.wind.data && charts.wind.data.datasets && charts.wind.data.datasets[0]) {
        charts.wind.data.datasets[0].data.push({ x: now, y: windValue });
        if (charts.wind.data.datasets[0].data.length > maxDataPoints) charts.wind.data.datasets[0].data.shift();
        const windAvg = calculateAverage(charts.wind.data.datasets[0].data);
        updateAverageLine(charts.wind, charts.wind.data.datasets[0].data);
        charts.wind.options.scales.y.title = { display: true, text: units.wind === 'mph' ? 'mph' : 'kph' };
        try { charts.wind.update(); } catch (e) { debugLog(logLevels.ERROR, 'Wind chart update failed', { error: e.message }); }
    }

    // Rain chart (defensive)
    let rainValue = (typeof weatherData.rainAccum === 'number' && Number.isFinite(weatherData.rainAccum)) ? weatherData.rainAccum : 0;
    if (units.rain === 'mm') rainValue = inchesToMm(rainValue);
    if (charts.rain && charts.rain.data && charts.rain.data.datasets && charts.rain.data.datasets[0]) {
        charts.rain.data.datasets[0].data.push({ x: now, y: rainValue });
        if (charts.rain.data.datasets[0].data.length > maxDataPoints) charts.rain.data.datasets[0].data.shift();
        const rainAvg = calculateAverage(charts.rain.data.datasets[0].data);
        updateAverageLine(charts.rain, charts.rain.data.datasets[0].data);
        try { update24HourAccumulationLine(charts.rain, weatherData.rainDailyTotal, units); } catch (e) { debugLog(logLevels.ERROR, 'update24HourAccumulationLine failed', { error: e.message }); }
        charts.rain.options.scales.y.title = { display: true, text: units.rain === 'inches' ? 'in' : 'mm' };
        try { charts.rain.update(); } catch (e) { debugLog(logLevels.ERROR, 'Rain chart update failed', { error: e.message }); }
    }

    // Pressure chart (defensive)
    let pressureValue = (typeof weatherData.pressure === 'number' && Number.isFinite(weatherData.pressure)) ? weatherData.pressure : 0;
    if (units.pressure === 'inHg') pressureValue = mbToInHg(pressureValue);
    if (charts.pressure && charts.pressure.data && charts.pressure.data.datasets && charts.pressure.data.datasets[0]) {
        charts.pressure.data.datasets[0].data.push({ x: now, y: pressureValue });
        if (charts.pressure.data.datasets[0].data.length > maxDataPoints) charts.pressure.data.datasets[0].data.shift();
        const pressureAvg = calculateAverage(charts.pressure.data.datasets[0].data);
        updateAverageLine(charts.pressure, charts.pressure.data.datasets[0].data);
        updateTrendLine(charts.pressure, charts.pressure.data.datasets[0].data);
        charts.pressure.options.scales.y.title = { display: true, text: units.pressure === 'mb' ? 'mb' : 'inHg' };
        try { charts.pressure.update(); } catch (e) { debugLog(logLevels.ERROR, 'Pressure chart update failed', { error: e.message }); }
    }

    // Light chart (defensive)
    const illuminanceValue = (typeof weatherData.illuminance === 'number' && Number.isFinite(weatherData.illuminance)) ? weatherData.illuminance : 0;
    if (charts.light && charts.light.data && charts.light.data.datasets && charts.light.data.datasets[0]) {
        charts.light.data.datasets[0].data.push({ x: now, y: illuminanceValue });
        if (charts.light.data.datasets[0].data.length > maxDataPoints) charts.light.data.datasets[0].data.shift();
        const lightAvg = calculateAverage(charts.light.data.datasets[0].data);
        updateAverageLine(charts.light, charts.light.data.datasets[0].data);
        charts.light.options.scales.y.title = { display: true, text: 'lux' };
        try { if (charts.light) validateAndSortChartData(charts.light); } catch (e) { debugLog(logLevels.ERROR, 'Failed to validate/sort light chart data', { error: e.message }); }
        try { charts.light.update(); } catch (e) { debugLog(logLevels.ERROR, 'Light chart update failed', { error: e.message }); }
        debugLog(logLevels.DEBUG, 'Light chart updated', { dataPoints: charts.light.data.datasets[0].data.length, currentValue: illuminanceValue, average: lightAvg });
    }

    // UV chart - only update if it exists
    if (charts.uv && charts.uv.data && charts.uv.data.datasets && charts.uv.data.datasets[0]) {
        const uvValue = (typeof weatherData.uv === 'number' && Number.isFinite(weatherData.uv)) ? weatherData.uv : 0;
        charts.uv.data.datasets[0].data.push({ x: now, y: uvValue });
        if (charts.uv.data.datasets[0].data.length > maxDataPoints) charts.uv.data.datasets[0].data.shift();
        const uvAvg = calculateAverage(charts.uv.data.datasets[0].data);
        updateAverageLine(charts.uv, charts.uv.data.datasets[0].data);
        charts.uv.options.scales.y.title = { display: true, text: 'UVI' };
        try { validateAndSortChartData(charts.uv); } catch (e) { debugLog(logLevels.ERROR, 'Failed to validate/sort UV chart data', { error: e.message }); }
        try { charts.uv.update(); } catch (e) { debugLog(logLevels.ERROR, 'UV chart update failed', { error: e.message }); }
        debugLog(logLevels.DEBUG, 'UV chart updated', { dataPoints: charts.uv.data.datasets[0].data.length, currentValue: uvValue, average: uvAvg });
    } else {
        debugLog(logLevels.DEBUG, 'UV chart not available, skipping UV update');
    }

    debugLog(logLevels.INFO, 'All charts updated successfully');
}

function recalculateAverages(changedSensor) {
    // Only recalculate data for the sensor that actually changed units
    // This prevents double-conversion issues
    
    // Recalculate temperature data and average
    if (changedSensor === 'temperature' && charts.temperature.data.datasets[0].data.length > 0) {
        charts.temperature.data.datasets[0].data.forEach(point => {
            if (units.temperature === 'fahrenheit') {
                point.y = celsiusToFahrenheit(point.y);
            } else {
                point.y = fahrenheitToCelsius(point.y);
            }
        });
        const tempAvg = calculateAverage(charts.temperature.data.datasets[0].data);
        updateAverageLine(charts.temperature, charts.temperature.data.datasets[0].data);
        charts.temperature.update();
    }

    // Recalculate wind data and average
    if (changedSensor === 'wind' && charts.wind.data.datasets[0].data.length > 0) {
        charts.wind.data.datasets[0].data.forEach(point => {
            if (units.wind === 'kph') {
                point.y = mphToKph(point.y);
            } else {
                point.y = kphToMph(point.y);
            }
        });
        const windAvg = calculateAverage(charts.wind.data.datasets[0].data);
        updateAverageLine(charts.wind, charts.wind.data.datasets[0].data);
        charts.wind.options.scales.y.title = {
            display: true,
            text: units.wind === 'mph' ? 'mph' : 'kph'
        };
        charts.wind.update();
    }

    // Recalculate rain data and average
    if (changedSensor === 'rain' && charts.rain.data.datasets[0].data.length > 0) {
        charts.rain.data.datasets[0].data.forEach(point => {
            if (units.rain === 'mm') {
                point.y = inchesToMm(point.y);
            } else {
                point.y = mmToInches(point.y);
            }
        });
        const rainAvg = calculateAverage(charts.rain.data.datasets[0].data);
        updateAverageLine(charts.rain, charts.rain.data.datasets[0].data);
        // Update 24-hour accumulation line with current units and weatherData
        if (weatherData && weatherData.rainDailyTotal !== undefined) {
            update24HourAccumulationLine(charts.rain, weatherData.rainDailyTotal, units);
        }
        charts.rain.update();
    }

    // Recalculate pressure data and average
    if (changedSensor === 'pressure' && charts.pressure.data.datasets[0].data.length > 0) {
        charts.pressure.data.datasets[0].data.forEach(point => {
            if (units.pressure === 'inHg') {
                point.y = mbToInHg(point.y);
            } else {
                point.y = inHgToMb(point.y);
            }
        });
        const pressureAvg = calculateAverage(charts.pressure.data.datasets[0].data);
        updateAverageLine(charts.pressure, charts.pressure.data.datasets[0].data);
        updateTrendLine(charts.pressure, charts.pressure.data.datasets[0].data);
        charts.pressure.update();
    }

    // Update pressure analysis when units change - AGGRESSIVE DEBUGGING (v3.0)
    if (weatherData) {
        console.log('🔄 UNITS CHANGE - Pressure Update:', weatherData);
        const conditionEl = document.getElementById('pressure-condition');
        const trendEl = document.getElementById('pressure-trend');
        const forecastEl = document.getElementById('pressure-forecast');
        
        if (conditionEl) conditionEl.textContent = weatherData.pressure_condition || '--';
        if (trendEl) trendEl.textContent = weatherData.pressure_trend || '--';
        if (forecastEl) forecastEl.textContent = weatherData.weather_forecast || '--';

        // Update daily rain total when units change
        const dailyRainElement = document.getElementById('daily-rain-total');
        if (dailyRainElement && weatherData.rainDailyTotal !== undefined) {
            let dailyRain = weatherData.rainDailyTotal;
            if (units.rain === 'mm') {
                dailyRain = inchesToMm(dailyRain);
            }
            const rainUnit = units.rain === 'inches' ? 'in' : 'mm';
            dailyRainElement.textContent = dailyRain.toFixed(3) + ' ' + rainUnit;
        }
    }
    
    // Always update chart axis titles to ensure they reflect current units
    // This doesn't affect data conversion, just display labels
    if (charts.wind && charts.wind.options && charts.wind.options.scales) {
        charts.wind.options.scales.y.title = {
            display: true,
            text: units.wind === 'mph' ? 'mph' : 'kph'
        };
        charts.wind.update();
    }
}

// Unit conversion functions
function formatTemperature(celsius) {
    if (units.temperature === 'fahrenheit') {
        return `${(celsius * 9/5 + 32).toFixed(1)}°F`;
    }
    return `${celsius.toFixed(1)}°C`;
}

function formatPressure(mb) {
    if (units.pressure === 'inHg') {
        return `${(mb * 0.02953).toFixed(3)} inHg`;
    }
    return `${mb.toFixed(1)} mb`;
}

function formatWindSpeed(mps) {
    if (units.wind === 'mph') {
        return `${(mps * 2.23694).toFixed(1)} mph`;
    } else if (units.wind === 'kmh') {
        return `${(mps * 3.6).toFixed(1)} km/h`;
    }
    return `${mps.toFixed(1)} m/s`;
}

function formatRain(mm) {
    if (units.rain === 'inches') {
        return `${(mm * 0.0393701).toFixed(2)} in`;
    }
    return `${mm.toFixed(1)} mm`;
}

function formatRainRate(mmPerHour) {
    if (units.rain === 'inches') {
        return `${(mmPerHour * 0.0393701).toFixed(2)} in/hr`;
    }
    return `${mmPerHour.toFixed(1)} mm/hr`;
}

async function fetchWeather() {
    const startTime = performance.now();
    debugLog(logLevels.DEBUG, 'Starting weather API call');
    
    try {
        const response = await fetch('/api/weather');
        const endTime = performance.now();
        const responseTime = endTime - startTime;
        
        debugLog(logLevels.DEBUG, 'Weather API response received', {
            status: response.status,
            statusText: response.statusText,
            responseTime: responseTime.toFixed(2) + 'ms',
            headers: Object.fromEntries(response.headers.entries())
        });
        
        if (response.ok) {
            const rawData = await response.text();
            debugLog(logLevels.DEBUG, 'Raw weather API response', {
                responseLength: rawData.length,
                responsePreview: rawData.substring(0, 200) + (rawData.length > 200 ? '...' : '')
            });
            
            weatherData = JSON.parse(rawData);
            debugLog(logLevels.INFO, 'Weather data successfully parsed', {
                temperature: weatherData.temperature,
                humidity: weatherData.humidity,
                pressure: weatherData.pressure,
                pressure_condition: weatherData.pressure_condition,
                pressure_trend: weatherData.pressure_trend,
                weather_forecast: weatherData.weather_forecast,
                illuminance: weatherData.illuminance,
                uv: weatherData.uv,
                lastUpdate: weatherData.lastUpdate
            });
            
            // Guard updateDisplay to prevent a single UI error from aborting the pipeline
            try {
                updateDisplay();
            } catch (displayErr) {
                console.error('❌ ERROR in updateDisplay:', displayErr);
                debugLog(logLevels.ERROR, 'updateDisplay error', { error: displayErr.message, stack: displayErr.stack });
            }

            // Explicit chart update with error handling
            try {
                console.log('🚀 DEBUG: About to call updateCharts with weatherData:', weatherData);
                updateCharts();
                console.log('🚀 DEBUG: updateCharts completed successfully');
            } catch (error) {
                console.error('❌ ERROR in updateCharts:', error);
                // Log error to server as well
                debugLog(logLevels.ERROR, 'updateCharts error', { error: error.message, stack: error.stack });
            }
            
            document.getElementById('status').textContent = 'Connected to Tempest station';
            document.getElementById('status').style.background = 'rgba(40,  167, 69, 0.1)';
            
            console.log('🚀 DEBUG: fetchWeather completed, calling updateCharts');
            debugLog(logLevels.INFO, 'Weather fetch completed successfully', {
                totalTime: (performance.now() - startTime).toFixed(2) + 'ms'
            });
            // mark initial weather fetch completed for readiness gating
            __weatherFetched = true;
            trySetDashboardReady();
        } else {
            throw new Error(`Weather API error: ${response.status} ${response.statusText}`);
        }
    } catch (error) {
        debugLog(logLevels.ERROR, 'Error fetching weather data', {
            error: error.message,
            stack: error.stack
        });
        
        document.getElementById('status').textContent = 'Disconnected from weather station';
        document.getElementById('status').style.background = 'rgba(220, 53, 69, 0.1)';
    }
}

async function fetchStatus() {
    const startTime = performance.now();
    const responseTime = (performance.now() - startTime).toFixed(2);
    
    try {
        const response = await fetch('/api/status');
        if (response.ok) {
            const status = await response.json();
            // expose raw status JSON string for headless tests to inspect exact payload
            try {
                window.__lastStatusRaw = JSON.stringify(status);
            } catch(e) {
                // ignore
            }
            debugLog(logLevels.DEBUG, 'Status API response received', {
                responseTime: responseTime + 'ms',
                connected: status.connected,
                stationName: status.stationName,
                uptime: status.uptime,
                homekitStatus: status.homekit
            });
            
            updateStatusDisplay(status);
            updateForecastDisplay(status);
            // mark initial status fetch completed for readiness gating
            __statusFetched = true;
            trySetDashboardReady();
        } else {
            throw new Error(`Status API error: ${response.status}`);
        }
    } catch (error) {
        debugLog(logLevels.ERROR, 'Error fetching status', {
            error: error.message,
            responseTime: (performance.now() - startTime).toFixed(2) + 'ms'
        });
    }
}

function updateStatusDisplay(status) {
    debugLog(logLevels.DEBUG, 'Updating status display', status);
    debugLog(logLevels.DEBUG, '🔍 STATUS DEBUG - Full status object:', JSON.stringify(status, null, 2));
    
    // Store status data globally for unit conversions
    statusData = status;
    
    // Update Tempest status
    const tempestStatus = document.getElementById('tempest-status');
    const tempestStation = document.getElementById('tempest-station');
    const tempestStationURL = document.getElementById('tempest-station-url');
    const tempestElevation = document.getElementById('tempest-elevation');
    const tempestLastUpdate = document.getElementById('tempest-last-update');
    const tempestUptime = document.getElementById('tempest-uptime');
    const tempestDataCount = document.getElementById('tempest-data-count');
    const tempestHistoricalRow = document.getElementById('tempest-historical-row');
    const tempestHistoricalCount = document.getElementById('tempest-historical-count');

    // Handle historical data loading progress first
    if (status.historyLoadingProgress && status.historyLoadingProgress.isLoading) {
        // Update status to show we're reading historical data
        if (tempestStatus) {
            tempestStatus.textContent = 'Reading Historical Observations';
            tempestStatus.style.color = '#ffc107'; // Yellow color for loading
        }
    } else {
        // Not loading - show normal connection status
        if (tempestStatus) {
            if (status.generatedWeather && status.generatedWeather.enabled) {
                tempestStatus.textContent = 'Generated';
                tempestStatus.style.color = '#17a2b8'; // Info blue for generated
            } else {
                tempestStatus.textContent = status.connected ? 'Connected' : 'Disconnected';
                tempestStatus.style.color = status.connected ? '#28a745' : '#dc3545';
            }
        }
    }
    
    // Update station name or generated weather location
    if (tempestStation) {
        if (status.generatedWeather && status.generatedWeather.enabled) {
            tempestStation.innerHTML = `<span style="cursor: pointer; color: #007bff; text-decoration: underline;" onclick="regenerateWeather()" title="Click to regenerate with new location/season">${status.generatedWeather.location} (${status.generatedWeather.season})</span>`;
        } else {
            tempestStation.textContent = status.stationName || '--';
        }
    }
    
    // Update station URL
    if (tempestStationURL) {
        if (status.stationURL) {
            // Make the URL clickable and truncate if too long
            // Truncate display label to 15 characters for compact card layout, show full URL on hover
            const maxLabelLen = 15;
            let displayURL = status.stationURL;
            if (displayURL.length > maxLabelLen) {
                displayURL = displayURL.substring(0, maxLabelLen - 1) + '…';
            }
            // Provide full URL in title and aria-label for hover and accessibility
            tempestStationURL.innerHTML = `<a href="${status.stationURL}" target="_blank" style="color: #007bff; text-decoration: none;" title="${status.stationURL}" aria-label="${status.stationURL}">${displayURL}</a>`;
        } else {
            tempestStationURL.textContent = '--';
        }
    }
    
    // Update elevation display with unit conversion
    updateElevationDisplay();
    
    if (tempestLastUpdate) tempestLastUpdate.textContent = status.lastUpdate ? new Date(status.lastUpdate).toLocaleString('en-US', {
        year: 'numeric',
        month: '2-digit', 
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
    }) : '--';
    if (tempestUptime) tempestUptime.textContent = status.uptime || '--';
    
    // Update battery level and device uptime from station status
    const tempestBattery = document.getElementById('tempest-battery');
    
    // Use station status data if available, otherwise fall back to weather data
    if (status.stationStatus && status.stationStatus.batteryVoltage) {
        tempestBattery.textContent = status.stationStatus.batteryVoltage;
        debugLog(logLevels.DEBUG, '🔋 Battery updated from station status:', status.stationStatus.batteryVoltage);
        
        // Update uptime with more accurate device uptime
        if (tempestUptime && status.stationStatus.deviceUptime) {
            tempestUptime.textContent = status.stationStatus.deviceUptime;
            debugLog(logLevels.DEBUG, '⏱️ Device uptime updated from station status:', status.stationStatus.deviceUptime);
        }
    } else if (tempestBattery && status.dataHistory && status.dataHistory.length > 0) {
        // Fallback to weather data
        const latestData = status.dataHistory[status.dataHistory.length - 1];
        if (latestData.battery !== undefined && latestData.battery !== 0) {
            tempestBattery.textContent = `${latestData.battery.toFixed(1)}V`;
            debugLog(logLevels.DEBUG, '🔋 Battery set from weather data:', `${latestData.battery.toFixed(1)}V`);
        } else {
            tempestBattery.textContent = 'N/A';
            debugLog(logLevels.DEBUG, '🔋 Battery data not available');
        }
    } else if (tempestBattery) {
        tempestBattery.textContent = '--';
        debugLog(logLevels.DEBUG, '🔋 No battery data available');
    }
    
    if (tempestDataCount) tempestDataCount.textContent = status.observationCount || '0';
    
    // Handle historical data loading progress
    if (status.historyLoadingProgress && status.historyLoadingProgress.isLoading) {
        // Show and update historical row with progress
        if (tempestHistoricalRow && tempestHistoricalCount) {
            tempestHistoricalRow.style.display = '';
            const progressText = `${status.historyLoadingProgress.currentStep}/${status.historyLoadingProgress.totalSteps}`;
            tempestHistoricalCount.textContent = progressText;
        }
    } else {
        // Show/hide historical data row and update count (normal state)
        if (tempestHistoricalRow && tempestHistoricalCount) {
            if (status.historicalDataLoaded && status.observationCount > 0) {
                tempestHistoricalRow.style.display = '';
                tempestHistoricalCount.textContent = `${status.observationCount} data points`;
            } else {
                tempestHistoricalRow.style.display = 'none';
            }
        }
    }

    // Populate charts with historical data if available
    if (status.dataHistory && status.dataHistory.length > 0) {
        populateChartsWithHistoricalData(status.dataHistory);
    }

    // Update detailed station status from stationStatus data
    updateDetailedStationStatus(status.stationStatus);

    // Update HomeKit status
    const homekitStatus = document.getElementById('homekit-status');
    const homekitAccessories = document.getElementById('homekit-accessories');
    const homekitBridge = document.getElementById('homekit-bridge');
    const homekitPin = document.getElementById('homekit-pin');

    const hk = status.homekit || {};
    if (homekitStatus) {
        homekitStatus.textContent = hk.bridge ? 'Active' : 'Disabled';
        homekitStatus.style.color = hk.bridge ? '#28a745' : '#dc3545';
    }
    if (homekitAccessories) homekitAccessories.textContent = hk.accessories || '--';
    if (homekitBridge) homekitBridge.textContent = hk.name || '--';
    if (homekitPin) homekitPin.textContent = hk.pin || '--';

    // Update accessories list
    updateAccessoriesList(hk.accessoryNames || []);
    
    debugLog(logLevels.DEBUG, 'Status display update completed', {
        tempestConnected: status.connected,
        homekitActive: hk.bridge,
        accessoryCount: hk.accessories
    });
}

function updateDetailedStationStatus(stationStatus) {
    debugLog(logLevels.DEBUG, 'updateDetailedStationStatus called', stationStatus);
    
    // Device Status Fields
    const deviceUptime = document.getElementById('tempest-device-uptime');
    const deviceNetwork = document.getElementById('tempest-device-network');
    const deviceSignal = document.getElementById('tempest-device-signal');
    const deviceLastObs = document.getElementById('tempest-device-last-obs');
    const deviceSerial = document.getElementById('tempest-device-serial');
    const deviceFirmware = document.getElementById('tempest-device-firmware');
    const sensorStatus = document.getElementById('tempest-sensor-status');
    const batteryStatus = document.getElementById('tempest-battery-status');

    // Hub Status Fields
    const hubUptime = document.getElementById('tempest-hub-uptime');
    const hubNetwork = document.getElementById('tempest-hub-network');
    const hubWifi = document.getElementById('tempest-hub-wifi');
    const hubLastStatus = document.getElementById('tempest-hub-last-status');
    const hubSerial = document.getElementById('tempest-hub-serial');
    const hubFirmware = document.getElementById('tempest-hub-firmware');

    // Data Source Field
    const dataSource = document.getElementById('tempest-data-source');

    // Debug: Check which fields are missing
    console.log('🔍 Field Availability Check:', {
        'deviceLastObs exists': !!deviceLastObs,
        'deviceSerial exists': !!deviceSerial,
        'deviceFirmware exists': !!deviceFirmware,
        'hubLastStatus exists': !!hubLastStatus,
        'hubSerial exists': !!hubSerial,
        'hubFirmware exists': !!hubFirmware
    });

    if (stationStatus && stationStatus.batteryVoltage) {
        // Update Data Source field
        if (dataSource) dataSource.textContent = stationStatus.dataSource || 'api';
        
        // Update Device Status fields from actual station status
        if (deviceUptime) deviceUptime.textContent = stationStatus.deviceUptime || '--';
        if (deviceNetwork) deviceNetwork.textContent = stationStatus.deviceNetworkStatus || '--';
        if (deviceSignal) deviceSignal.textContent = stationStatus.deviceSignal || '--';
        if (deviceLastObs) deviceLastObs.textContent = stationStatus.deviceLastObs || '--';
        if (deviceSerial) deviceSerial.textContent = stationStatus.deviceSerialNumber || '--';
        if (deviceFirmware) deviceFirmware.textContent = stationStatus.deviceFirmware || '--';
        if (sensorStatus) sensorStatus.textContent = stationStatus.sensorStatus || '--';
        if (batteryStatus) batteryStatus.textContent = stationStatus.batteryStatus || '--';

        // Update Hub Status fields from actual station status
        if (hubUptime) hubUptime.textContent = stationStatus.hubUptime || '--';
        if (hubNetwork) hubNetwork.textContent = stationStatus.hubNetworkStatus || '--';
        if (hubWifi) hubWifi.textContent = stationStatus.hubWiFiSignal || '--';
        if (hubLastStatus) hubLastStatus.textContent = stationStatus.hubLastStatus || '--';
        if (hubSerial) hubSerial.textContent = stationStatus.hubSerialNumber || '--';
        if (hubFirmware) hubFirmware.textContent = stationStatus.hubFirmware || '--';

        debugLog(logLevels.DEBUG, 'Detailed station status updated from stationStatus data');
    } else {
        // Update Data Source field for fallback case
        if (dataSource) dataSource.textContent = 'api';
        
        // Use "--" for all fields when station status is not available
        const allStatusFields = [
            deviceUptime, deviceNetwork, deviceSignal, deviceLastObs, deviceSerial, deviceFirmware,
            sensorStatus, batteryStatus, hubUptime, hubNetwork, hubWifi, hubLastStatus, hubSerial, hubFirmware
        ];
        
        allStatusFields.forEach((field, index) => {
            if (field) {
                field.textContent = '--';
            }
        });
        
        debugLog(logLevels.DEBUG, 'Station status data unavailable, using "--" for all detailed status fields');
    }
}

function updateForecastDisplay(status) {
    debugLog(logLevels.DEBUG, 'Updating forecast display', status.forecast);
    
    if (!status.forecast) {
        debugLog(logLevels.DEBUG, 'No forecast data available');
        return;
    }

    const forecast = status.forecast;
    
    // Store forecast data globally for unit conversions
    forecastData = forecast;
    
    // Update current conditions
    updateCurrentConditions(forecast.current_conditions);
    
    // Update daily forecast
    updateDailyForecast(forecast.forecast.daily);
}

function refreshForecastDisplay() {
    // Refresh forecast display with current units (called when units are toggled)
    if (!forecastData) {
        debugLog(logLevels.DEBUG, 'No cached forecast data available for refresh');
        return;
    }
    
    debugLog(logLevels.DEBUG, 'Refreshing forecast display with current units');
    
    // Update current conditions with current units
    updateCurrentConditions(forecastData.current_conditions);
    
    // Update daily forecast with current units  
    updateDailyForecast(forecastData.forecast.daily);
}

function updateCurrentConditions(current) {
    debugLog(logLevels.DEBUG, 'Updating current conditions with data:', current);
    
    const elements = {
        icon: document.getElementById('forecast-current-icon'),
        temp: document.getElementById('forecast-current-temp'),
        feelsLike: document.getElementById('forecast-current-feels-like'),
        conditions: document.getElementById('forecast-current-conditions'),
        humidity: document.getElementById('forecast-current-humidity'),
        wind: document.getElementById('forecast-current-wind'),
        pressure: document.getElementById('forecast-current-pressure'),
        precip: document.getElementById('forecast-current-precip')
    };

    // Convert temperatures based on current unit setting
    let currentTemp = current.air_temperature;
    let feelsLikeTemp = current.feels_like;
    let tempUnit = '°C';
    
    debugLog(logLevels.DEBUG, 'Raw temperature values from API:', {
        air_temperature: current.air_temperature,
        feels_like: current.feels_like,
        relative_humidity: current.relative_humidity
    });
    
    if (units.temperature === 'fahrenheit') {
        currentTemp = celsiusToFahrenheit(currentTemp);
        feelsLikeTemp = celsiusToFahrenheit(feelsLikeTemp);
        tempUnit = '°F';
    }
    
    // Convert pressure based on current unit setting
    let pressure = current.sea_level_pressure;
    let pressureUnit = 'mb';
    
    if (units.pressure === 'inHg') {
        pressure = mbToInHg(pressure);
        pressureUnit = 'inHg';
    }

    if (elements.icon) elements.icon.textContent = getWeatherIcon(current.icon);
    if (elements.temp) elements.temp.textContent = `${Math.round(currentTemp)}${tempUnit}`;
    if (elements.feelsLike) elements.feelsLike.textContent = `${Math.round(feelsLikeTemp)}${tempUnit}`;
    if (elements.conditions) elements.conditions.textContent = current.conditions;
    if (elements.humidity) elements.humidity.textContent = `${current.relative_humidity}%`;
    if (elements.wind) elements.wind.textContent = `${Math.round(current.wind_avg)} mph`;
    if (elements.pressure) elements.pressure.textContent = `${Math.round(pressure)} ${pressureUnit}`;
    if (elements.precip) elements.precip.textContent = `${current.precip_probability}%`;
}

function updateDailyForecast(dailyForecast) {
    const container = document.getElementById('forecast-daily-container');
    if (!container || !dailyForecast) return;

    // Clear existing forecast items
    container.innerHTML = '';

    // Show first 5 days
    const daysToShow = Math.min(5, dailyForecast.length);
    
    for (let i = 0; i < daysToShow; i++) {
        const day = dailyForecast[i];
        
        debugLog(logLevels.DEBUG, `Daily forecast day ${i} raw data:`, {
            air_temperature: day.air_temperature,
            air_temp_high: day.air_temp_high,
            air_temp_low: day.air_temp_low,
            relative_humidity: day.relative_humidity,
            time: day.time,
            conditions: day.conditions
        });
        
        const forecastDay = document.createElement('div');
        forecastDay.className = 'forecast-day';
        
        // Calculate the date for this forecast day
        // If the API provides correct timestamps for each day, use them
        // Otherwise, fall back to calculating from today's date
        let date;
        if (i === 0) {
            // Today - use current date
            date = new Date();
        } else {
            // Try using the API timestamp first
            date = new Date(day.time * 1000);
            // Check if this timestamp seems wrong (same as first day or invalid)
            if (i > 0 && dailyForecast[0] && Math.abs(day.time - dailyForecast[0].time) < 3600) {
                // Timestamps are too similar (less than 1 hour difference), calculate manually
                date = new Date();
                date.setDate(date.getDate() + i);
            }
        }
        
        const dayName = i === 0 ? 'Today' : date.toLocaleDateString('en-US', { weekday: 'short' });
        
        // Convert temperature based on current unit setting
        // Use high/low temperatures if available, otherwise use air_temperature as high and calculate reasonable low
        let highTemp, lowTemp;
        
        if (day.air_temp_high && day.air_temp_low) {
            // Use actual high/low values from API
            highTemp = day.air_temp_high;
            lowTemp = day.air_temp_low;
        } else {
            // Fallback: use air_temperature as average and estimate high/low
            // This is a reasonable estimate based on typical daily temperature ranges
            const avgTemp = day.air_temperature;
            highTemp = avgTemp + 3; // Add 3°C for high
            lowTemp = avgTemp - 3;  // Subtract 3°C for low
        }
        
        debugLog(logLevels.DEBUG, `Day ${i} temperature conversion:`, {
            before_conversion: { high: highTemp, low: lowTemp, unit: 'C' },
            units_setting: units.temperature
        });
        
        let tempUnit = '°C';
        
        if (units.temperature === 'fahrenheit') {
            highTemp = celsiusToFahrenheit(highTemp);
            lowTemp = celsiusToFahrenheit(lowTemp);
            tempUnit = '°F';
        }
        
        debugLog(logLevels.DEBUG, `Day ${i} after conversion:`, {
            after_conversion: { high: Math.round(highTemp), low: Math.round(lowTemp), unit: tempUnit }
        });
        
        forecastDay.innerHTML = `
            <div class="forecast-day-name">${dayName}</div>
            <div class="forecast-day-icon">${getWeatherIcon(day.icon)}</div>
            <div class="forecast-day-conditions">${day.conditions}</div>
            <div class="forecast-day-temps">
                <span class="forecast-day-low">/${Math.round(lowTemp)}${tempUnit}</span>
            </div>
            <div class="forecast-day-precip">${day.precip_probability}%</div>
        `;
        
        container.appendChild(forecastDay);
    }
}

function getWeatherIcon(iconCode) {
    const iconMap = {
        'clear-day': '☀️',
        'clear-night': '🌙',
        'partly-cloudy-day': '⛅',
        'partly-cloudy-night': '☁️',
        'cloudy': '☁️',
        'rain': '🌧️',
        'snow': '❄️',
        'sleet': '🌨️',
        'wind': '💨',
        'fog': '🌫️',
        'thunderstorm': '⛈️'
    };
    
    return iconMap[iconCode] || '🌤️';
}

function populateChartsWithHistoricalData(dataHistory) {
    debugLog(logLevels.DEBUG, 'Populating charts with historical data', {
        dataPoints: dataHistory.length
    });

    // Check if we have any historical data with actual timestamps
    const hasActualTimestamps = dataHistory.some(obs => obs.lastUpdate);
    const currentDataLength = charts.temperature.data.datasets[0].data.length;
    
    // Always process historical data if it has actual timestamps (real weather data)
    // or if charts are completely empty (generated weather data)
    const shouldPopulate = hasActualTimestamps || currentDataLength === 0;
    
    if (shouldPopulate) {
        debugLog(logLevels.INFO, 'Processing historical data', {
            reason: hasActualTimestamps ? 'has actual timestamps' : 'charts empty',
            currentDataPoints: currentDataLength,
            historicalDataPoints: dataHistory.length
        });
        
        // Clear existing chart data to populate with historical data
        charts.temperature.data.datasets[0].data = [];
        charts.humidity.data.datasets[0].data = [];
        charts.wind.data.datasets[0].data = [];
        charts.rain.data.datasets[0].data = [];
        charts.pressure.data.datasets[0].data = [];
        charts.light.data.datasets[0].data = [];
        if (charts.uv && charts.uv.data) {
            charts.uv.data.datasets[0].data = [];
        }
    } else {
        debugLog(logLevels.INFO, 'Skipping historical data population - charts already have live data without timestamps');
        return; // Don't overwrite existing live data
    }

    // Process each historical data point
    for (let i = 0; i < dataHistory.length; i++) {
        const obs = dataHistory[i];
        
        // Use actual timestamp from historical data if available, otherwise create backwards timeline
        let timestamp;
        if (obs.lastUpdate) {
            // Use the actual historical timestamp
            timestamp = new Date(obs.lastUpdate);
        } else {
            // Fallback to backwards timeline for generated data (for compatibility)
            const now = new Date();
            const secondsBack = (dataHistory.length - i - 1) * 10; // 10 seconds between each point
            timestamp = new Date(now.getTime() - (secondsBack * 1000));
        }

        debugLog(logLevels.DEBUG, `Historical data point ${i + 1}/${dataHistory.length}`, {
            timestamp: timestamp.toISOString(),
            hasActualTimestamp: !!obs.lastUpdate,
            temperature: obs.temperature,
            rain: obs.rainAccum
        });

        // Defensive normalization for historical observation fields
        const safeNumber = (v, fallback = 0) => (typeof v === 'number' && Number.isFinite(v)) ? v : fallback;

        // Temperature
        let tempValue = safeNumber(obs.temperature, 0);
        try { if (units.temperature === 'fahrenheit') tempValue = celsiusToFahrenheit(tempValue); } catch (e) { debugLog(logLevels.ERROR, 'Temperature conversion failed for historical point', { error: e.message }); }
        if (charts.temperature && charts.temperature.data && charts.temperature.data.datasets && charts.temperature.data.datasets[0]) {
            charts.temperature.data.datasets[0].data.push({ x: timestamp, y: tempValue });
        }

        // Humidity
        const humidityVal = safeNumber(obs.humidity, 0);
        if (charts.humidity && charts.humidity.data && charts.humidity.data.datasets && charts.humidity.data.datasets[0]) {
            charts.humidity.data.datasets[0].data.push({ x: timestamp, y: humidityVal });
        }

        // Wind
        let windValue = safeNumber(obs.windSpeed, 0);
        try { if (units.wind === 'kph') windValue = mphToKph(windValue); } catch (e) { debugLog(logLevels.ERROR, 'Wind conversion failed for historical point', { error: e.message }); }
        if (charts.wind && charts.wind.data && charts.wind.data.datasets && charts.wind.data.datasets[0]) {
            charts.wind.data.datasets[0].data.push({ x: timestamp, y: windValue });
        }

        // Rain
        let rainValue = safeNumber(obs.rainAccum, 0);
        try { if (units.rain === 'mm') rainValue = inchesToMm(rainValue); } catch (e) { debugLog(logLevels.ERROR, 'Rain conversion failed for historical point', { error: e.message }); }
        if (charts.rain && charts.rain.data && charts.rain.data.datasets && charts.rain.data.datasets[0]) {
            charts.rain.data.datasets[0].data.push({ x: timestamp, y: rainValue });
        }

        // Pressure
        let pressureValue = safeNumber(obs.pressure, 0);
        try { if (units.pressure === 'inHg') pressureValue = mbToInHg(pressureValue); } catch (e) { debugLog(logLevels.ERROR, 'Pressure conversion failed for historical point', { error: e.message }); }
        if (charts.pressure && charts.pressure.data && charts.pressure.data.datasets && charts.pressure.data.datasets[0]) {
            charts.pressure.data.datasets[0].data.push({ x: timestamp, y: pressureValue });
        }

        // Light
        const illumVal = safeNumber(obs.illuminance, 0);
        if (charts.light && charts.light.data && charts.light.data.datasets && charts.light.data.datasets[0]) {
            charts.light.data.datasets[0].data.push({ x: timestamp, y: illumVal });
        }

        // UV (if chart exists)
        const uvVal = safeNumber(obs.uv, 0);
        if (charts.uv && charts.uv.data && charts.uv.data.datasets && charts.uv.data.datasets[0]) {
            charts.uv.data.datasets[0].data.push({ x: timestamp, y: uvVal });
        }
    }

    // Ensure all datasets are validated and sorted after population
    ['temperature','humidity','wind','rain','pressure','light','uv'].forEach(name => {
        const chart = charts[name];
        if (chart && chart.data && chart.data.datasets && chart.data.datasets[0]) {
            validateAndSortChartData(chart);
            // Update average/trend where applicable
            if (chart === charts.temperature || chart === charts.humidity || chart === charts.wind || chart === charts.rain || chart === charts.pressure) {
                updateAverageLine(chart, chart.data.datasets[0].data);
                if (chart === charts.pressure) updateTrendLine(chart, chart.data.datasets[0].data);
            }
        }
    });

    // Update averages and trend lines for all charts
    updateAverageAndTrendLines();

    // Update all charts
    charts.temperature.update();
    charts.humidity.update();
    charts.wind.update();
    charts.rain.update();
    charts.pressure.update();
    charts.light.update();
    if (charts.uv && charts.uv.data) {
        charts.uv.update();
    }

    debugLog(logLevels.INFO, 'Charts populated with historical data', {
        temperaturePoints: charts.temperature.data.datasets[0].data.length,
        humidityPoints: charts.humidity.data.datasets[0].data.length,
        windPoints: charts.wind.data.datasets[0].data.length,
        rainPoints: charts.rain.data.datasets[0].data.length,
        pressurePoints: charts.pressure.data.datasets[0].data.length,
        lightPoints: charts.light.data.datasets[0].data.length,
        uvPoints: charts.uv && charts.uv.data ? charts.uv.data.datasets[0].data.length : 0
    });
}

function updateAverageAndTrendLines() {
    // Update temperature average
    if (charts.temperature.data.datasets[0].data.length > 0) {
        validateAndSortChartData(charts.temperature);
        const tempAvg = calculateAverage(charts.temperature.data.datasets[0].data);
        updateAverageLine(charts.temperature, charts.temperature.data.datasets[0].data);
    }

    // Update humidity average
    if (charts.humidity.data.datasets[0].data.length > 0) {
        validateAndSortChartData(charts.humidity);
        const humidityAvg = calculateAverage(charts.humidity.data.datasets[0].data);
        updateAverageLine(charts.humidity, charts.humidity.data.datasets[0].data);
    }

    // Update wind average
    if (charts.wind.data.datasets[0].data.length > 0) {
        debugLog(logLevels.INFO, '🌬️ BEFORE wind validation - data points:', charts.wind.data.datasets[0].data.length);
        validateAndSortChartData(charts.wind);
        debugLog(logLevels.INFO, '🌬️ AFTER wind validation - data points:', charts.wind.data.datasets[0].data.length);
        const windAvg = calculateAverage(charts.wind.data.datasets[0].data);
        debugLog(logLevels.INFO, '🌬️ BEFORE wind updateAverageLine - avg:', windAvg);
        updateAverageLine(charts.wind, charts.wind.data.datasets[0].data);
        debugLog(logLevels.INFO, '🌬️ AFTER wind updateAverageLine - avg points:', charts.wind.data.datasets[1].data.length);
    }

    // Update rain average
    if (charts.rain.data.datasets[0].data.length > 0) {
        validateAndSortChartData(charts.rain);
        const rainAvg = calculateAverage(charts.rain.data.datasets[0].data);
        updateAverageLine(charts.rain, charts.rain.data.datasets[0].data);
        // Update 24-hour accumulation line if weatherData is available
        if (weatherData && weatherData.rainDailyTotal !== undefined) {
            update24HourAccumulationLine(charts.rain, weatherData.rainDailyTotal, units);
        }
    }

    // Update pressure average and trend
    if (charts.pressure.data.datasets[0].data.length > 0) {
        validateAndSortChartData(charts.pressure);
        const pressureAvg = calculateAverage(charts.pressure.data.datasets[0].data);
        updateAverageLine(charts.pressure, charts.pressure.data.datasets[0].data);
        updateTrendLine(charts.pressure, charts.pressure.data.datasets[0].data);
    }

    // Light data naturally goes to zero at night - no average needed

    // UV data naturally goes to zero at night - no average needed
}

// Continue with HomeKit status update
function updateHomekitStatus(status) {
    // Update HomeKit status
    const homekitStatus = document.getElementById('homekit-status');
    const homekitAccessories = document.getElementById('homekit-accessories');
    const homekitBridge = document.getElementById('homekit-bridge');
    const homekitPin = document.getElementById('homekit-pin');

    const hk = status.homekit || {};
    if (homekitStatus) {
        homekitStatus.textContent = hk.bridge ? 'Active' : 'Inactive';
        homekitStatus.style.color = hk.bridge ? '#28a745' : '#dc3545';
    }
    if (homekitAccessories) homekitAccessories.textContent = hk.accessories || '--';
    if (homekitBridge) homekitBridge.textContent = hk.name || '--';
    if (homekitPin) homekitPin.textContent = hk.pin || '--';

    // Update accessories list
    updateAccessoriesList(hk.accessoryNames || []);
    
    debugLog(logLevels.DEBUG, 'Status display update completed', {
        tempestConnected: status.connected,
        homekitActive: hk.bridge,
        accessoryCount: hk.accessories
    });
}

function updateAccessoriesList(accessoryNames) {
    debugLog(logLevels.DEBUG, 'Updating accessories list', { accessoryNames });
    
    const accessoriesList = document.getElementById('accessories-list');
    if (!accessoriesList) {
        debugLog(logLevels.WARN, 'Accessories list element not found');
        return;
    }
    
    accessoriesList.innerHTML = '';

    if (!accessoryNames || accessoryNames.length === 0) {
        accessoriesList.innerHTML = '<div class="accessory-item">No accessories available</div>';
        debugLog(logLevels.DEBUG, 'No accessories to display');
        return;
    }

    // Define all possible sensors with their icons
    const allSensors = [
        { name: 'Temperature', icon: '🌡️', key: 'Temperature' },
        { name: 'Humidity', icon: '💧', key: 'Humidity' },
        { name: 'Light', icon: '☀️', key: 'Light' },
        { name: 'UV Index', icon: '🌞', key: 'UV' },
        { name: 'Wind Speed', icon: '🌬️', key: 'Wind Speed' },
        { name: 'Wind Direction', icon: '🧭', key: 'Wind Direction' },
        { name: 'Rain', icon: '🌧️', key: 'Rain' },
        { name: 'Pressure', icon: '📊', key: 'Pressure' },
        { name: 'Lightning', icon: '⚡', key: 'Lightning' }
    ];

    // Determine which sensors are enabled based on accessoryNames
    const enabledSensors = [];
    const disabledSensors = [];

    allSensors.forEach(sensor => {
        const isEnabled = accessoryNames && accessoryNames.some(name => name.includes(sensor.key));
        if (isEnabled) {
            enabledSensors.push({ ...sensor, enabled: true });
        } else {
            disabledSensors.push({ ...sensor, enabled: false });
        }
    });

    // Sort enabled sensors to the top, then disabled
    const sortedSensors = [...enabledSensors, ...disabledSensors];

    if (sortedSensors.length === 0) {
        accessoriesList.innerHTML = '<div class="accessory-item">No sensors configured</div>';
        return;
    }

    sortedSensors.forEach(sensor => {
        const accessoryDiv = document.createElement('div');
        accessoryDiv.className = 'accessory-item' + (sensor.enabled ? '' : ' disabled');

        const statusClass = sensor.enabled ? 'enabled' : 'disabled';
        const statusText = sensor.enabled ? 'Active' : 'Disabled';
        const nameClass = sensor.enabled ? '' : ' disabled';

        accessoryDiv.innerHTML = 
            '<span class="accessory-icon">' + sensor.icon + '</span>' +
            '<span class="accessory-name' + nameClass + '">' + sensor.name + '</span>' +
            '<span class="accessory-status ' + statusClass + '">' + statusText + '</span>';

        accessoriesList.appendChild(accessoryDiv);
    });
    
    debugLog(logLevels.DEBUG, 'Accessories list updated', {
        totalSensors: sortedSensors.length,
        enabled: enabledSensors.length,
        disabled: disabledSensors.length
    });
}

function toggleAccessoriesExpansion() {
    const expandedDiv = document.getElementById('accessories-expanded');
    const expandIcon = document.getElementById('accessories-expand-icon');

    if (expandedDiv && expandIcon) {
        const isExpanded = expandedDiv.style.display !== 'none' && expandedDiv.style.display !== '';
        
        if (!isExpanded) {
            expandedDiv.style.display = 'block';
            expandIcon.textContent = '▼';
        } else {
            expandedDiv.style.display = 'none';
            expandIcon.textContent = '▶';
        }
        
        debugLog(logLevels.DEBUG, 'Accessories expansion toggled', { expanded: !isExpanded });
    }
}

// Enhanced DOM ready checking and event listener attachment with debug logging
function attachEventListener(elementId, event, handler, description = '') {
    const element = document.getElementById(elementId);
    if (element) {
        element.addEventListener(event, handler);
        debugLog(logLevels.DEBUG, `Event listener attached: ${elementId}`, {
            event: event,
            description: description
        });
    } else {
        debugLog(logLevels.WARN, `Element not found for event listener: ${elementId}`, {
            description: description,
            retrying: true
        });
        
        // Use MutationObserver to wait for the element to appear
        const observer = new MutationObserver((mutations, obs) => {
            const targetElement = document.getElementById(elementId);
            if (targetElement) {
                targetElement.addEventListener(event, handler);
                debugLog(logLevels.INFO, `Event listener attached via MutationObserver: ${elementId}`, {
                    event: event,
                    description: description
                });
                obs.disconnect();
                return;
            }
        });
        
        observer.observe(document.body, {
            childList: true,
            subtree: true
        });
        
        // Fallback: Try again after delays with more attempts
        let retryCount = 0;
        const maxRetries = 10;
        const retryDelay = 300; // 300ms delay
        
        const retryAttachment = () => {
            retryCount++;
            const retryElement = document.getElementById(elementId);
            if (retryElement) {
                retryElement.addEventListener(event, handler);
                debugLog(logLevels.INFO, `Event listener attached on retry ${retryCount}: ${elementId}`, {
                    event: event,
                    description: description
                });
                observer.disconnect(); // Stop observing if we succeed
            } else if (retryCount < maxRetries) {
                debugLog(logLevels.WARN, `Retry ${retryCount}/${maxRetries} failed for: ${elementId}`, {
                    nextRetryIn: retryDelay + 'ms',
                    allElements: Array.from(document.querySelectorAll('[id]')).map(el => el.id)
                });
                setTimeout(retryAttachment, retryDelay);
            } else {
                debugLog(logLevels.ERROR, `Failed to attach event listener after ${maxRetries} retries: ${elementId}`, {
                    event: event,
                    description: description,
                    elementExists: !!document.getElementById(elementId),
                    bodyHTML: document.body.innerHTML.includes(elementId),
                    searchResult: document.body.innerHTML.indexOf(elementId)
                });
                observer.disconnect(); // Stop observing after max retries
            }
        };
        
        setTimeout(retryAttachment, retryDelay);
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    debugLog(logLevels.INFO, 'DOM Content Loaded - Initializing dashboard');
    
    // Check if debug mode should be enabled from URL or localStorage
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('loglevel') === 'debug') {
        localStorage.setItem('loglevel', 'debug');
        DEBUG_MODE = true;
        debugLog(logLevels.INFO, 'Debug mode enabled via URL parameter');
    }
    
    // IMMEDIATE DOM INSPECTION for pressure-info-icon
    console.log('🔍 IMMEDIATE DOM CHECK:');
    console.log('- pressure-info-icon exists:', !!document.getElementById('pressure-info-icon'));
    console.log('- pressure-card exists:', !!document.getElementById('pressure-card'));
    console.log('- pressure-unit exists:', !!document.getElementById('pressure-unit'));
    
    const pressureUnit = document.getElementById('pressure-unit');
    if (pressureUnit) {
        console.log('- pressure-unit innerHTML:', pressureUnit.innerHTML);
        console.log('- pressure-unit children:', pressureUnit.children.length);
        for (let i = 0; i < pressureUnit.children.length; i++) {
            const child = pressureUnit.children[i];
            console.log(`  Child ${i}:`, child.tagName, child.id, child.className);
        }
    } else {
        console.log('- pressure-unit element not found!');
    }
    
    // Check all elements with "pressure" in ID
    const pressureElements = Array.from(document.querySelectorAll('[id*="pressure"]'));
    console.log('- All pressure-related elements:', pressureElements.map(el => el.id));
    
    // Check for info-icon class elements
    const infoIcons = Array.from(document.querySelectorAll('.info-icon'));
    console.log('- All info-icon elements:', infoIcons.map(el => el.id || el.tagName));
    
    console.log('🔍 DOM CHECK COMPLETE');
    
    // Load units configuration from server first, then update units
    loadUnitsConfig().then(() => {
        updateUnits();
    });

    // Start data fetching immediately so the status UI updates even if Chart.js
    // is unavailable or takes time to load. This prevents a ReferenceError in
    // initCharts (when Chart is undefined) from blocking network calls and
    // leaving the page stuck on "Connecting to weather station...".
    try {
        console.log('🚀 DEBUG: Starting initial data fetch (before charts)');
        fetchWeather();
        fetchStatus();
    } catch (e) {
        debugLog(logLevels.ERROR, 'Error triggering initial fetches', e);
    }
    
    // Ensure pressure info icon has proper event listener attached initially
    const initialPressureInfoIcon = document.getElementById('pressure-info-icon');
    if (initialPressureInfoIcon) {
        initialPressureInfoIcon.addEventListener('click', function(e) {
            e.stopPropagation();
            togglePressureTooltip(e);
        });
        console.log('🔧 Initial setup - Attached click event listener to pressure info icon');
    }
    
    // Initialize charts, but be resilient if Chart.js is not yet loaded.
    // If Chart is undefined, attempt to load the local vendored copy and
    // initialize charts when it finishes loading. Meanwhile, data fetches
    // are already running so the UI won't be blocked.
    (function initChartsResilient() {
        if (__chartsInitialized) return; // already initialized successfully

        __chartInitAttempts++;
        if (__chartInitAttempts > 8) {
            debugLog(logLevels.ERROR, 'Aborting chart initialization after multiple failed attempts', { attempts: __chartInitAttempts });
            return;
        }
        console.log('🚀 DEBUG: Attempting chart initialization (resilient) attempt=' + __chartInitAttempts);
        if (typeof Chart !== 'undefined' && !__chartsInitialized) {
            try {
                initCharts();
                    console.log('🚀 DEBUG: Charts initialized:', Object.keys(charts));
                    console.log('🚀 DEBUG: Temperature chart exists:', !!charts.temperature);
                    console.log('🚀 DEBUG: Rain chart exists:', !!charts.rain);
                    __chartsInitialized = true;
                    trySetDashboardReady();
                return;
            } catch (e) {
                debugLog(logLevels.ERROR, 'Error during initCharts()', e);
                // If we detect the canvas-in-use problem, attempt a single global destroy sweep
                if (String(e).toLowerCase().includes('canvas is already in use') && !__didGlobalChartDestroy) {
                    debugLog(logLevels.WARN, 'Detected canvas-in-use error; performing one-time destroyAllCharts sweep before retry');
                    destroyAllCharts();
                    // Try once more
                    try {
                        initCharts();
                        __chartsInitialized = true;
                        debugLog(logLevels.INFO, 'Charts initialized after global destroy');
                        return;
                    } catch (e2) {
                        debugLog(logLevels.ERROR, 'Second attempt to initCharts failed after destroyAllCharts', e2);
                    }
                }
            }
        }

        // If Chart is missing, dynamically load local vendored Chart.js and adapter
        if (typeof Chart === 'undefined' && !__chartVendorLoading) {
            debugLog(logLevels.WARN, 'Chart.js not found - dynamically loading local vendored copy');

            __chartVendorLoading = true;

            const vendorScript = document.createElement('script');
            vendorScript.src = '/pkg/web/static/chart.umd.js';
            vendorScript.async = false;
            vendorScript.onload = function() {
                debugLog(logLevels.INFO, 'Local Chart.js loaded, attempting to initialize charts');
                // Try to load adapter as well (if present)
                const adapter = document.createElement('script');
                adapter.src = '/pkg/web/static/chartjs-adapter-date-fns.bundle.min.js';
                adapter.async = false;
                adapter.onload = function() {
                    try {
                        initCharts();
                        __chartsInitialized = true;
                        trySetDashboardReady();
                        console.log('🚀 DEBUG: Charts initialized after loading local Chart.js');
                    } catch (e) {
                        debugLog(logLevels.ERROR, 'initCharts failed after loading vendor scripts', e);
                    }
                };
                adapter.onerror = function(err) {
                    debugLog(logLevels.ERROR, 'Failed to load local Chart adapter', err || {});
                    // Still try to initCharts in case adapter isn't required
                    try { initCharts(); } catch (e) { debugLog(logLevels.ERROR, 'initCharts fallback failed', e); }
                };
                document.head.appendChild(adapter);
            };
            vendorScript.onerror = function(err) {
                debugLog(logLevels.ERROR, 'Failed to load local Chart.js vendor script', err || {});
                __chartVendorLoading = false;
            };
            document.head.appendChild(vendorScript);
            return;
        }

        // As a final fallback, try again shortly
        setTimeout(initChartsResilient, 500);
    })();

    // Attach event listeners with debug logging
    debugLog(logLevels.DEBUG, 'Starting to attach event listeners');
    
    // Debug: Check if pressure elements exist and log all IDs
    const pressureCard = document.getElementById('pressure-card');
    const pressureInfoIcon = document.getElementById('pressure-info-icon');
    const allElements = Array.from(document.querySelectorAll('[id]'));
    const allIds = allElements.map(el => el.id);
    
    debugLog(logLevels.DEBUG, 'Pressure elements check', {
        pressureCard: !!pressureCard,
        pressureInfoIcon: !!pressureInfoIcon,
        pressureCardDisplay: pressureCard ? pressureCard.style.display : 'N/A',
        totalElementsWithIds: allElements.length,
        pressureRelatedIds: allIds.filter(id => id.includes('pressure')),
        allIds: allIds
    });
    
    attachEventListener('accessories-row', 'click', toggleAccessoriesExpansion, 'Toggle accessories expansion');
    attachEventListener('lux-info-icon', 'click', toggleLuxTooltip, 'Show/hide lux information tooltip');
    attachEventListener('lux-tooltip-close', 'click', closeLuxTooltip, 'Close lux tooltip');
    attachEventListener('rain-info-icon', 'click', toggleRainTooltip, 'Show/hide rain information tooltip');
    attachEventListener('rain-tooltip-close', 'click', closeRainTooltip, 'Close rain tooltip');
    attachEventListener('humidity-info-icon', 'click', toggleHumidityTooltip, 'Show/hide humidity information tooltip');
    attachEventListener('humidity-tooltip-close', 'click', closeHumidityTooltip, 'Close humidity tooltip');
    attachEventListener('heat-index-info-icon', 'click', toggleHeatIndexTooltip, 'Show/hide heat index tooltip');
    attachEventListener('heat-index-tooltip-close', 'click', closeHeatIndexTooltip, 'Close heat index tooltip');
    attachEventListener('uv-info-icon', 'click', toggleUVTooltip, 'Show/hide UV information tooltip');
    attachEventListener('uv-tooltip-close', 'click', closeUVTooltip, 'Close UV tooltip');
    attachEventListener('pressure-info-icon', 'click', togglePressureTooltip, 'Show/hide pressure tooltip');
    attachEventListener('pressure-tooltip-close', 'click', closePressureTooltip, 'Close pressure tooltip');

    // Global click handlers for closing tooltips
    document.addEventListener('click', handleLuxTooltipClickOutside);
    document.addEventListener('click', handleRainTooltipClickOutside);
    document.addEventListener('click', handleHumidityTooltipClickOutside);
    document.addEventListener('click', handleHeatIndexTooltipClickOutside);
    document.addEventListener('click', handleUVTooltipClickOutside);
    document.addEventListener('click', handlePressureTooltipClickOutside);
    
    debugLog(logLevels.DEBUG, 'All event listeners attached');

    // Start data fetching
    debugLog(logLevels.INFO, 'Starting periodic data fetching (10-second intervals)');
    console.log('🚀 DEBUG: Starting initial data fetch');
    fetchWeather();
    fetchStatus();
    
    setInterval(() => {
        console.log('🚀 DEBUG: Periodic data fetch triggered');
        debugLog(logLevels.DEBUG, 'Periodic data fetch triggered');
        fetchWeather();
        fetchStatus();
    }, 10000);
    
    debugLog(logLevels.INFO, 'Dashboard initialization completed');
});

// Regenerate weather data for testing (for generated weather mode)
async function regenerateWeather() {
    try {
        debugLog(logLevels.INFO, 'Regenerating weather data...');
        
        const response = await fetch('/api/regenerate-weather', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            }
        });
        
        if (!response.ok) {
            throw new Error(`Failed to regenerate weather: ${response.status}`);
        }
        
        const result = await response.json();
        debugLog(logLevels.INFO, 'Weather regenerated successfully', result);
        
        // Show a brief notification
        const stationElement = document.getElementById('tempest-station');
        if (stationElement) {
            const originalContent = stationElement.innerHTML;
            stationElement.innerHTML = '<span style="color: #28a745;">Regenerated! Refreshing...</span>';
            
            // Refresh the data after a short delay
            setTimeout(() => {
                stationElement.innerHTML = originalContent;
                fetchWeather();
                fetchStatus();
            }, 1500);
        }
        
    } catch (error) {
        debugLog(logLevels.ERROR, 'Failed to regenerate weather', error);
        console.error('Failed to regenerate weather:', error);
        
        // Show error notification
        const stationElement = document.getElementById('tempest-station');
        if (stationElement) {
            const originalContent = stationElement.innerHTML;
            stationElement.innerHTML = '<span style="color: #dc3545;">Regeneration failed</span>';
            setTimeout(() => {
                stationElement.innerHTML = originalContent;
            }, 2000);
        }
    }
}

// Debug function to test chart functionality
if (DEBUG_MODE) {
    window.testCharts = function() {
        console.log('🔧 Testing charts functionality');
        console.log('Charts object:', charts);
        console.log('Chart.js available:', typeof Chart !== 'undefined');
        console.log('Temperature chart exists:', !!charts.temperature);
        console.log('Rain chart exists:', !!charts.rain);
        
        if (charts.temperature) {
            try {
                console.log('Temperature chart data length:', charts.temperature.data.datasets[0].data.length);
            } catch (e) { console.log('Temperature chart data not available yet'); }
        }
        
        if (charts.rain) {
            try {
                console.log('Rain chart data length:', charts.rain.data.datasets[0].data.length);
                console.log('Rain chart datasets:', charts.rain.data.datasets.length);
            } catch (e) { console.log('Rain chart data not available yet'); }
        }
        
        console.log('weatherData:', weatherData);
        
        // Try to add a test point
        if (weatherData && charts.temperature) {
            try {
                const testPoint = { x: new Date(), y: 25 };
                charts.temperature.data.datasets[0].data.push(testPoint);
                charts.temperature.update();
                console.log('✅ Successfully added test point to temperature chart');
            } catch (error) {
                console.error('❌ Failed to add test point:', error);
            }
        }
    };

    // Deep chart diagnosis function
    window.diagnoseCharts = function() {
        console.log("🔬 DEEP CHART DIAGNOSIS:");
        
        Object.keys(charts).forEach(chartName => {
            const chart = charts[chartName];
            const canvas = document.getElementById(`${chartName}-chart`);
            
            console.log(`\n📊 ${chartName.toUpperCase()} CHART:`);
            console.log("  Chart Object:", !!chart);
            console.log("  Canvas Element:", !!canvas);
            console.log("  Canvas Visible:", canvas ? (canvas.offsetWidth > 0 && canvas.offsetHeight > 0) : false);
            console.log("  Canvas Dimensions:", canvas ? `${canvas.width}x${canvas.height}` : 'N/A');
            
            if (chart) {
                console.log("  Chart Data:");
                console.log("    Datasets:", chart.data?.datasets?.length || 0);
                console.log("    Labels:", chart.data?.labels?.length || 0);
                
                chart.data?.datasets?.forEach((dataset, idx) => {
                    console.log(`    Dataset ${idx}:`, {
                        label: dataset.label,
                        dataPoints: dataset.data?.length || 0,
                        lastPoint: dataset.data?.[dataset.data.length - 1],
                        borderColor: dataset.borderColor,
                        backgroundColor: dataset.backgroundColor
                    });
                });
                
                // Try to force update
                try {
                    chart.update('none');
                    console.log("  ✅ Chart.update() succeeded");
                } catch (error) {
                    console.log("  ❌ Chart.update() failed:", error);
                }
            }
        });
        
        // Also test Chart.js availability
        console.log("\n🔧 CHART.JS STATUS:");
        console.log("  Chart.js loaded:", typeof Chart !== 'undefined');
        console.log("  Chart version:", Chart?.version || 'Unknown');
        console.log("  Chart instances:", Chart?.instances?.length || 0);
    };
}

// Canvas inspection function
window.inspectCanvas = function() {
    console.log("🎨 CANVAS INSPECTION:");
    
    const chartNames = ['temperature', 'humidity', 'wind', 'rain', 'pressure', 'light', 'uv'];
    
    chartNames.forEach(chartName => {
        const canvas = document.getElementById(`${chartName}-chart`);
        const container = canvas?.parentElement;
        
        console.log(`\n🖼️ ${chartName.toUpperCase()} CANVAS:`);
        console.log("  Canvas exists:", !!canvas);
        
        if (canvas) {
            const rect = canvas.getBoundingClientRect();
            console.log("  Canvas position:", {
                width: canvas.width,
                height: canvas.height,
                clientWidth: canvas.clientWidth,
                clientHeight: canvas.clientHeight,
                offsetWidth: canvas.offsetWidth,
                offsetHeight: canvas.offsetHeight,
                boundingRect: {
                    width: rect.width,
                    height: rect.height,
                    top: rect.top,
                    left: rect.left
                }
            });
            
            console.log("  Canvas style:", {
                display: canvas.style.display || getComputedStyle(canvas).display,
                visibility: canvas.style.visibility || getComputedStyle(canvas).visibility,
                position: getComputedStyle(canvas).position
            });
            
            console.log("  Container:", {
                exists: !!container,
                className: container?.className,
                offsetWidth: container?.offsetWidth,
                offsetHeight: container?.offsetHeight
            });
        }
    });
};

// Force chart render function
window.forceRenderCharts = function() {
    console.log("🔄 FORCING CHART RENDERS:");
    
    Object.keys(charts).forEach(chartName => {
        const chart = charts[chartName];
        if (chart) {
            try {
                console.log(`📊 Forcing render for ${chartName} chart...`);
                
                // Force multiple update strategies
                chart.update('none'); // No animation
                chart.render(); // Force immediate render
                chart.draw(); // Force draw
                
                console.log(`✅ ${chartName} chart render completed`);
                console.log(`   Data points: ${chart.data.datasets[0]?.data?.length || 0}`);
                console.log(`   Canvas attached: ${!!chart.canvas}`);
                console.log(`   Chart rendered: ${chart.rendered || 'unknown'}`);
            } catch (error) {
                console.error(`❌ Error rendering ${chartName} chart:`, error);
            }
        }
    });
};

// Recreate charts function 
window.recreateCharts = function() {
    console.log("🔨 RECREATING ALL CHARTS:");
    
    // Destroy existing charts
    Object.keys(charts).forEach(chartName => {
        if (charts[chartName]) {
            try {
                charts[chartName].destroy();
                console.log(`🗑️ Destroyed ${chartName} chart`);
            } catch (error) {
                console.error(`Error destroying ${chartName} chart:`, error);
            }
        }
    });
    
    // Clear charts object
    Object.keys(charts).forEach(key => delete charts[key]);
    
    // Wait a moment then reinitialize
    setTimeout(() => {
        console.log("🔄 Reinitializing charts...");
        initCharts();
        
        // Try to repopulate with current data if available
        if (typeof weatherData !== 'undefined' && weatherData) {
            setTimeout(() => {
                updateCharts(weatherData);
                console.log("📊 Charts recreated and populated with current data");
            }, 100);
        }
    }, 100);
};

// Debug chart scales and force visible data
window.debugChartScales = function() {
    console.log("🔍 DEBUGGING CHART SCALES:");
    
    Object.keys(charts).forEach(chartName => {
        const chart = charts[chartName];
        if (chart) {
            console.log(`\n📊 ${chartName.toUpperCase()} CHART SCALE DEBUG:`);
            
            // Check data ranges
            chart.data.datasets.forEach((dataset, idx) => {
                console.log(`  Dataset ${idx} (${dataset.label}):`);
                console.log(`    Data points: ${dataset.data.length}`);
                if (dataset.data.length > 0) {
                    const values = dataset.data.map(d => typeof d === 'object' ? d.y : d);
                    console.log(`    Values: [${values.join(', ')}]`);
                    console.log(`    Min: ${Math.min(...values)}, Max: ${Math.max(...values)}`);
                }
            });
            
            // Check scales
            if (chart.scales) {
                Object.values(chart.scales).forEach(scale => {
                    console.log(`  Scale ${scale.id}:`);
                    console.log(`    Type: ${scale.type}`);
                    console.log(`    Min: ${scale.min}, Max: ${scale.max}`);
                    console.log(`    Range: ${scale.max - scale.min}`);
                });
            }
            
            // Force a visible range and update
            try {
                // Add some test data points with visible values
                const testData = [
                    { x: new Date(Date.now() - 60000), y: 20 },
                    { x: new Date(Date.now() - 30000), y: 25 },
                    { x: new Date(), y: 30 }
                ];
                
                chart.data.datasets[0].data = [...chart.data.datasets[0].data, ...testData];
                chart.update('none');
                
                console.log(`  ✅ Added test data to ${chartName} chart`);
            } catch (error) {
                console.error(`  ❌ Error adding test data to ${chartName}:`, error);
            }
        }
    });
};

// Fix timestamp issues in all charts
window.fixChartTimestamps = function() {
    console.log("🔧 FIXING CHART TIMESTAMPS:");
    
    const now = new Date();
    const oneMinuteAgo = new Date(now.getTime() - 60000);
    const twoMinutesAgo = new Date(now.getTime() - 120000);
    
    Object.keys(charts).forEach(chartName => {
        const chart = charts[chartName];
        if (chart) {
            console.log(`📊 Fixing ${chartName} chart timestamps...`);
            
            // Clear existing data and add properly timestamped data
            chart.data.datasets.forEach((dataset, idx) => {
                // Get current values or use defaults
                let value1 = 20, value2 = 25, value3 = 30;
                
                if (dataset.data.length > 0) {
                    const lastPoint = dataset.data[dataset.data.length - 1];
                    const baseValue = typeof lastPoint === 'object' ? lastPoint.y : lastPoint;
                    value1 = baseValue * 0.9;
                    value2 = baseValue * 0.95; 
                    value3 = baseValue;
                }
                
                // Set properly timestamped data
                dataset.data = [
                    { x: twoMinutesAgo, y: value1 },
                    { x: oneMinuteAgo, y: value2 },
                    { x: now, y: value3 }
                ];
                
                console.log(`  Dataset ${idx}: ${dataset.data.length} points with proper timestamps`);
            });
            
            // Force chart update
            chart.update('none');
            console.log(`  ✅ ${chartName} chart timestamps fixed`);
        }
    });
    
    console.log("🎯 All chart timestamps fixed! Charts should now be visible.");
};