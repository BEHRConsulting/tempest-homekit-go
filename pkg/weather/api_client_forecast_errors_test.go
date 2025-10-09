package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetForecast_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	// reuse override helper from other tests
	restore := overrideTransportToTestServer(srv)
	defer restore()

	_, err := GetForecast(1, "token")
	if err == nil {
		t.Fatalf("expected error when forecast endpoint returns non-200")
	}
}

func TestGetForecast_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid-json"))
	}))
	defer srv.Close()

	restore := overrideTransportToTestServer(srv)
	defer restore()

	_, err := GetForecast(1, "token")
	if err == nil {
		t.Fatalf("expected JSON parse error for malformed forecast response")
	}
}

func TestGetObservationFromURL_MalformedJSON_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-a-json"))
	}))
	defer srv.Close()

	restore := overrideTransportToTestServer(srv)
	defer restore()

	_, err := GetObservationFromURL(srv.URL)
	if err == nil {
		t.Fatalf("expected error when observation URL returns malformed JSON")
	}
}

func TestGetTempestDeviceID_NoTempest(t *testing.T) {
	s := &Station{StationID: 1, Devices: []Device{{DeviceID: 10, DeviceType: "HB"}}}
	if _, err := GetTempestDeviceID(s); err == nil {
		t.Fatalf("expected error when no Tempest device present")
	}
}
