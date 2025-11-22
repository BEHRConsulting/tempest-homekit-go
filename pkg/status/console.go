// Package status provides a curses-based terminal UI for monitoring the Tempest HomeKit service.
package status

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/service"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// LogBuffer is a thread-safe circular buffer for capturing log messages
type LogBuffer struct {
	mu      sync.Mutex
	lines   []string
	maxSize int
}

// NewLogBuffer creates a new log buffer with the specified maximum size
func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		lines:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

// Write implements io.Writer interface for capturing log output
func (lb *LogBuffer) Write(p []byte) (n int, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Split incoming data into lines
	data := string(p)
	lines := strings.Split(strings.TrimRight(data, "\n"), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Strip ANSI color codes for cleaner display
		line = stripANSI(line)
		lb.lines = append(lb.lines, line)
		if len(lb.lines) > lb.maxSize {
			lb.lines = lb.lines[1:]
		}
	}

	return len(p), nil
}

// GetLines returns a copy of all stored log lines
func (lb *LogBuffer) GetLines() []string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	result := make([]string, len(lb.lines))
	copy(result, lb.lines)
	return result
}

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	// Simple ANSI stripper - removes escape sequences
	result := ""
	inEscape := false
	for _, ch := range s {
		if ch == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
			continue
		}
		result += string(ch)
	}
	return result
}

