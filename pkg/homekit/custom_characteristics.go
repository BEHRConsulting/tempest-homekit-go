package homekit

import (
	"github.com/brutella/hap/characteristic"
)

// Custom characteristics that don't use Temperature type to avoid C->F conversion

// WindSpeedCharacteristic - Custom characteristic for wind speed in mph
type WindSpeedCharacteristic struct {
	*characteristic.Float
}

func NewWindSpeedCharacteristic() *WindSpeedCharacteristic {
	c := characteristic.NewFloat("F001-0001-1000-8000-0026BB765291") // Custom UUID
	c.Format = characteristic.FormatFloat
	c.Unit = "mph"
	c.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
	c.SetMinValue(0.0)
	c.SetMaxValue(200.0)
	c.SetStepValue(0.1)
	c.SetValue(0.0)

	return &WindSpeedCharacteristic{c}
}

// WindDirectionCharacteristic - Custom characteristic for wind direction in degrees
type WindDirectionCharacteristic struct {
	*characteristic.Float
}

func NewWindDirectionCharacteristic() *WindDirectionCharacteristic {
	c := characteristic.NewFloat("F002-0001-1000-8000-0026BB765291") // Custom UUID
	c.Format = characteristic.FormatFloat
	c.Unit = "degrees"
	c.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
	c.SetMinValue(0.0)
	c.SetMaxValue(360.0)
	c.SetStepValue(1.0)
	c.SetValue(0.0)

	return &WindDirectionCharacteristic{c}
}

// RainAccumulationCharacteristic - Custom characteristic for rain in inches
type RainAccumulationCharacteristic struct {
	*characteristic.Float
}

func NewRainAccumulationCharacteristic() *RainAccumulationCharacteristic {
	c := characteristic.NewFloat("F003-0001-1000-8000-0026BB765291") // Custom UUID
	c.Format = characteristic.FormatFloat
	c.Unit = "inches"
	c.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
	c.SetMinValue(0.0)
	c.SetMaxValue(100.0)
	c.SetStepValue(0.001)
	c.SetValue(0.0)

	return &RainAccumulationCharacteristic{c}
}

// UVIndexCharacteristic - Custom characteristic for UV Index
type UVIndexCharacteristic struct {
	*characteristic.Float
}

func NewUVIndexCharacteristic() *UVIndexCharacteristic {
	c := characteristic.NewFloat("00000001-0000-1000-8000-0026BB765291") // HomeKit-compliant custom UUID for UV Index
	c.Format = characteristic.FormatFloat
	c.Unit = "UV Index" // Set unit to UV Index
	c.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
	c.SetMinValue(0.0)
	c.SetMaxValue(15.0)
	c.SetStepValue(0.1)
	c.SetValue(0.0)

	return &UVIndexCharacteristic{c}
}

// LightningCountCharacteristic - Custom characteristic for lightning strike count
type LightningCountCharacteristic struct {
	*characteristic.Int
}

func NewLightningCountCharacteristic() *LightningCountCharacteristic {
	c := characteristic.NewInt("F005-0001-1000-8000-0026BB765291") // Custom UUID
	c.Format = characteristic.FormatInt32
	c.Unit = "strikes"
	c.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
	c.SetMinValue(0)
	c.SetMaxValue(10000)
	c.SetStepValue(1)
	c.SetValue(0)

	return &LightningCountCharacteristic{c}
}

// LightningDistanceCharacteristic - Custom characteristic for lightning distance in km
type LightningDistanceCharacteristic struct {
	*characteristic.Float
}

func NewLightningDistanceCharacteristic() *LightningDistanceCharacteristic {
	c := characteristic.NewFloat("F006-0001-1000-8000-0026BB765291") // Custom UUID
	c.Format = characteristic.FormatFloat
	c.Unit = "km"
	c.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
	c.SetMinValue(0.0)
	c.SetMaxValue(100.0)
	c.SetStepValue(0.1)
	c.SetValue(0.0)

	return &LightningDistanceCharacteristic{c}
}

// PrecipitationTypeCharacteristic - Custom characteristic for precipitation type
type PrecipitationTypeCharacteristic struct {
	*characteristic.Int
}

func NewPrecipitationTypeCharacteristic() *PrecipitationTypeCharacteristic {
	c := characteristic.NewInt("F007-0001-1000-8000-0026BB765291") // Custom UUID
	c.Format = characteristic.FormatInt32
	c.Unit = "type"
	c.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
	c.SetMinValue(0)
	c.SetMaxValue(10)
	c.SetStepValue(1)
	c.SetValue(0)

	return &PrecipitationTypeCharacteristic{c}
}

// PressureCharacteristic - Custom characteristic for atmospheric pressure in mb
type PressureCharacteristic struct {
	*characteristic.Float
}

func NewPressureCharacteristic() *PressureCharacteristic {
	c := characteristic.NewFloat("F008-0001-1000-8000-0026BB765291") // Custom UUID
	c.Format = characteristic.FormatFloat
	c.Unit = "mb"
	c.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
	c.SetMinValue(800.0)
	c.SetMaxValue(1100.0)
	c.SetStepValue(0.1)
	c.SetValue(1013.25) // Standard atmospheric pressure

	return &PressureCharacteristic{c}
}
