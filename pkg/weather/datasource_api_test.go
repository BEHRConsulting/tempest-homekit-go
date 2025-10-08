package weather

import "testing"

func TestAPIDataSource_GetType(t *testing.T) {
	a1 := NewAPIDataSource(0, "", "", APIDataSourceOptions{CustomURL: "http://localhost:8080/api/generate-weather", GeneratedPath: "/api/generate-weather"})
	if a1.GetType() != DataSourceGenerated {
		t.Fatalf("expected Generated, got %s", a1.GetType())
	}

	a2 := NewAPIDataSource(0, "", "", APIDataSourceOptions{CustomURL: "http://localhost:8080/custom-endpoint", GeneratedPath: "/api/generate-weather"})
	if a2.GetType() != DataSourceCustomURL {
		t.Fatalf("expected CustomURL, got %s", a2.GetType())
	}

	a3 := NewAPIDataSource(0, "", "", APIDataSourceOptions{CustomURL: "", GeneratedPath: "/api/generate-weather"})
	if a3.GetType() != DataSourceAPI {
		t.Fatalf("expected API, got %s", a3.GetType())
	}
}