// RunStatusConsole starts the curses-based status console
func RunStatusConsole(cfg *config.Config, version string) error {
	// Get theme (mutable for theme cycling)
	currentThemeName := cfg.StatusTheme
	theme := GetTheme(currentThemeName)

	// Create log buffer to capture stdout/stderr
	logBuf := NewLogBuffer(1000)

	// Create pipe for capturing output
	r, w, _ := os.Pipe()

	// Also open a debug log file so we can capture logs outside the TUI
	debugFile, err := os.OpenFile("/tmp/status_console_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// If we cannot open the debug file, continue without it
		log.Printf("warning: could not open debug log file: %v", err)
		debugFile = nil
	}

	// Redirect log output to pipe (and debug file) so tview doesn't consume stdout
	if debugFile != nil {
		log.SetOutput(io.MultiWriter(w, debugFile))
	} else {
		log.SetOutput(w)
	}

	// Start goroutine to read from pipe into buffer
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if err != nil {
				break
			}
			if n > 0 {
				logBuf.Write(buf[:n])
			}
		}
	}()

	// Resize events are handled by tcell; additional OS signal logging requires
	// careful integration and is skipped here to avoid interference.

	// Create context for goroutine coordination
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start service in background AFTER log redirection is set up
	go func() {
		if err := service.StartService(cfg, version); err != nil {
			log.Printf("Service error: %v", err)
		}
	}()

	// Give service a moment to initialize
	time.Sleep(2 * time.Second)

	// Restore log output when done
	defer func() {
		w.Close()
		if debugFile != nil {
			debugFile.Close()
		}
		log.SetOutput(os.Stderr)
	}()

	// Create tview application
	app := tview.NewApplication()

	// Create main layout with flex boxes
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Title
	title := tview.NewTextView().
		SetText(fmt.Sprintf(" Tempest HomeKit v%s ", version)).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(theme.TitleColor)
	mainFlex.AddItem(title, 1, 0, false)

	// Content area - split into left (logs + station) and right (alarms + homekit)
	contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Left column
	leftFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Console Logs (top-left)
	tvConsole := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.QueueUpdateDraw(func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("panic in tvConsole QueueUpdateDraw: %v", r)
					}
				}()
			})
		})
	tvConsole.SetBorder(true).SetTitle(" Console Logs ").SetBorderColor(theme.BorderColor)
	leftFlex.AddItem(tvConsole, 0, 3, false) // 3/6 of left column

	// Tempest Sensors (middle-left)
	tvSensors := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	tvSensors.SetBorder(true).SetTitle(" Tempest Sensors ").SetBorderColor(theme.BorderColor)
	leftFlex.AddItem(tvSensors, 0, 2, false) // 2/6 of left column

	// Station Status (bottom-left)
	tvStation := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	tvStation.SetBorder(true).SetTitle(" Station Status ").SetBorderColor(theme.BorderColor)
	leftFlex.AddItem(tvStation, 0, 1, false) // 1/6 of left column

	contentFlex.AddItem(leftFlex, 0, 1, false) // Left column is 1/2 of width

	// Right column
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Alarm Status (top-right)
	tvAlarms := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	tvAlarms.SetBorder(true).SetTitle(" Alarm Status ").SetBorderColor(theme.BorderColor)
	rightFlex.AddItem(tvAlarms, 0, 1, false)

	// HomeKit Status (middle-right)
	tvHomeKit := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	tvHomeKit.SetBorder(true).SetTitle(" HomeKit Status ").SetBorderColor(theme.BorderColor)
	rightFlex.AddItem(tvHomeKit, 0, 1, false)

	// System Info (bottom-right)
	tvSystem := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	tvSystem.SetBorder(true).SetTitle(" System Info ").SetBorderColor(theme.BorderColor)
	rightFlex.AddItem(tvSystem, 0, 1, false)

	contentFlex.AddItem(rightFlex, 0, 1, false) // Right column is 1/2 of width

	mainFlex.AddItem(contentFlex, 0, 1, false)

	// Footer with timers
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(theme.FooterColor)
	mainFlex.AddItem(footer, 1, 0, false)

	// Set initial footer
	startTime := time.Now()
	updateFooter := func(nextRefresh int) {
		elapsed := time.Since(startTime)
		hours := int(elapsed.Hours())
		minutes := int(elapsed.Minutes()) % 60
		seconds := int(elapsed.Seconds()) % 60
		currentTheme := GetTheme(currentThemeName) // Look up current theme
		timerTag := colorToTviewTag(currentTheme.TimerColor)
		footer.SetText(fmt.Sprintf(" Running: [%s]%02d:%02d:%02d[-] | Refresh: [%s]%ds[-] | Theme: [%s]%s[-] | 'q':quit 'r':refresh 't':theme ",
			timerTag, hours, minutes, seconds, timerTag, nextRefresh, timerTag, currentThemeName))
	}
	updateFooter(cfg.StatusRefresh)

	// Function to apply theme to all UI elements
	applyTheme := func() {
		currentTheme := GetTheme(currentThemeName) // Look up current theme
		title.SetTextColor(currentTheme.TitleColor)
		tvConsole.SetBorderColor(currentTheme.BorderColor)
		tvSensors.SetBorderColor(currentTheme.BorderColor)
		tvStation.SetBorderColor(currentTheme.BorderColor)
		tvAlarms.SetBorderColor(currentTheme.BorderColor)
		tvHomeKit.SetBorderColor(currentTheme.BorderColor)
		tvSystem.SetBorderColor(currentTheme.BorderColor)
		footer.SetTextColor(currentTheme.FooterColor)
	}

	// Function to fetch and update status data
	baseURL := fmt.Sprintf("http://localhost:%s", cfg.WebPort)
	updateStatus := func() {
		// Check if context is cancelled before doing expensive work
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Fetch all data first (outside UI update to avoid blocking)
		weatherData, weatherErr := fetchStatus(baseURL + "/api/weather")
		statusData, statusErr := fetchStatus(baseURL + "/api/status")
		alarmData, alarmErr := fetchStatus(baseURL + "/api/alarm-status")

		// Check again before UI update
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Schedule UI update on the application's event loop (with recover)
		app.QueueUpdateDraw(func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("panic in updateStatus QueueUpdateDraw: %v", r)
				}
			}()
			currentTheme := GetTheme(currentThemeName)
			labelTag := colorToTviewTag(currentTheme.LabelColor)
			valueTag := colorToTviewTag(currentTheme.ValueColor)
			successTag := colorToTviewTag(currentTheme.SuccessColor)
			dangerTag := colorToTviewTag(currentTheme.DangerColor)
			warningTag := colorToTviewTag(currentTheme.WarningColor)
			mutedTag := colorToTviewTag(currentTheme.MutedColor)

			// Update console logs
			logLines := logBuf.GetLines()
			tvConsole.Clear()
			for _, line := range logLines {
				coloredLine := currentTheme.ColorizeLogLine(line)
				fmt.Fprintln(tvConsole, coloredLine)
			}
			tvConsole.ScrollToEnd()

			// Update Tempest sensor values
			tvSensors.Clear()
			if weatherErr == nil {
				if temp, ok := weatherData["temperature"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]Temperature:[-] [%s]%.1f°C[-]\n", labelTag, valueTag, temp)
				}
				if humidity, ok := weatherData["humidity"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]Humidity:[-] [%s]%.0f%%[-]\n", labelTag, valueTag, humidity)
				}
				if pressure, ok := weatherData["pressure"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]Pressure:[-] [%s]%.1f mb[-]\n", labelTag, valueTag, pressure)
				}
				if windSpeed, ok := weatherData["windSpeed"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]Wind Speed:[-] [%s]%.1f mph[-]\n", labelTag, valueTag, windSpeed)
				}
				if windGust, ok := weatherData["windGust"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]Wind Gust:[-] [%s]%.1f mph[-]\n", labelTag, valueTag, windGust)
				}
				if windDir, ok := weatherData["windDirection"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]Wind Direction:[-] [%s]%.0f°[-]\n", labelTag, valueTag, windDir)
				}
				if rain, ok := weatherData["rainAccum"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]Rain Accum:[-] [%s]%.3f in[-]\n", labelTag, valueTag, rain)
				}
				if illuminance, ok := weatherData["illuminance"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]Illuminance:[-] [%s]%.0f lux[-]\n", labelTag, valueTag, illuminance)
				}
				if uv, ok := weatherData["uv"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]UV Index:[-] [%s]%.0f[-]\n", labelTag, valueTag, uv)
				}
				if battery, ok := weatherData["battery"].(float64); ok {
					fmt.Fprintf(tvSensors, "[%s]Battery:[-] [%s]%.2fV[-]\n", labelTag, valueTag, battery)
				}
			}

			// Update station status
			tvStation.Clear()
			if statusErr == nil {
				fmt.Fprintf(tvStation, "[%s]Connected:[-] [%s]%v[-]\n", labelTag, valueTag, statusData["connected"])
				if station, ok := statusData["station"].(string); ok {
					fmt.Fprintf(tvStation, "[%s]Station:[-] [%s]%s[-]\n", labelTag, valueTag, station)
				}
				if dataSource, ok := statusData["dataSource"].(string); ok {
					fmt.Fprintf(tvStation, "[%s]Data Source:[-] [%s]%s[-]\n", labelTag, valueTag, dataSource)
				}
				if lastUpdate, ok := statusData["lastUpdate"].(string); ok {
					fmt.Fprintf(tvStation, "[%s]Last Update:[-] [%s]%s[-]\n", labelTag, valueTag, lastUpdate)
				}
				if history, ok := statusData["dataHistory"].([]interface{}); ok {
					fmt.Fprintf(tvStation, "[%s]History:[-] [%s]%d pts[-]\n", labelTag, valueTag, len(history))
				}
			}

			// Update alarm status
			tvAlarms.Clear()
			if alarmErr == nil {
				if enabled, ok := alarmData["enabled"].(bool); ok {
					fmt.Fprintf(tvAlarms, "[%s]Enabled:[-] [%s]%v[-]\n", labelTag, valueTag, enabled)
				}
				if total, ok := alarmData["totalAlarms"].(float64); ok {
					fmt.Fprintf(tvAlarms, "[%s]Total:[-] [%s]%.0f[-]\n", labelTag, valueTag, total)
				}
				if enabledCount, ok := alarmData["enabledAlarms"].(float64); ok {
					fmt.Fprintf(tvAlarms, "[%s]Active:[-] [%s]%.0f[-]\n\n", labelTag, valueTag, enabledCount)
				}

				// Show triggered and cooling down alarms
				if alarms, ok := alarmData["alarms"].([]interface{}); ok {
					triggered := []string{}
					cooling := []string{}
					for _, a := range alarms {
						if alarmMap, ok := a.(map[string]interface{}); ok {
							name, _ := alarmMap["name"].(string)
							lastTriggered, _ := alarmMap["lastTriggered"].(string)
							inCooldown, _ := alarmMap["inCooldown"].(bool)
							cooldownRemaining, _ := alarmMap["cooldownRemaining"].(float64)

							if lastTriggered != "" && lastTriggered != "Never" {
								triggered = append(triggered, fmt.Sprintf("%s (%s)", name, lastTriggered))
							}
							if inCooldown && cooldownRemaining > 0 {
								cooling = append(cooling, fmt.Sprintf("%s (%.0fs)", name, cooldownRemaining))
							}
						}
					}

					if len(triggered) > 0 {
						fmt.Fprintf(tvAlarms, "[%s]Triggered:[-]\n", dangerTag)
						for i, t := range triggered {
							if i >= 3 {
								break
							}
							fmt.Fprintf(tvAlarms, "  [%s]%s[-]\n", valueTag, t)
						}
						if len(triggered) > 3 {
							fmt.Fprintf(tvAlarms, "  [%s]+%d more...[-]\n", valueTag, len(triggered)-3)
						}
						fmt.Fprintf(tvAlarms, "\n")
					}

					if len(cooling) > 0 {
						fmt.Fprintf(tvAlarms, "[%s]Cooling Down:[-]\n", warningTag)
						for i, c := range cooling {
							if i >= 3 {
								break
							}
							fmt.Fprintf(tvAlarms, "  [%s]%s[-]\n", valueTag, c)
						}
						if len(cooling) > 3 {
							fmt.Fprintf(tvAlarms, "  [%s]+%d more...[-]\n", valueTag, len(cooling)-3)
						}
					}
				}
			}

			// Update HomeKit status
			tvHomeKit.Clear()
			if statusErr == nil {
				if homekit, ok := statusData["homekit"].(map[string]interface{}); ok {
					if bridge, ok := homekit["bridge"].(bool); ok {
						if bridge {
							fmt.Fprintf(tvHomeKit, "[%s]Status:[-] [%s]Active[-]\n", labelTag, successTag)
						} else {
							fmt.Fprintf(tvHomeKit, "[%s]Status:[-] [%s]Disabled[-]\n", labelTag, dangerTag)
						}
					}

					if pin, ok := homekit["pin"].(string); ok {
						fmt.Fprintf(tvHomeKit, "[%s]PIN:[-] [%s]%s[-]\n", labelTag, valueTag, pin)
					}

					if accessories, ok := homekit["accessories"].(float64); ok {
						fmt.Fprintf(tvHomeKit, "[%s]Accessories:[-] [%s]%.0f[-]\n\n", labelTag, valueTag, accessories)
					}

					// Show published sensors
					if accessoryNames, ok := homekit["accessoryNames"].([]interface{}); ok && len(accessoryNames) > 0 {
						fmt.Fprintf(tvHomeKit, "[%s]Published Sensors:[-]\n", labelTag)
						for _, acc := range accessoryNames {
							if name, ok := acc.(string); ok {
								fmt.Fprintf(tvHomeKit, "  [%s]• %s[-]\n", successTag, name)
							}
						}
					} else {
						fmt.Fprintf(tvHomeKit, "[%s]No sensors published[-]\n", mutedTag)
					}
				}
			}

			// Update system info
			tvSystem.Clear()
			fmt.Fprintf(tvSystem, "[%s]Web Port:[-] [%s]%s[-]\n", labelTag, valueTag, cfg.WebPort)
			fmt.Fprintf(tvSystem, "[%s]Log Level:[-] [%s]%s[-]\n", labelTag, valueTag, cfg.LogLevel)
			fmt.Fprintf(tvSystem, "[%s]Refresh Interval:[-] [%s]%ds[-]\n", labelTag, valueTag, cfg.StatusRefresh)
			if cfg.StatusTimeout > 0 {
				remaining := cfg.StatusTimeout - int(time.Since(startTime).Seconds())
				if remaining < 0 {
					remaining = 0
				}
				fmt.Fprintf(tvSystem, "[%s]Timeout:[-] [%s]%ds (remaining: %ds)[-]\n", labelTag, valueTag, cfg.StatusTimeout, remaining)
			} else {
				fmt.Fprintf(tvSystem, "[%s]Timeout:[-] [%s]Never[-]\n", labelTag, valueTag)
			}
			fmt.Fprintf(tvSystem, "[%s]Theme:[-] [%s]%s[-]\n", labelTag, valueTag, currentThemeName)
		})
	}

	// Shared state for refresh countdown
	var refreshMutex sync.Mutex
	nextRefreshSeconds := cfg.StatusRefresh

	// Initial update - must be done AFTER app starts, so schedule it
	go func() {
		time.Sleep(100 * time.Millisecond) // Give app time to start
		updateStatus()
	}()

	// Auto-refresh goroutine
	refreshTicker := time.NewTicker(time.Duration(cfg.StatusRefresh) * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-refreshTicker.C:
				updateStatus()
				refreshMutex.Lock()
				nextRefreshSeconds = cfg.StatusRefresh
				refreshMutex.Unlock()
			}
		}
	}()
	defer refreshTicker.Stop()

	// Footer timer goroutine
	footerTicker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-footerTicker.C:
				refreshMutex.Lock()
				nextRefreshSeconds--
				if nextRefreshSeconds <= 0 {
					nextRefreshSeconds = cfg.StatusRefresh
				}
				currentNext := nextRefreshSeconds
				refreshMutex.Unlock()

				// Schedule footer update on the application's event loop
				app.QueueUpdateDraw(func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("panic in footer QueueUpdateDraw: %v", r)
						}
					}()
					updateFooter(currentNext)
				})

				// Check for timeout
				if cfg.StatusTimeout > 0 && time.Since(startTime).Seconds() >= float64(cfg.StatusTimeout) {
					cancel()
					app.Stop()
				}
			}
		}
	}()
	defer footerTicker.Stop()

	// Input handling with non-blocking approach
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				// Stop everything immediately
				cancel()
				app.Stop()
				return nil
			case 'r', 'R':
				// Manual refresh - reset countdown and trigger immediate update
				log.Printf("input: 'r' pressed - enqueueing refresh")
				refreshMutex.Lock()
				nextRefreshSeconds = cfg.StatusRefresh
				refreshMutex.Unlock()
				// Force an immediate UI repaint and run update in background
				app.QueueUpdateDraw(func() {})
				go updateStatus()
				return nil
			case 't', 'T':
				// Cycle to next theme - do it synchronously in the event handler
				nextTheme := GetNextTheme(currentThemeName)
				currentThemeName = nextTheme.Name
				// Apply changes directly (we're already in the event loop)
				applyTheme()
				refreshMutex.Lock()
				currentNext := nextRefreshSeconds
				refreshMutex.Unlock()
				updateFooter(currentNext)
				// The next auto-refresh will pick up the new theme
				return nil
			}
		case tcell.KeyEscape, tcell.KeyCtrlC:
			// Stop everything immediately
			cancel()
			app.Stop()
			return nil
		case tcell.KeyCtrlL:
			// Treat Ctrl-L (clear screen) like a manual refresh: force redraw
			refreshMutex.Lock()
			nextRefreshSeconds = cfg.StatusRefresh
			refreshMutex.Unlock()
			app.QueueUpdateDraw(func() {})
			go updateStatus()
			return nil
		}
		return event
	})

	// Set root and run
	app.SetRoot(mainFlex, true).SetFocus(mainFlex)

	// Enable mouse support
	app.EnableMouse(true)

	// Start the app (this will take over the terminal screen)
	if err := app.Run(); err != nil {
		return fmt.Errorf("failed to run status console: %v", err)
	}

	return nil
}

// fetchStatus fetches JSON data from the given URL
func fetchStatus(url string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
