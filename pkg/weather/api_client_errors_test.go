package weather

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// small transport helper to route any request to the test server
type rt func(*http.Request) (*http.Response, error)

func (f rt) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func overrideTransportToTestServer(srv *httptest.Server) func() {
	old := http.DefaultTransport
	target, _ := url.Parse(srv.URL)

	http.DefaultTransport = rt(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		return old.RoundTrip(req)
	})

	return func() { http.DefaultTransport = old }
}

func TestGetStations_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	restore := overrideTransportToTestServer(srv)
	defer restore()

	_, err := GetStations("token")
	if err == nil {
		t.Fatalf("expected error when stations endpoint returns non-200")
	}
}

func TestGetHistoricalObservationsWithProgress_StationDetailsError(t *testing.T) {
	// Server returns 500 for station details
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	restore := overrideTransportToTestServer(srv)
	defer restore()

	_, err := GetHistoricalObservationsWithProgress(9999, "token", "", nil, 100)
	if err == nil {
		t.Fatalf("expected error when station details cannot be fetched")
	}
}

func TestParseDeviceObservations_SkipsIncomplete(t *testing.T) {
	// Provide an incomplete observation (less than 18 elements)
	incomplete := [][]interface{}{
		{float64(1234567890), 1.0, 2.0},
	}

	out := parseDeviceObservations(incomplete)
	if len(out) != 0 {
		t.Fatalf("expected incomplete observations to be skipped, got %d entries", len(out))
	}
}
