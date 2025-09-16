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

let weatherData = null;
let forecastData = null; // Store current forecast data for unit conversions
let statusData = null; // Store current status data for unit conversions
const charts = {};
const maxDataPoints = 1000; // As specified in requirements

function initCharts() {
    debugLog(logLevels.DEBUG, 'Initializing all charts with configuration');
    
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
                    callbacks: {
                        title: function(context) {
                            // Force 24-hour format for tooltip titles
                            const date = new Date(context[0].parsed.x);
                            return date.toLocaleDateString('en-US', { 
                                month: 'short', 
                                day: '2-digit' 
                            }) + ', ' + date.toLocaleTimeString('en-GB', { 
                                hour: '2-digit', 
                                minute: '2-digit',
                                hour12: false 
                            });
                        }
                    }
                }
            },
            scales: {
                x: {
                    display: true,
                    type: 'time',
                    time: {
                        displayFormats: {
                            minute: 'HH:mm',
                            hour: 'HH:mm',
                            day: 'MMM dd'
                        },
                        tooltipFormat: 'MMM dd, HH:mm'
                    },
                    grid: {
                        display: true,
                        color: 'rgba(0,0,0,0.1)'
                    },
                    ticks: {
                        maxTicksLimit: 6,
                        color: '#666',
                        font: {
                            size: 10
                        },
                        callback: function(value, index, values) {
                            // Force 24-hour format for tick labels
                            return new Date(value).toLocaleTimeString('en-GB', { 
                                hour: '2-digit', 
                                minute: '2-digit',
                                hour12: false 
                            });
                        }
                    },
                    title: {
                        display: true,
                        text: 'Time',
                        color: '#666',
                        font: {
                            size: 12
                        }
                    }
                },
                y: {
                    display: true,
                    grid: {
                        display: true,
                        color: 'rgba(0,0,0,0.1)'
                    },
                    ticks: {
                        maxTicksLimit: 5,
                        color: '#666',
                        font: {
                            size: 10
                        },
                        callback: function(value) {
                            return value.toFixed(1);
                        }
                    },
                    title: {
                        display: true,
                        text: 'Value',
                        color: '#666',
                        font: {
                            size: 12
                        }
                    }
                }
            },
            elements: {
                point: { radius: 0 },
                line: { borderWidth: 2 }
            },
            interaction: {
                intersect: false,
                mode: 'index'
            }
        }
    };

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
    
    // Force the colors after creation
    charts.temperature.data.datasets[1].borderColor = '#00cc66';
    charts.temperature.data.datasets[1].backgroundColor = 'rgba(0, 204, 102, 0.2)';
    
    debugLog(logLevels.INFO, 'Temperature chart created with colors:', {
        dataColor: charts.temperature.data.datasets[0].borderColor,
        avgColor: charts.temperature.data.datasets[1].borderColor
    });

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
    
    // Force the colors after creation
    charts.humidity.data.datasets[1].borderColor = '#ff8533';
    charts.humidity.data.datasets[1].backgroundColor = 'rgba(255, 133, 51, 0.2)';
    
    debugLog(logLevels.INFO, 'Humidity chart created with colors:', {
        dataColor: charts.humidity.data.datasets[0].borderColor,
        avgColor: charts.humidity.data.datasets[1].borderColor
    });

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
                label: 'Rain'
            }, {
                data: [],
                borderColor: '#66ff66',
                backgroundColor: 'rgba(102, 255, 102, 0.2)',
                borderDash: [5, 5],
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                tension: 0,
                label: 'Average'
            }]
        }
    });

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
            }, {
                data: [],
                borderColor: '#8a56ff',
                backgroundColor: 'rgba(138, 86, 255, 0.2)',
                borderDash: [5, 5],
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                tension: 0,
                label: 'Average'
            }]
        }
    });

    // Check if UV chart element exists before creating chart
    if (ctxUV) {
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
                }, {
                    data: [],
                    borderColor: '#66ff66',
                    backgroundColor: 'rgba(102, 255, 102, 0.2)',
                    borderDash: [5, 5],
                    borderWidth: 2,
                    fill: false,
                    pointRadius: 0,
                    tension: 0,
                    label: 'Average'
                }]
            }
        });
        debugLog(logLevels.DEBUG, 'UV chart created successfully');
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
        charts.temperature.data.datasets[1].borderColor = '#00cc66';
        charts.temperature.data.datasets[1].backgroundColor = 'rgba(0, 204, 102, 0.2)';
        charts.temperature.update('none');
    }
    
    // Humidity: Blue data → Orange average  
    if (charts.humidity) {
        charts.humidity.data.datasets[1].borderColor = '#ff8533';
        charts.humidity.data.datasets[1].backgroundColor = 'rgba(255, 133, 51, 0.2)';
        charts.humidity.update('none');
    }
    
    // Wind: Teal data → Bright Red average (more visible)
    if (charts.wind) {
        charts.wind.data.datasets[1].borderColor = '#FF0000';
        charts.wind.data.datasets[1].backgroundColor = 'rgba(255, 0, 0, 0.3)';
        charts.wind.data.datasets[1].borderWidth = 3;
        charts.wind.update('none');
        
        debugLog(logLevels.INFO, '🌬️ Wind chart colors applied:', {
            dataColor: charts.wind.data.datasets[0].borderColor,
            avgColor: charts.wind.data.datasets[1].borderColor,
            avgDataPoints: charts.wind.data.datasets[1].data.length
        });
    }
    
    // Rain: Purple data → Yellow-green average
    if (charts.rain) {
        charts.rain.data.datasets[1].borderColor = '#66ff66';
        charts.rain.data.datasets[1].backgroundColor = 'rgba(102, 255, 102, 0.2)';
        charts.rain.update('none');
    }
    
    // Pressure: Orange data → Blue average
    if (charts.pressure) {
        charts.pressure.data.datasets[1].borderColor = '#4080ff';
        charts.pressure.data.datasets[1].backgroundColor = 'rgba(64, 128, 255, 0.2)';
        charts.pressure.update('none');
    }
    
    // Light: Yellow data → Purple average
    if (charts.light) {
        charts.light.data.datasets[1].borderColor = '#8a56ff';
        charts.light.data.datasets[1].backgroundColor = 'rgba(138, 86, 255, 0.2)';
        charts.light.update('none');
    }
    
    // UV: Purple data → Yellow-green average
    if (charts.uv) {
        charts.uv.data.datasets[1].borderColor = '#66ff66';
        charts.uv.data.datasets[1].backgroundColor = 'rgba(102, 255, 102, 0.2)';
        charts.uv.update('none');
    }
    
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
    document.getElementById('temperature-unit').textContent = units.temperature === 'celsius' ? '°C' : '°F';
    document.getElementById('wind-unit').textContent = units.wind === 'mph' ? 'mph' : 'kph';
    
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
    if (data.length === 0) {
        chart.data.datasets[1].data = [];
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

    chart.data.datasets[1].data = movingAverageData;
    
    debugLog(logLevels.DEBUG, 'Moving average updated', {
        chartLabel: chart.data.datasets[0].label || 'Unknown',
        dataPoints: data.length,
        windowSize: windowSize,
        averagePoints: movingAverageData.length
    });
}

