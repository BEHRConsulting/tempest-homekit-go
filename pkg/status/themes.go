// Package status provides themed color schemes for the terminal UI
package status

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Theme defines color scheme for the status console
type Theme struct {
	Name        string
	Description string
	Background  string // "light" or "dark"

	// UI Elements
	TitleColor  tcell.Color
	BorderColor tcell.Color
	LabelColor  tcell.Color
	ValueColor  tcell.Color
	FooterColor tcell.Color
	TimerColor  tcell.Color

	// Log Level Colors
	ErrorColor   tcell.Color
	WarningColor tcell.Color
	InfoColor    tcell.Color
	DebugColor   tcell.Color

	// Status Colors
	SuccessColor tcell.Color
	DangerColor  tcell.Color
	MutedColor   tcell.Color
}

// Predefined color themes
var themes = map[string]*Theme{
	// Dark background themes
	"dark-ocean": {
		Name:         "dark-ocean",
		Description:  "Deep ocean blues with aqua accents (dark background)",
		Background:   "dark",
		TitleColor:   tcell.ColorAqua,
		BorderColor:  tcell.ColorTeal,
		LabelColor:   tcell.ColorSkyblue,
		ValueColor:   tcell.ColorWhite,
		FooterColor:  tcell.ColorDarkCyan,
		TimerColor:   tcell.ColorAqua,
		ErrorColor:   tcell.ColorOrangeRed,
		WarningColor: tcell.ColorOrange,
		InfoColor:    tcell.ColorLightSkyBlue,
		DebugColor:   tcell.ColorGray,
		SuccessColor: tcell.ColorLightGreen,
		DangerColor:  tcell.ColorRed,
		MutedColor:   tcell.ColorDimGray,
	},
	"dark-forest": {
		Name:         "dark-forest",
		Description:  "Rich forest greens with emerald highlights (dark background)",
		Background:   "dark",
		TitleColor:   tcell.ColorLightGreen,
		BorderColor:  tcell.ColorGreen,
		LabelColor:   tcell.ColorLimeGreen,
		ValueColor:   tcell.ColorWhiteSmoke,
		FooterColor:  tcell.ColorDarkGreen,
		TimerColor:   tcell.ColorSpringGreen,
		ErrorColor:   tcell.ColorCrimson,
		WarningColor: tcell.ColorGold,
		InfoColor:    tcell.ColorPaleGreen,
		DebugColor:   tcell.ColorGray,
		SuccessColor: tcell.ColorGreen,
		DangerColor:  tcell.ColorRed,
		MutedColor:   tcell.ColorDarkSlateGray,
	},
	"dark-sunset": {
		Name:         "dark-sunset",
		Description:  "Warm sunset oranges and purples (dark background)",
		Background:   "dark",
		TitleColor:   tcell.ColorOrange,
		BorderColor:  tcell.ColorDarkOrange,
		LabelColor:   tcell.ColorGold,
		ValueColor:   tcell.ColorLightYellow,
		FooterColor:  tcell.ColorDarkGoldenrod,
		TimerColor:   tcell.ColorYellow,
		ErrorColor:   tcell.ColorRed,
		WarningColor: tcell.ColorOrangeRed,
		InfoColor:    tcell.ColorLightSalmon,
		DebugColor:   tcell.ColorDarkGray,
		SuccessColor: tcell.ColorLightGreen,
		DangerColor:  tcell.ColorDarkRed,
		MutedColor:   tcell.ColorGray,
	},
	"dark-twilight": {
		Name:         "dark-twilight",
		Description:  "Cool twilight purples and blues (dark background)",
		Background:   "dark",
		TitleColor:   tcell.ColorMediumPurple,
		BorderColor:  tcell.ColorPurple,
		LabelColor:   tcell.ColorLightSteelBlue,
		ValueColor:   tcell.ColorLavender,
		FooterColor:  tcell.ColorDarkSlateBlue,
		TimerColor:   tcell.ColorViolet,
		ErrorColor:   tcell.ColorHotPink,
		WarningColor: tcell.ColorPaleVioletRed,
		InfoColor:    tcell.ColorLightBlue,
		DebugColor:   tcell.ColorSlateGray,
		SuccessColor: tcell.ColorMediumSpringGreen,
		DangerColor:  tcell.ColorDeepPink,
		MutedColor:   tcell.ColorDimGray,
	},
	"dark-matrix": {
		Name:         "dark-matrix",
		Description:  "Classic terminal green on black (dark background)",
		Background:   "dark",
		TitleColor:   tcell.ColorLime,
		BorderColor:  tcell.ColorGreen,
		LabelColor:   tcell.ColorGreenYellow,
		ValueColor:   tcell.ColorLightGreen,
		FooterColor:  tcell.ColorDarkGreen,
		TimerColor:   tcell.ColorChartreuse,
		ErrorColor:   tcell.ColorRed,
		WarningColor: tcell.ColorYellow,
		InfoColor:    tcell.ColorPaleGreen,
		DebugColor:   tcell.ColorDarkGray,
		SuccessColor: tcell.ColorLime,
		DangerColor:  tcell.ColorOrangeRed,
		MutedColor:   tcell.ColorGray,
	},
	"dark-cyberpunk": {
		Name:         "dark-cyberpunk",
		Description:  "Neon pinks and cyans (dark background)",
		Background:   "dark",
		TitleColor:   tcell.ColorFuchsia,
		BorderColor:  tcell.ColorDeepPink,
		LabelColor:   tcell.ColorAqua,
		ValueColor:   tcell.ColorWhite,
		FooterColor:  tcell.ColorDarkMagenta,
		TimerColor:   tcell.ColorHotPink,
		ErrorColor:   tcell.ColorRed,
		WarningColor: tcell.ColorYellow,
		InfoColor:    tcell.ColorAqua,
		DebugColor:   tcell.ColorGray,
		SuccessColor: tcell.ColorLime,
		DangerColor:  tcell.ColorCrimson,
		MutedColor:   tcell.ColorDimGray,
	},

	// Light background themes
	"light-sky": {
		Name:         "light-sky",
		Description:  "Bright sky blues and clouds (light background)",
		Background:   "light",
		TitleColor:   tcell.ColorDodgerBlue,
		BorderColor:  tcell.ColorSteelBlue,
		LabelColor:   tcell.ColorRoyalBlue,
		ValueColor:   tcell.ColorBlack,
		FooterColor:  tcell.ColorCornflowerBlue,
		TimerColor:   tcell.ColorBlue,
		ErrorColor:   tcell.ColorDarkRed,
		WarningColor: tcell.ColorDarkOrange,
		InfoColor:    tcell.ColorNavy,
		DebugColor:   tcell.ColorDimGray,
		SuccessColor: tcell.ColorDarkGreen,
		DangerColor:  tcell.ColorRed,
		MutedColor:   tcell.ColorGray,
	},
	"light-garden": {
		Name:         "light-garden",
		Description:  "Fresh garden greens and earth tones (light background)",
		Background:   "light",
		TitleColor:   tcell.ColorForestGreen,
		BorderColor:  tcell.ColorSeaGreen,
		LabelColor:   tcell.ColorDarkGreen,
		ValueColor:   tcell.ColorBlack,
		FooterColor:  tcell.ColorOliveDrab,
		TimerColor:   tcell.ColorGreen,
		ErrorColor:   tcell.ColorDarkRed,
		WarningColor: tcell.ColorDarkGoldenrod,
		InfoColor:    tcell.ColorDarkSlateGray,
		DebugColor:   tcell.ColorGray,
		SuccessColor: tcell.ColorGreen,
		DangerColor:  tcell.ColorCrimson,
		MutedColor:   tcell.ColorDarkGray,
	},
	"light-autumn": {
		Name:         "light-autumn",
		Description:  "Warm autumn browns and oranges (light background)",
		Background:   "light",
		TitleColor:   tcell.ColorSaddleBrown,
		BorderColor:  tcell.ColorSienna,
		LabelColor:   tcell.ColorChocolate,
		ValueColor:   tcell.ColorBlack,
		FooterColor:  tcell.ColorPeru,
		TimerColor:   tcell.ColorDarkOrange,
		ErrorColor:   tcell.ColorMaroon,
		WarningColor: tcell.ColorOrange,
		InfoColor:    tcell.ColorSaddleBrown,
		DebugColor:   tcell.ColorDimGray,
		SuccessColor: tcell.ColorDarkGreen,
		DangerColor:  tcell.ColorDarkRed,
		MutedColor:   tcell.ColorGray,
	},
	"light-lavender": {
		Name:         "light-lavender",
		Description:  "Soft lavender and purple accents (light background)",
		Background:   "light",
		TitleColor:   tcell.ColorDarkOrchid,
		BorderColor:  tcell.ColorMediumPurple,
		LabelColor:   tcell.ColorDarkViolet,
		ValueColor:   tcell.ColorBlack,
		FooterColor:  tcell.ColorPlum,
		TimerColor:   tcell.ColorPurple,
		ErrorColor:   tcell.ColorDarkRed,
		WarningColor: tcell.ColorDarkOrange,
		InfoColor:    tcell.ColorIndigo,
		DebugColor:   tcell.ColorGray,
		SuccessColor: tcell.ColorDarkGreen,
		DangerColor:  tcell.ColorRed,
		MutedColor:   tcell.ColorDarkGray,
	},
	"light-monochrome": {
		Name:         "light-monochrome",
		Description:  "Classic black and white (light background)",
		Background:   "light",
		TitleColor:   tcell.ColorBlack,
		BorderColor:  tcell.ColorDimGray,
		LabelColor:   tcell.ColorDarkSlateGray,
		ValueColor:   tcell.ColorBlack,
		FooterColor:  tcell.ColorGray,
		TimerColor:   tcell.ColorBlack,
		ErrorColor:   tcell.ColorDarkRed,
		WarningColor: tcell.ColorDarkOrange,
		InfoColor:    tcell.ColorBlack,
		DebugColor:   tcell.ColorGray,
		SuccessColor: tcell.ColorDarkGreen,
		DangerColor:  tcell.ColorRed,
		MutedColor:   tcell.ColorDarkGray,
	},
	"light-ocean": {
		Name:         "light-ocean",
		Description:  "Oceanic teals and turquoise (light background)",
		Background:   "light",
		TitleColor:   tcell.ColorDarkCyan,
		BorderColor:  tcell.ColorTeal,
		LabelColor:   tcell.ColorDarkTurquoise,
		ValueColor:   tcell.ColorBlack,
		FooterColor:  tcell.ColorCadetBlue,
		TimerColor:   tcell.ColorDarkCyan,
		ErrorColor:   tcell.ColorDarkRed,
		WarningColor: tcell.ColorDarkOrange,
		InfoColor:    tcell.ColorDarkSlateGray,
		DebugColor:   tcell.ColorGray,
		SuccessColor: tcell.ColorDarkGreen,
		DangerColor:  tcell.ColorCrimson,
		MutedColor:   tcell.ColorDarkGray,
	},
}

