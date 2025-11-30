package e2e

import (
    "net/http"
    "testing"
)

func TestHealthEndpoint(t *testing.T) {
    resp, err := http.Get("http://localhost:8080/api/tasks")
    if err != nil {
        t.Fatalf("failed to reach backend: %v", err)
    }

    if resp.StatusCode != 200 {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
}