function validateAndSortChartData(chart) {
    // Validate and sort data for the main dataset
    if (chart.data.datasets[0] && chart.data.datasets[0].data) {
        let data = chart.data.datasets[0].data;
        
        // Filter out invalid data points
        data = data.filter(point => 
            point && 
            typeof point.x !== 'undefined' && 
            typeof point.y === 'number' && 
            !isNaN(point.y) &&
            point.x instanceof Date || typeof point.x === 'number'
        );
        
        // Sort by timestamp
        data.sort((a, b) => new Date(a.x) - new Date(b.x));
        
        // Remove duplicate timestamps, keeping the most recent value
        const uniqueData = [];
        for (let i = 0; i < data.length; i++) {
            const current = data[i];
            const next = data[i + 1];
            
            if (!next || new Date(current.x).getTime() !== new Date(next.x).getTime()) {
                uniqueData.push(current);
            }
        }
        
        chart.data.datasets[0].data = uniqueData;
        
        debugLog(logLevels.DEBUG, 'Chart data validated and sorted', {
            originalPoints: data.length,
            filteredPoints: uniqueData.length,
            removed: data.length - uniqueData.length
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
    let temp = weatherData.temperature;
    if (units.temperature === 'fahrenheit') {
        temp = celsiusToFahrenheit(temp);
    }
    document.getElementById('temperature').textContent = temp.toFixed(1);
    debugLog(logLevels.DEBUG, 'Temperature updated', {
        original: weatherData.temperature,
        converted: temp,
        unit: units.temperature
    });

    // Humidity and heat index
    document.getElementById('humidity').textContent = weatherData.humidity.toFixed(1);
    document.getElementById('humidity-description').textContent = getHumidityDescription(weatherData.humidity);
    
    // Calculate and display heat index
    const heatIndexC = calculateHeatIndex(weatherData.temperature, weatherData.humidity);
    let heatIndexDisplay = heatIndexC;
    if (units.temperature === 'fahrenheit') {
        heatIndexDisplay = celsiusToFahrenheit(heatIndexC);
    }
    const tempUnit = units.temperature === 'celsius' ? '°C' : '°F';
    const heatIndexElement = document.getElementById('heat-index');
    if (heatIndexElement) {
        heatIndexElement.textContent = heatIndexDisplay.toFixed(1) + tempUnit;
    }
    debugLog(logLevels.DEBUG, 'Heat index calculated and displayed', {
        heatIndexC: heatIndexC,
        heatIndexDisplay: heatIndexDisplay,
        unit: tempUnit
    });

    // Wind data
    let windSpeed = weatherData.windSpeed;
    let windGust = weatherData.windGust;
    if (units.wind === 'kph') {
        windSpeed = mphToKph(windSpeed);
        windGust = mphToKph(windGust);
    }
    document.getElementById('wind-speed').textContent = windSpeed.toFixed(1);
    
    // Wind gust information
    const windUnit = units.wind === 'kph' ? 'kph' : 'mph';
    if (windGust > windSpeed) {
        document.getElementById('wind-gust-info').textContent = `Winds gusting to ${windGust.toFixed(1)} ${windUnit}`;
    } else if (windGust > 0) {
        document.getElementById('wind-gust-info').textContent = `Gusts up to ${windGust.toFixed(1)} ${windUnit}`;
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
    let rain = weatherData.rainAccum;
    let dailyRain = weatherData.rainDailyTotal || 0;
    if (units.rain === 'mm') {
        rain = inchesToMm(rain);
        dailyRain = inchesToMm(dailyRain);
    }
    document.getElementById('rain').textContent = rain.toFixed(3);
    
    // Display daily rain total
    const dailyRainElement = document.getElementById('daily-rain-total');
    if (dailyRainElement) {
        const rainUnit = units.rain === 'inches' ? 'in' : 'mm';
        dailyRainElement.textContent = dailyRain.toFixed(3) + ' ' + rainUnit;
    }
    
    // Display rain description based on current accumulated rain
    const rainDescElement = document.getElementById('rain-description');
    if (rainDescElement) {
        // Convert to mm for description calculation (standard meteorological unit)
        let rainMm = units.rain === 'inches' ? rain * 25.4 : rain;
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
        convertedRain: rain,
        originalDailyRain: weatherData.rainDailyTotal,
        convertedDailyRain: dailyRain,
        rainUnit: units.rain,
        rainDescription: getRainDescription(units.rain === 'inches' ? rain * 25.4 : rain),
        lightningCount: weatherData.lightningStrikeCount,
        lightningDistance: weatherData.lightningStrikeAvg
    });

    let pressure = weatherData.pressure;
    if (units.pressure === 'inHg') {
        pressure = mbToInHg(pressure);
    }
    document.getElementById('pressure').textContent = pressure.toFixed(1);
    
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
    if (!weatherData) {
        debugLog(logLevels.WARN, 'updateCharts called but no weatherData available');
        return;
    }
    
    debugLog(logLevels.DEBUG, 'Starting chart updates');
    const now = new Date(weatherData.lastUpdate);

    // Temperature chart
    let tempValue = weatherData.temperature;
    if (units.temperature === 'fahrenheit') {
        tempValue = celsiusToFahrenheit(tempValue);
    }
    charts.temperature.data.datasets[0].data.push({ x: now, y: tempValue });
    if (charts.temperature.data.datasets[0].data.length > maxDataPoints) {
        charts.temperature.data.datasets[0].data.shift();
    }
    const tempAvg = calculateAverage(charts.temperature.data.datasets[0].data);
    validateAndSortChartData(charts.temperature);
    updateAverageLine(charts.temperature, charts.temperature.data.datasets[0].data);
    charts.temperature.options.scales.y.title = {
        display: true,
        text: units.temperature === 'celsius' ? '°C' : '°F'
    };
    charts.temperature.update();
    
    debugLog(logLevels.DEBUG, 'Temperature chart updated', {
        dataPoints: charts.temperature.data.datasets[0].data.length,
        currentValue: tempValue,
        average: tempAvg
    });

    // Humidity chart
    charts.humidity.data.datasets[0].data.push({ x: now, y: weatherData.humidity });
    if (charts.humidity.data.datasets[0].data.length > maxDataPoints) {
        charts.humidity.data.datasets[0].data.shift();
    }
    const humidityAvg = calculateAverage(charts.humidity.data.datasets[0].data);
    updateAverageLine(charts.humidity, charts.humidity.data.datasets[0].data);
    charts.humidity.options.scales.y.title = {
        display: true,
        text: '%'
    };
    charts.humidity.update();

    // Wind chart
    let windValue = weatherData.windSpeed;
    if (units.wind === 'kph') {
        windValue = mphToKph(windValue);
    }
    charts.wind.data.datasets[0].data.push({ x: now, y: windValue });
    if (charts.wind.data.datasets[0].data.length > maxDataPoints) {
        charts.wind.data.datasets[0].data.shift();
    }
    const windAvg = calculateAverage(charts.wind.data.datasets[0].data);
    updateAverageLine(charts.wind, charts.wind.data.datasets[0].data);
    charts.wind.options.scales.y.title = {
        display: true,
        text: units.wind === 'mph' ? 'mph' : 'kph'
    };
    charts.wind.update();

    // Rain chart
    let rainValue = weatherData.rainAccum;
    if (units.rain === 'mm') {
        rainValue = inchesToMm(rainValue);
    }
    charts.rain.data.datasets[0].data.push({ x: now, y: rainValue });
    if (charts.rain.data.datasets[0].data.length > maxDataPoints) {
        charts.rain.data.datasets[0].data.shift();
    }
    const rainAvg = calculateAverage(charts.rain.data.datasets[0].data);
    updateAverageLine(charts.rain, charts.rain.data.datasets[0].data);
    charts.rain.options.scales.y.title = {
        display: true,
        text: units.rain === 'inches' ? 'in' : 'mm'
    };
    charts.rain.update();

    // Pressure chart
    let pressureValue = weatherData.pressure;
    if (units.pressure === 'inHg') {
        pressureValue = mbToInHg(pressureValue);
    }
    charts.pressure.data.datasets[0].data.push({ x: now, y: pressureValue });
    if (charts.pressure.data.datasets[0].data.length > maxDataPoints) {
        charts.pressure.data.datasets[0].data.shift();
    }
    const pressureAvg = calculateAverage(charts.pressure.data.datasets[0].data);
    updateAverageLine(charts.pressure, charts.pressure.data.datasets[0].data);
    updateTrendLine(charts.pressure, charts.pressure.data.datasets[0].data);
    charts.pressure.options.scales.y.title = {
        display: true,
        text: units.pressure === 'mb' ? 'mb' : 'inHg'
    };
    charts.pressure.update();

    // Light chart
    charts.light.data.datasets[0].data.push({ x: now, y: weatherData.illuminance });
    if (charts.light.data.datasets[0].data.length > maxDataPoints) {
        charts.light.data.datasets[0].data.shift();
    }
    const lightAvg = calculateAverage(charts.light.data.datasets[0].data);
    updateAverageLine(charts.light, charts.light.data.datasets[0].data);
    charts.light.options.scales.y.title = {
        display: true,
        text: 'lux'
    };
    charts.light.update();
    
    debugLog(logLevels.DEBUG, 'Light chart updated', {
        dataPoints: charts.light.data.datasets[0].data.length,
        currentValue: weatherData.illuminance,
        average: lightAvg
    });

    // UV chart - only update if it exists
    if (charts.uv && charts.uv.data) {
        charts.uv.data.datasets[0].data.push({ x: now, y: weatherData.uv });
        if (charts.uv.data.datasets[0].data.length > maxDataPoints) {
            charts.uv.data.datasets[0].data.shift();
        }
        const uvAvg = calculateAverage(charts.uv.data.datasets[0].data);
        updateAverageLine(charts.uv, charts.uv.data.datasets[0].data);
        charts.uv.options.scales.y.title = {
            display: true,
            text: 'UVI'
        };
        charts.uv.update();
        
        debugLog(logLevels.DEBUG, 'UV chart updated', {
            dataPoints: charts.uv.data.datasets[0].data.length,
            currentValue: weatherData.uv,
            average: uvAvg
        });
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
            
            updateDisplay();
            updateCharts();
            document.getElementById('status').textContent = 'Connected to Tempest station';
            document.getElementById('status').style.background = 'rgba(40, 167, 69, 0.1)';
            
            debugLog(logLevels.INFO, 'Weather fetch completed successfully', {
                totalTime: (performance.now() - startTime).toFixed(2) + 'ms'
            });
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
    debugLog(logLevels.DEBUG, 'Starting status API call');
    
    try {
        const response = await fetch('/api/status');
        const responseTime = (performance.now() - startTime).toFixed(2);
        
        if (response.ok) {
            const status = await response.json();
            debugLog(logLevels.DEBUG, 'Status API response received', {
                responseTime: responseTime + 'ms',
                connected: status.connected,
                stationName: status.stationName,
                uptime: status.uptime,
                homekitStatus: status.homekit
            });
            
            updateStatusDisplay(status);
            updateForecastDisplay(status);
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
            tempestStatus.textContent = status.connected ? 'Connected' : 'Disconnected';
            tempestStatus.style.color = status.connected ? '#28a745' : '#dc3545';
        }
    }
    if (tempestStation) tempestStation.textContent = status.stationName || '--';
    
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
                <span class="forecast-day-high">${Math.round(highTemp)}${tempUnit}</span>
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

    // Clear existing chart data
    charts.temperature.data.datasets[0].data = [];
    charts.humidity.data.datasets[0].data = [];
    charts.wind.data.datasets[0].data = [];
    charts.rain.data.datasets[0].data = [];
    charts.pressure.data.datasets[0].data = [];
    charts.light.data.datasets[0].data = [];
    if (charts.uv && charts.uv.data) {
        charts.uv.data.datasets[0].data = [];
    }

    // Process each historical data point
    for (let i = 0; i < dataHistory.length; i++) {
        const obs = dataHistory[i];
        const timestamp = new Date(obs.lastUpdate);

        // Temperature
        let tempValue = obs.temperature;
        if (units.temperature === 'fahrenheit') {
            tempValue = celsiusToFahrenheit(tempValue);
        }
        charts.temperature.data.datasets[0].data.push({ x: timestamp, y: tempValue });

        // Humidity
        charts.humidity.data.datasets[0].data.push({ x: timestamp, y: obs.humidity });

        // Wind
        let windValue = obs.windSpeed;
        if (units.wind === 'kph') {
            windValue = mphToKph(windValue);
        }
        charts.wind.data.datasets[0].data.push({ x: timestamp, y: windValue });

        // Rain
        let rainValue = obs.rainAccum;
        if (units.rain === 'mm') {
            rainValue = inchesToMm(rainValue);
        }
        charts.rain.data.datasets[0].data.push({ x: timestamp, y: rainValue });

        // Pressure
        let pressureValue = obs.pressure;
        if (units.pressure === 'inHg') {
            pressureValue = mbToInHg(pressureValue);
        }
        charts.pressure.data.datasets[0].data.push({ x: timestamp, y: pressureValue });

        // Light
        charts.light.data.datasets[0].data.push({ x: timestamp, y: obs.illuminance });

        // UV (if chart exists)
        if (charts.uv && charts.uv.data) {
            charts.uv.data.datasets[0].data.push({ x: timestamp, y: obs.uv });
        }
    }

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
    }

    // Update pressure average and trend
    if (charts.pressure.data.datasets[0].data.length > 0) {
        validateAndSortChartData(charts.pressure);
        const pressureAvg = calculateAverage(charts.pressure.data.datasets[0].data);
        updateAverageLine(charts.pressure, charts.pressure.data.datasets[0].data);
        updateTrendLine(charts.pressure, charts.pressure.data.datasets[0].data);
    }

    // Update light average
    if (charts.light.data.datasets[0].data.length > 0) {
        validateAndSortChartData(charts.light);
        const lightAvg = calculateAverage(charts.light.data.datasets[0].data);
        updateAverageLine(charts.light, charts.light.data.datasets[0].data);
    }

    // Update UV average
    if (charts.uv && charts.uv.data && charts.uv.data.datasets[0].data.length > 0) {
        validateAndSortChartData(charts.uv);
        const uvAvg = calculateAverage(charts.uv.data.datasets[0].data);
        updateAverageLine(charts.uv, charts.uv.data.datasets[0].data);
    }
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
    
    updateUnits();
    
    // Ensure pressure info icon has proper event listener attached initially
    const initialPressureInfoIcon = document.getElementById('pressure-info-icon');
    if (initialPressureInfoIcon) {
        initialPressureInfoIcon.addEventListener('click', function(e) {
            e.stopPropagation();
            togglePressureTooltip(e);
        });
        console.log('🔧 Initial setup - Attached click event listener to pressure info icon');
    }
    
    initCharts();

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
    fetchWeather();
    fetchStatus();
    
    setInterval(() => {
        debugLog(logLevels.DEBUG, 'Periodic data fetch triggered');
        fetchWeather();
        fetchStatus();
    }, 10000);
    
    debugLog(logLevels.INFO, 'Dashboard initialization completed');
});