// GetTheme returns the theme by name, or default if not found
func GetTheme(name string) *Theme {
	if theme, ok := themes[name]; ok {
		return theme
	}
	return themes["dark-ocean"] // Default theme
}

// GetAllThemeNames returns a list of all theme names in order
func GetAllThemeNames() []string {
	return []string{
		"dark-ocean", "dark-forest", "dark-sunset", "dark-twilight", "dark-matrix", "dark-cyberpunk",
		"light-sky", "light-garden", "light-autumn", "light-lavender", "light-monochrome", "light-ocean",
	}
}

// GetNextTheme returns the next theme in the cycle
func GetNextTheme(currentName string) *Theme {
	names := GetAllThemeNames()
	for i, name := range names {
		if name == currentName {
			nextIndex := (i + 1) % len(names)
			return themes[names[nextIndex]]
		}
	}
	return GetTheme(currentName) // Fallback to current
}

// ListThemes prints all available themes grouped by background type
func ListThemes() {
	fmt.Println("Available Status Console Themes")
	fmt.Println("================================\n")

	fmt.Println("DARK BACKGROUND THEMES:")
	fmt.Println("-----------------------")
	for _, name := range []string{"dark-ocean", "dark-forest", "dark-sunset", "dark-twilight", "dark-matrix", "dark-cyberpunk"} {
		theme := themes[name]
		fmt.Printf("  %-20s %s\n", theme.Name, theme.Description)
	}

	fmt.Println("\nLIGHT BACKGROUND THEMES:")
	fmt.Println("------------------------")
	for _, name := range []string{"light-sky", "light-garden", "light-autumn", "light-lavender", "light-monochrome", "light-ocean"} {
		theme := themes[name]
		fmt.Printf("  %-20s %s\n", theme.Name, theme.Description)
	}

	fmt.Println("\nUSAGE:")
	fmt.Println("  --status-theme <name>        Set theme (default: dark-ocean)")
	fmt.Println("  STATUS_THEME=<name>          Set theme via environment variable")
	fmt.Println("\nEXAMPLES:")
	fmt.Println("  tempest-homekit-go --status --status-theme dark-matrix")
	fmt.Println("  tempest-homekit-go --status --status-theme light-sky")
	fmt.Println("  STATUS_THEME=dark-cyberpunk tempest-homekit-go --status")
}

