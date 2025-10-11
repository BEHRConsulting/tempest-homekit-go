//go:build !darwin
// +build !darwin

package alarm

import (
	"fmt"

	"tempest-homekit-go/pkg/weather"
)

// OSLogNotifier is a placeholder for non-macOS platforms
type OSLogNotifier struct{}

func (n *OSLogNotifier) Send(alarm *Alarm, channel *Channel, obs *weather.Observation, stationName string) error {
	return fmt.Errorf("oslog notification type is only supported on macOS")
}
