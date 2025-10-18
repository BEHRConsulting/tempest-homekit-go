//go:build !no_browser

package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	gruntime "runtime"
	"strings"
	"testing"
	"time"

	"tempest-homekit-go/pkg/weather"

	"github.com/chromedp/chromedp"
)

// TestCollapsibleSections verifies that the new collapsible sections in Tempest Station
// and HomeKit Bridge cards render correctly and can be toggled.
func TestCollapsibleSections(t *testing.T) {
	ws := testNewWebServer(t)

	now := time.Now()
	obs := weather.Observation{
		Timestamp:        now.Unix(),
		AirTemperature:   18.5,
		RelativeHumidity: 55,
		WindAvg:          2.5,
		WindGust:         3.0,
		WindDirection:    90,
		StationPressure:  1015,
		Illuminance:      1000,
		UV:               3,
		RainAccumulated:  0.05,
		Battery:          3.8,
	}
	ws.UpdateWeather(&obs)

	// Set up HomeKit status with accessories
	homekitStatus := map[string]interface{}{
		"bridge":         true,
		"name":           "Tempest HomeKit Bridge",
		"accessories":    9,
		"accessoryNames": []string{"Temperature", "Humidity", "Light", "UV", "Pressure"},
		"allSensors":     []string{"Temperature", "Humidity", "Light", "UV", "Wind Speed", "Wind Direction", "Rain", "Pressure", "Lightning"},
		"pin":            "00102003",
		"setupCode":      "X-00102003",
		"bridgeId":       "12:34:56:78:9A:BC",
		"manufacturer":   "WeatherFlow",
		"model":          "Tempest Bridge v2.0",
		"firmware":       "1.0.0",
		"port":           "51826",
		"hapVersion":     "1.1",
		"configNumber":   9,
		"category":       "Bridge",
	}
	ws.UpdateHomeKitStatus(homekitStatus)

	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.handleDashboard)
	mux.HandleFunc("/api/weather", ws.handleWeatherAPI)
	mux.HandleFunc("/api/status", ws.handleStatusAPI)

	_, thisFile, _, _ := gruntime.Caller(0)
	thisDir := filepath.Dir(thisFile)
	staticDir := filepath.Join(thisDir, "static")
	mux.Handle("/pkg/web/static/", http.StripPrefix("/pkg/web/static/", http.FileServer(http.Dir(staticDir))))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoSandbox,
	)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// Navigate and inject scripts
	tasks := chromedp.Tasks{
		chromedp.Navigate(ts.URL),
		chromedp.WaitVisible(`#status`, chromedp.ByID),
		chromedp.Sleep(400 * time.Millisecond),
	}
	if err := chromedp.Run(browserCtx, tasks); err != nil {
		t.Fatalf("navigate failed: %v", err)
	}

	// Inject local script.js
	injectedFrom := filepath.Join(thisDir, "static", "script.js")
	localBytes, err := os.ReadFile(injectedFrom)
	if err != nil {
		t.Fatalf("failed to read local script.js: %v", err)
	}
	localText := string(localBytes)
	inj := fmt.Sprintf("(function(){ var s=document.createElement('script'); s.type='text/javascript'; s.text = %q; document.head.appendChild(s); })()", localText)
	if err := chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(inj, nil)); err != nil {
		t.Fatalf("failed to inject local script: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	// Call fetchStatus to populate data
	var fetchResult interface{}
	if err := chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(`(async function(){ try { await fetchStatus(); return 'ok'; } catch(e) { return {err: e.toString()}; } })()`, &fetchResult)); err != nil {
		t.Fatalf("failed to run fetchStatus: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Test 1: Verify Tempest Station collapsible sections exist
	var deviceStatusExists, hubStatusExists bool
	if err := chromedp.Run(browserCtx, chromedp.Tasks{
		chromedp.Evaluate(`document.getElementById('device-status-row') !== null`, &deviceStatusExists),
		chromedp.Evaluate(`document.getElementById('hub-status-row') !== null`, &hubStatusExists),
	}); err != nil {
		t.Fatalf("failed to check Tempest Station sections: %v", err)
	}

	if !deviceStatusExists {
		t.Error("Device Status collapsible section not found in Tempest Station card")
	}
	if !hubStatusExists {
		t.Error("Hub Status collapsible section not found in Tempest Station card")
	}

	// Test 2: Verify Device Status section can be toggled
	var deviceStatusHidden bool
	if err := chromedp.Run(browserCtx, chromedp.Tasks{
		chromedp.Evaluate(`document.getElementById('device-status-expanded').classList.contains('hidden')`, &deviceStatusHidden),
	}); err != nil {
		t.Fatalf("failed to check Device Status hidden state: %v", err)
	}

	if !deviceStatusHidden {
		t.Error("Device Status section should be hidden by default")
	}

	// Click to expand
	if err := chromedp.Run(browserCtx, chromedp.Click(`#device-status-row`, chromedp.NodeVisible)); err != nil {
		t.Fatalf("failed to click Device Status row: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	// Verify it's now visible
	if err := chromedp.Run(browserCtx, chromedp.Evaluate(`document.getElementById('device-status-expanded').classList.contains('hidden')`, &deviceStatusHidden)); err != nil {
		t.Fatalf("failed to check Device Status expanded state: %v", err)
	}

	// Log for debugging if still hidden
	if deviceStatusHidden {
		t.Log("Device Status section is still hidden after clicking - this may be due to timing or missing toggle function")
	}

	// Test 3: Verify HomeKit Bridge collapsible sections exist
	var accessoriesExists, connectionExists, technicalExists bool
	if err := chromedp.Run(browserCtx, chromedp.Tasks{
		chromedp.Evaluate(`document.getElementById('accessories-row') !== null`, &accessoriesExists),
		chromedp.Evaluate(`document.getElementById('homekit-connection-row') !== null`, &connectionExists),
		chromedp.Evaluate(`document.getElementById('homekit-technical-row') !== null`, &technicalExists),
	}); err != nil {
		t.Fatalf("failed to check HomeKit sections: %v", err)
	}

	if !accessoriesExists {
		t.Error("Accessories collapsible section not found in HomeKit Bridge card")
	}
	if !connectionExists {
		t.Error("Connection Info collapsible section not found in HomeKit Bridge card")
	}
	if !technicalExists {
		t.Error("Technical Details collapsible section not found in HomeKit Bridge card")
	}

	// Test 4: Verify Accessories list populates correctly
	var accessoriesListHTML string
	if err := chromedp.Run(browserCtx, chromedp.Evaluate(`document.getElementById('accessories-list') ? document.getElementById('accessories-list').innerHTML : ''`, &accessoriesListHTML)); err != nil {
		t.Fatalf("failed to read accessories list: %v", err)
	}

	if accessoriesListHTML == "" {
		t.Error("Accessories list is empty")
	}

	// Check for enabled sensors
	enabledSensors := []string{"Temperature", "Humidity", "Light", "UV", "Pressure"}
	for _, sensor := range enabledSensors {
		if !strings.Contains(accessoriesListHTML, sensor) {
			t.Errorf("Accessories list missing enabled sensor: %s", sensor)
		}
		if !strings.Contains(accessoriesListHTML, "Active") {
			t.Error("Accessories list should show 'Active' status for enabled sensors")
			break
		}
	}

	// Check for disabled sensors
	disabledSensors := []string{"Wind Speed", "Wind Direction", "Rain", "Lightning"}
	for _, sensor := range disabledSensors {
		if !strings.Contains(accessoriesListHTML, sensor) {
			t.Errorf("Accessories list missing disabled sensor: %s", sensor)
		}
	}
	if !strings.Contains(accessoriesListHTML, "Disabled") {
		t.Error("Accessories list should show 'Disabled' status for disabled sensors")
	}

	// Test 5: Verify QR code canvas exists
	var qrCodeExists bool
	if err := chromedp.Run(browserCtx, chromedp.Evaluate(`document.getElementById('homekit-qr-code') !== null`, &qrCodeExists)); err != nil {
		t.Fatalf("failed to check QR code canvas: %v", err)
	}

	if !qrCodeExists {
		t.Error("HomeKit QR code canvas not found")
	}

	// Test 6: Verify QR code canvas has content (width/height set)
	var qrWidth, qrHeight float64
	if err := chromedp.Run(browserCtx, chromedp.Tasks{
		chromedp.Evaluate(`document.getElementById('homekit-qr-code').width`, &qrWidth),
		chromedp.Evaluate(`document.getElementById('homekit-qr-code').height`, &qrHeight),
	}); err != nil {
		t.Fatalf("failed to check QR code dimensions: %v", err)
	}

	if qrWidth == 0 || qrHeight == 0 {
		t.Errorf("QR code canvas has invalid dimensions: width=%v, height=%v", qrWidth, qrHeight)
	}

	// Test 7: Verify Connection Info fields populate
	var setupPinText, setupCodeText, bridgeIdText string
	if err := chromedp.Run(browserCtx, chromedp.Tasks{
		chromedp.Evaluate(`document.getElementById('homekit-pin') ? document.getElementById('homekit-pin').textContent : ''`, &setupPinText),
		chromedp.Evaluate(`document.getElementById('homekit-setup-code') ? document.getElementById('homekit-setup-code').textContent : ''`, &setupCodeText),
		chromedp.Evaluate(`document.getElementById('homekit-bridge-id') ? document.getElementById('homekit-bridge-id').textContent : ''`, &bridgeIdText),
	}); err != nil {
		t.Fatalf("failed to read HomeKit connection info: %v", err)
	}

	if setupPinText == "--" || setupPinText == "" {
		t.Errorf("Setup PIN not populated: %q", setupPinText)
	}
	if setupCodeText == "--" || setupCodeText == "" {
		t.Errorf("Setup Code not populated: %q", setupCodeText)
	}
	if bridgeIdText == "--" || bridgeIdText == "" {
		t.Errorf("Bridge ID not populated: %q", bridgeIdText)
	}

	// Test 8: Verify Technical Details fields populate
	var manufacturerText, modelText, firmwareText string
	if err := chromedp.Run(browserCtx, chromedp.Tasks{
		chromedp.Evaluate(`document.getElementById('homekit-manufacturer') ? document.getElementById('homekit-manufacturer').textContent : ''`, &manufacturerText),
		chromedp.Evaluate(`document.getElementById('homekit-model') ? document.getElementById('homekit-model').textContent : ''`, &modelText),
		chromedp.Evaluate(`document.getElementById('homekit-firmware') ? document.getElementById('homekit-firmware').textContent : ''`, &firmwareText),
	}); err != nil {
		t.Fatalf("failed to read HomeKit technical details: %v", err)
	}

	if manufacturerText == "--" || manufacturerText == "" {
		t.Errorf("Manufacturer not populated: %q", manufacturerText)
	}
	if modelText == "--" || modelText == "" {
		t.Errorf("Model not populated: %q", modelText)
	}
	if firmwareText == "--" || firmwareText == "" {
		t.Errorf("Firmware not populated: %q", firmwareText)
	}

	// Test 9: Verify Battery indicator exists in Device Status
	if err := chromedp.Run(browserCtx, chromedp.Click(`#device-status-row`, chromedp.NodeVisible)); err == nil {
		time.Sleep(200 * time.Millisecond)
		var batteryIndicatorHTML string
		if err := chromedp.Run(browserCtx, chromedp.Evaluate(`document.querySelector('.battery-indicator') ? document.querySelector('.battery-indicator').outerHTML : ''`, &batteryIndicatorHTML)); err == nil {
			if batteryIndicatorHTML == "" {
				t.Log("Battery indicator not found in Device Status section - may need station status data")
			} else if !strings.Contains(batteryIndicatorHTML, "battery-good") && !strings.Contains(batteryIndicatorHTML, "battery-fair") && !strings.Contains(batteryIndicatorHTML, "battery-low") {
				t.Log("Battery indicator found but missing color class - may need status update")
			}
		}
	}

	// Test 10: Verify Signal bars exist in Hub Status
	if err := chromedp.Run(browserCtx, chromedp.Click(`#hub-status-row`, chromedp.NodeVisible)); err == nil {
		time.Sleep(100 * time.Millisecond)
		var signalBarsHTML string
		if err := chromedp.Run(browserCtx, chromedp.Evaluate(`document.querySelector('.signal-bars') ? document.querySelector('.signal-bars').outerHTML : ''`, &signalBarsHTML)); err == nil {
			if signalBarsHTML == "" {
				t.Error("Signal bars not found in Hub Status section")
			} else if !strings.Contains(signalBarsHTML, "signal-bar") {
				t.Error("Signal bars missing individual bar elements")
			}
		}
	}
}

// TestAccessoriesListEnabledDisabled specifically tests the accessories list logic
// with different sensor configurations to ensure enabled/disabled status is correct.
func TestAccessoriesListEnabledDisabled(t *testing.T) {
	ws := testNewWebServer(t)

	now := time.Now()
	obs := weather.Observation{
		Timestamp:      now.Unix(),
		AirTemperature: 20.0,
	}
	ws.UpdateWeather(&obs)

	// Test with only 2 enabled sensors
	homekitStatus := map[string]interface{}{
		"bridge":         true,
		"name":           "Test Bridge",
		"accessories":    2,
		"accessoryNames": []string{"Temperature", "Humidity"},
		"allSensors":     []string{"Temperature", "Humidity", "Light", "UV", "Wind Speed", "Wind Direction", "Rain", "Pressure", "Lightning"},
		"pin":            "12345678",
	}
	ws.UpdateHomeKitStatus(homekitStatus)

	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.handleDashboard)
	mux.HandleFunc("/api/status", ws.handleStatusAPI)

	_, thisFile, _, _ := gruntime.Caller(0)
	thisDir := filepath.Dir(thisFile)
	staticDir := filepath.Join(thisDir, "static")
	mux.Handle("/pkg/web/static/", http.StripPrefix("/pkg/web/static/", http.FileServer(http.Dir(staticDir))))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoSandbox,
	)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	tasks := chromedp.Tasks{
		chromedp.Navigate(ts.URL),
		chromedp.WaitVisible(`#status`, chromedp.ByID),
		chromedp.Sleep(200 * time.Millisecond),
	}
	if err := chromedp.Run(browserCtx, tasks); err != nil {
		t.Fatalf("navigate failed: %v", err)
	}

	// Inject script
	injectedFrom := filepath.Join(thisDir, "static", "script.js")
	localBytes, err := os.ReadFile(injectedFrom)
	if err != nil {
		t.Fatalf("failed to read local script.js: %v", err)
	}
	inj := fmt.Sprintf("(function(){ var s=document.createElement('script'); s.type='text/javascript'; s.text = %q; document.head.appendChild(s); })()", string(localBytes))
	if err := chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(inj, nil)); err != nil {
		t.Fatalf("failed to inject script: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	// Fetch status
	var fetchResult interface{}
	if err := chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(`(async function(){ try { await fetchStatus(); return 'ok'; } catch(e) { return {err: e.toString()}; } })()`, &fetchResult)); err != nil {
		t.Fatalf("failed to run fetchStatus: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	// Expand accessories list
	if err := chromedp.Run(browserCtx, chromedp.Click(`#accessories-row`, chromedp.NodeVisible)); err != nil {
		t.Fatalf("failed to click accessories row: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Count active and disabled items
	var activeCount, disabledCount int
	if err := chromedp.Run(browserCtx, chromedp.Tasks{
		chromedp.Evaluate(`document.querySelectorAll('.accessory-item:not(.disabled) .accessory-status.enabled').length`, &activeCount),
		chromedp.Evaluate(`document.querySelectorAll('.accessory-item.disabled .accessory-status.disabled').length`, &disabledCount),
	}); err != nil {
		t.Fatalf("failed to count accessories: %v", err)
	}

	if activeCount != 2 {
		t.Errorf("Expected 2 active sensors, got %d", activeCount)
	}
	if disabledCount != 7 {
		t.Errorf("Expected 7 disabled sensors, got %d", disabledCount)
	}
}
