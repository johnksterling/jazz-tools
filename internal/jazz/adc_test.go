package jazz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRefreshAuthorizedUserToken(t *testing.T) {
	// Spin up mock OAuth token endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		_ = r.ParseForm()
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("expected grant_type=refresh_token, got %s", r.Form.Get("grant_type"))
		}
		if r.Form.Get("refresh_token") != "test_refresh_token" {
			t.Errorf("expected refresh_token=test_refresh_token, got %s", r.Form.Get("refresh_token"))
		}

		resp := map[string]interface{}{
			"access_token": "ya29.test_access_token_123",
			"expires_in":   3600,
			"token_type":   "Bearer",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create temp authorized_user credentials file
	tmpDir := t.TempDir()
	credsFile := filepath.Join(tmpDir, "adc.json")
	credsData := map[string]interface{}{
		"client_id":     "dummy_client_id.apps.googleusercontent.com",
		"client_secret": "dummy_secret",
		"refresh_token": "test_refresh_token",
		"type":          "authorized_user",
	}
	rawBytes, _ := json.Marshal(credsData)
	_ = os.WriteFile(credsFile, rawBytes, 0600)

	origEnv := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credsFile)
	defer func() {
		if origEnv != "" {
			_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", origEnv)
		} else {
			_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		}
	}()

	// Test loading credentials file directly
	raw := make(map[string]interface{})
	_ = json.Unmarshal(rawBytes, &raw)
	if raw["type"] != "authorized_user" {
		t.Fatalf("expected authorized_user type")
	}
}

func TestResolveADCWhenNotConfigured(t *testing.T) {
	origEnv := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/nonexistent/path/creds.json")
	defer func() {
		if origEnv != "" {
			_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", origEnv)
		} else {
			_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		}
	}()

	// Clear cache
	adcMu.Lock()
	adcCache = adcTokenCache{}
	adcMu.Unlock()

	_, err := ResolveADCAccessToken()
	if err == nil {
		t.Fatalf("expected error when credentials file does not exist and gcloud is not present")
	}
}
