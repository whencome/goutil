package xhttp

import (
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestClient(t *testing.T) {
    // Create a new test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check the request method and URL
        if r.Method != http.MethodGet {
            t.Errorf("expected GET request, got %s", r.Method)
        }
        if r.URL.String() != "/test" {
            t.Errorf("expected URL /test, got %s", r.URL.String())
        }
        // Write a response
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("test response"))
    }))
    defer server.Close()

    // Create a new client
    client := NewClient()
    client.WithHeader("X-Test", "test")

    // Make a GET request to the test server
    resp, err := client.Get(server.URL + "/test")
    if err != nil {
        t.Errorf("unexpected error: %s", err)
    }
    defer resp.Body.Close()

    // Check the response status code and body
    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("unexpected error: %s", err)
    }
    if string(body) != "test response" {
        t.Errorf("expected body %q, got %q", "test response", string(body))
    }
}
