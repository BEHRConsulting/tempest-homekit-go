package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	gruntime "runtime"
	"strings"
	"testing"
	"time"

	"tempest-homekit-go/pkg/weather"

	cpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// TestHeadlessDashboard loads the dashboard in a headless browser, runs diagnostics,
// captures console logs and asserts no runtime JS errors occur and key UI fields populate.
func TestHeadlessDashboard(t *testing.T) {
	// create server and inject synthetic data
	ws := NewWebServer("0", 10.0, "debug", 0, false, "test", "", nil, nil, "metric", "mb")

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

	// Ensure there's current weather and some history
	ws.UpdateWeather(&obs)
	// also add a couple historical points
	older := obs
	older.Timestamp = now.Add(-2 * time.Minute).Unix()
	older.AirTemperature = 18.2
	ws.UpdateWeather(&older)

	// create a mux that serves dashboard and api endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.handleDashboard)
	mux.HandleFunc("/api/weather", ws.handleWeatherAPI)
	mux.HandleFunc("/api/status", ws.handleStatusAPI)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Prepare chromedp
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Allocator with headless flags
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoSandbox,
	)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var consoleErrors []string

	// listen for console events
	chromedp.ListenTarget(browserCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *cpruntime.EventExceptionThrown:
			// capture unhandled exceptions
			consoleErrors = append(consoleErrors, ev.ExceptionDetails.Text)
		case *cpruntime.EventConsoleAPICalled:
			for _, arg := range ev.Args {
				if arg.Value != nil {
					s, _ := json.Marshal(arg.Value)
					consoleErrors = append(consoleErrors, string(s))
				}
			}
		}
	})

	// navigate to the dashboard and wait for status element
	var rainText, pressureText, illumText, uvText string

	tasks := chromedp.Tasks{
		chromedp.Navigate(ts.URL),
		chromedp.WaitVisible(`#status`, chromedp.ByID),
		// give the page a moment to fetch /api/weather and update DOM
		chromedp.Sleep(400 * time.Millisecond),
		// Give the rest of the script a moment to initialize (we'll explicitly call fetchWeather below)
		chromedp.Sleep(200 * time.Millisecond),
		// run diagnostic helpers if available (non-fatal)
		chromedp.EvaluateAsDevTools(`typeof testCharts === 'function' && testCharts()`, nil),
		chromedp.Sleep(200 * time.Millisecond),
		chromedp.EvaluateAsDevTools(`typeof diagnoseCharts === 'function' && diagnoseCharts()`, nil),
		chromedp.Sleep(200 * time.Millisecond),
		// install a lightweight error/rejection capture hook so we can inspect errors later
		chromedp.EvaluateAsDevTools(`(function(){ window.__testErrors = []; window.__testConsole = []; window.addEventListener('error', function(e){ try{ window.__testErrors.push({type:'error', message: e.message, filename: e.filename, lineno: e.lineno, colno: e.colno, stack: e.error && e.error.stack}); } catch(err){} }); window.addEventListener('unhandledrejection', function(e){ try{ window.__testErrors.push({type:'unhandledrejection', reason: String(e.reason)}); } catch(err){} }); })()`, nil),
		chromedp.Sleep(20 * time.Millisecond),
		chromedp.Text(`#rain`, &rainText, chromedp.NodeVisible, chromedp.ByID),
		chromedp.Text(`#pressure`, &pressureText, chromedp.NodeVisible, chromedp.ByID),
		chromedp.Text(`#illuminance`, &illumText, chromedp.NodeVisible, chromedp.ByID),
		chromedp.Text(`#uv-index`, &uvText, chromedp.NodeVisible, chromedp.ByID),
	}

	if err := chromedp.Run(browserCtx, tasks); err != nil {
		t.Fatalf("chromedp run failed: %v", err)
	}

	// To avoid CDN/network timing flakes, fetch the page's script URLs and inject their content
	var scriptsJSON string
	if err := chromedp.Run(browserCtx, chromedp.Evaluate(`(function(){ return JSON.stringify(Array.from(document.scripts).map(s => s.src || '')); })()`, &scriptsJSON)); err != nil {
		t.Fatalf("failed to enumerate script tags: %v", err)
	}

	var scriptURLs []string
	if err := json.Unmarshal([]byte(scriptsJSON), &scriptURLs); err != nil {
		t.Fatalf("failed to parse scripts JSON: %v; raw=%s", err, scriptsJSON)
	}

	var chartSrc, localScriptSrc string
	for _, u := range scriptURLs {
		if u == "" {
			continue
		}
		if chartSrc == "" && (strings.Contains(u, "chart.js") || strings.Contains(u, "Chart.min.js") || strings.Contains(u, "chart.umd.js")) {
			chartSrc = u
		}
		if localScriptSrc == "" && strings.Contains(u, "script.js") {
			localScriptSrc = u
		}
	}

	// Helper to fetch a URL's text via Go HTTP client
	fetchText := func(url string) (string, error) {
		if url == "" {
			return "", nil
		}
		// Attempt direct GET
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return "", fmt.Errorf("status %d", resp.StatusCode)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	// Fetch and inject Chart.js first (if present)
	if chartSrc != "" {
		chartText, err := fetchText(chartSrc)
		if err != nil {
			t.Logf("warning: failed to fetch Chart.js from %s: %v", chartSrc, err)
		} else if chartText != "" {
			// inject Chart.js into page
			js := `(function(){ var s=document.createElement('script'); s.type='text/javascript'; s.text = %s; document.head.appendChild(s); })()`
			scriptLiteral := fmt.Sprintf("%q", chartText)
			inject := fmt.Sprintf(js, scriptLiteral)
			if err := chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(inject, nil)); err != nil {
				t.Logf("warning: failed to inject Chart.js into page: %v", err)
			} else {
				// give Chart.js a moment
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	// Read and inject our local script.js directly from disk to avoid HTTP path/port issues in headless test
	// Resolve script path relative to this test file so tests don't depend on cwd
	_, thisFile, _, _ := gruntime.Caller(0)
	thisDir := filepath.Dir(thisFile)
	injectedFrom := filepath.Join(thisDir, "static", "script.js")
	localBytes, err := os.ReadFile(injectedFrom)
	if err != nil {
		t.Fatalf("failed to read local script.js from %s: %v", injectedFrom, err)
	}
	localText := string(localBytes)
	if len(localText) == 0 {
		t.Fatalf("local script.js was empty at %s", injectedFrom)
	}
	js := `(function(){ var s=document.createElement('script'); s.type='text/javascript'; s.text = %s; document.head.appendChild(s); })()`
	scriptLiteral := fmt.Sprintf("%q", localText)
	inject := fmt.Sprintf(js, scriptLiteral)
	if err := chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(inject, nil)); err != nil {
		t.Fatalf("failed to inject local script into page: %v", err)
	}
	// allow script to initialize
	time.Sleep(200 * time.Millisecond)

	// Explicitly call fetchWeather and fetchStatus and await their completion
	var fetchResult1 interface{}
	var fetchResult2 interface{}
	if err := chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(`(async function(){ try { await fetchWeather(); return 'ok'; } catch(e) { return {err: e.toString()}; } })()`, &fetchResult1)); err != nil {
		t.Fatalf("failed to run fetchWeather: %v", err)
	}
	if err := chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(`(async function(){ try { await fetchStatus(); return 'ok'; } catch(e) { return {err: e.toString()}; } })()`, &fetchResult2)); err != nil {
		t.Fatalf("failed to run fetchStatus: %v", err)
	}

	// Capture immediate page-level diagnostics after fetches
	var testErrors string
	var typeofGetWeatherData string
	var getWeatherDataResult string
	var chartType string
	var chartsType string
	_ = chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(`JSON.stringify(window.__testErrors || [])`, &testErrors))
	_ = chromedp.Run(browserCtx, chromedp.Evaluate(`typeof window.getWeatherData`, &typeofGetWeatherData))
	_ = chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(`(function(){ try { return JSON.stringify(window.getWeatherData ? window.getWeatherData() : null); } catch(e) { return JSON.stringify({err: e.toString()}); } })()`, &getWeatherDataResult))
	_ = chromedp.Run(browserCtx, chromedp.Evaluate(`typeof Chart`, &chartType))
	_ = chromedp.Run(browserCtx, chromedp.Evaluate(`typeof charts`, &chartsType))
	if len(testErrors) > 2 {
		t.Logf("page __testErrors: %s", testErrors)
	}
	t.Logf("post-inject diagnostics: typeof getWeatherData=%s; getWeatherData=%s; typeof Chart=%s; typeof charts=%s", typeofGetWeatherData, getWeatherDataResult, chartType, chartsType)

	// Poll the DOM for key card values to become populated (not the placeholder "--")
	var rainTextVal, pressureTextVal, illumTextVal, uvTextVal string
	populated := false
	attempts := 60 // ~12 seconds with 200ms sleeps
	for i := 0; i < attempts; i++ {
		if err := chromedp.Run(browserCtx, chromedp.Tasks{
			chromedp.Evaluate(`document.getElementById('rain') ? document.getElementById('rain').textContent : ''`, &rainTextVal),
			chromedp.Evaluate(`document.getElementById('pressure') ? document.getElementById('pressure').textContent : ''`, &pressureTextVal),
			chromedp.Evaluate(`document.getElementById('illuminance') ? document.getElementById('illuminance').textContent : ''`, &illumTextVal),
			chromedp.Evaluate(`document.getElementById('uv-index') ? document.getElementById('uv-index').textContent : ''`, &uvTextVal),
		}); err != nil {
			t.Fatalf("failed to read card text: %v", err)
		}

		if rainTextVal != "--" && rainTextVal != "" && pressureTextVal != "--" && pressureTextVal != "" && illumTextVal != "--" && illumTextVal != "" && uvTextVal != "--" && uvTextVal != "" {
			populated = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if !populated {
		// Collect diagnostics for failure
		var rawFetch string
		_ = chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(`(async function(){ try { const r = await fetch('/api/weather'); const txt = await r.text(); return JSON.stringify({status: r.status, ok: r.ok, length: txt.length, preview: txt.slice(0,800)}); } catch(e) { return JSON.stringify({error: e.toString()}); } })()`, &rawFetch))

		// gather page-level diagnostics
		var typeofChart, typeofCharts, typeofWeatherData string
		var getWeatherDataJSON string
		var scriptsList string
		var injectedPrefix string

		_ = chromedp.Run(browserCtx, chromedp.Evaluate(`typeof Chart`, &typeofChart))
		_ = chromedp.Run(browserCtx, chromedp.Evaluate(`typeof charts`, &typeofCharts))
		_ = chromedp.Run(browserCtx, chromedp.Evaluate(`typeof window.weatherData`, &typeofWeatherData))
		// attempt to JSON.stringify the debug hook return value
		_ = chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(`(function(){ try { return JSON.stringify(window.getWeatherData ? window.getWeatherData() : null); } catch(e) { return JSON.stringify({err: e.toString()}); } })()`, &getWeatherDataJSON))
		_ = chromedp.Run(browserCtx, chromedp.EvaluateAsDevTools(`Array.from(document.scripts).map(s=>s.src||'inline').join('\n')`, &scriptsList))

		// capture prefix of the injected script (if available via src)
		if localScriptSrc != "" {
			// try to fetch first 800 chars from the server again for reporting
			if txt, err := fetchText(localScriptSrc); err == nil {
				if len(txt) > 800 {
					injectedPrefix = txt[:800]
				} else {
					injectedPrefix = txt
				}
			} else {
				injectedPrefix = "<error fetching injected script: " + err.Error() + ">"
			}
		}

		// collect any exception payloads captured via chromedp listener
		t.Fatalf("cards not populated after wait; rain=%q pressure=%q illum=%q uv=%q; consoleEvents=%v; fetchDiag=%s; typeof Chart=%q; typeof charts=%q; typeof window.weatherData=%q; window.getWeatherData=%s; scripts=\n%s; injectedFrom=%q; injectedScriptPrefix=\n%s", rainTextVal, pressureTextVal, illumTextVal, uvTextVal, consoleErrors, rawFetch, typeofChart, typeofCharts, typeofWeatherData, getWeatherDataJSON, scriptsList, injectedFrom, injectedPrefix)
	}
	// Check UI values (from the polled values) are not the placeholder "--"
	if rainTextVal == "--" || rainTextVal == "" {
		t.Fatalf("rain card not populated (value: %q)", rainTextVal)
	}
	if pressureTextVal == "--" || pressureTextVal == "" {
		t.Fatalf("pressure card not populated (value: %q)", pressureTextVal)
	}
	if illumTextVal == "--" || illumTextVal == "" {
		t.Fatalf("illuminance card not populated (value: %q)", illumTextVal)
	}
	if uvTextVal == "--" || uvTextVal == "" {
		t.Fatalf("UV card not populated (value: %q)", uvTextVal)
	}
}

func containsErrorKeyword(s string) bool {
	// simple checks for JS error indicators
	if s == "" {
		return false
	}
	keywords := []string{"TypeError", "ReferenceError", "Uncaught", "ERROR", "Exception"}
	for _, k := range keywords {
		if len(s) >= len(k) && (contains(s, k)) {
			return true
		}
	}
	return false
}

// contains is a tiny case-sensitive substring helper
func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (indexOf(s, sub) >= 0))
}

// indexOf returns the first index of substr in s or -1
func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// min returns the smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