// GetLogLevelColor returns the appropriate color for a log level based on theme
func (t *Theme) GetLogLevelColor(level string) tcell.Color {
	switch level {
	case "ERROR", "error":
		return t.ErrorColor
	case "WARNING", "warning", "WARN", "warn":
		return t.WarningColor
	case "INFO", "info":
		return t.InfoColor
	case "DEBUG", "debug":
		return t.DebugColor
	default:
		return t.ValueColor
	}
}

// ColorizeLogLine applies theme colors to log messages based on level
func (t *Theme) ColorizeLogLine(line string) string {
	// Simple pattern matching for common log formats
	if len(line) == 0 {
		return line
	}

	// Check for log level indicators
	for _, level := range []string{"ERROR", "WARN", "INFO", "DEBUG"} {
		if contains(line, level+":") || contains(line, "["+level+"]") {
			color := t.GetLogLevelColor(level)
			return fmt.Sprintf("[%s]%s[-]", colorToTviewTag(color), line)
		}
	}

	// Check for alarm/alert indicators
	if contains(line, "ALARM") || contains(line, "🚨") {
		return fmt.Sprintf("[%s]%s[-]", colorToTviewTag(t.DangerColor), line)
	}

	return line
}

// colorToTviewTag converts tcell.Color to tview color tag
func colorToTviewTag(c tcell.Color) string {
	// Map common tcell colors to tview color names
	colorMap := map[tcell.Color]string{
		tcell.ColorRed:         "red",
		tcell.ColorGreen:       "green",
		tcell.ColorYellow:      "yellow",
		tcell.ColorBlue:        "blue",
		tcell.ColorDarkMagenta: "darkmagenta",
		tcell.ColorAqua:        "aqua",
		tcell.ColorWhite:       "white",
		tcell.ColorBlack:       "black",
		tcell.ColorGray:        "gray",
		tcell.ColorOrange:      "orange",
		tcell.ColorPurple:      "purple",
		tcell.ColorLime:        "lime",
		tcell.ColorFuchsia:     "fuchsia",
		tcell.ColorNavy:        "navy",
		tcell.ColorTeal:        "teal",
		tcell.ColorOlive:       "olive",
		tcell.ColorMaroon:      "maroon",
		tcell.ColorLightGreen:  "lightgreen",
		tcell.ColorDarkGreen:   "darkgreen",
		tcell.ColorOrangeRed:   "orangered",
		tcell.ColorDarkRed:     "darkred",
		tcell.ColorGold:        "gold",
		tcell.ColorLightBlue:   "lightblue",
		tcell.ColorDarkBlue:    "darkblue",
		tcell.ColorLightYellow: "lightyellow",
		tcell.ColorDarkGray:    "darkgray",
		tcell.ColorDimGray:     "dimgray",
	}

	if name, ok := colorMap[c]; ok {
		return name
	}

	// Fallback to color code
	return fmt.Sprintf("#%06x", c.Hex())
}

// contains checks if a string contains a substring (case-insensitive helper for internal use)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
