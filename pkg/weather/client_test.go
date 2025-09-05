package weather

import (
	"testing"
)

func TestFindStationByName(t *testing.T) {
	stations := []Station{
		{StationID: 1, Name: "station1", StationName: "Station 1"},
		{StationID: 2, Name: "tempest-homekit", StationName: "Tempest HomeKit"},
	}

	station := FindStationByName(stations, "tempest-homekit")
	if station == nil {
		t.Fatal("Station not found")
	}
	if station.StationID != 2 {
		t.Errorf("Expected ID 2, got %d", station.StationID)
	}
}

func TestFindStationByNameNotFound(t *testing.T) {
	stations := []Station{
		{StationID: 1, Name: "station1"},
	}

	station := FindStationByName(stations, "notfound")
	if station != nil {
		t.Error("Expected nil, got station")
	}
}
