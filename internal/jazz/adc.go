package jazz

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type adcTokenCache struct {
	token     string
	expiresAt time.Time
}

var (
	adcMu    sync.Mutex
	adcCache adcTokenCache
)

// ResolveADCAccessToken attempts to obtain an OAuth2 access token via Google Application
// Default Credentials (ADC), searching in order:
// 1. $GOOGLE_APPLICATION_CREDENTIALS file path
// 2. ~/.config/gcloud/application_default_credentials.json (or %APPDATA%\gcloud\...)
// 3. 'gcloud auth print-access-token' CLI execution
func ResolveADCAccessToken() (string, error) {
	adcMu.Lock()
	defer adcMu.Unlock()

	// Return cached token if valid for at least 2 more minutes
	if adcCache.token != "" && time.Now().Add(2*time.Minute).Before(adcCache.expiresAt) {
		return adcCache.token, nil
	}

	// 1. Check GOOGLE_APPLICATION_CREDENTIALS
	if envPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); envPath != "" {
		token, expiry, err := loadTokenFromCredentialsFile(envPath)
		if err == nil {
			adcCache = adcTokenCache{token: token, expiresAt: expiry}
			return token, nil
		}
	}

	// 2. Check standard gcloud user credentials path
	var userCredsPath string
	if home, err := os.UserHomeDir(); err == nil {
		userCredsPath = filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")
		if _, err := os.Stat(userCredsPath); err == nil {
			token, expiry, err := loadTokenFromCredentialsFile(userCredsPath)
			if err == nil {
				adcCache = adcTokenCache{token: token, expiresAt: expiry}
				return token, nil
			}
		}
	}

	// 3. Check 'gcloud auth print-access-token'
	token, err := runGcloudPrintAccessToken()
	if err == nil && token != "" {
		adcCache = adcTokenCache{token: token, expiresAt: time.Now().Add(50 * time.Minute)}
		return token, nil
	}

	return "", fmt.Errorf("no Google Application Default Credentials found (run 'gcloud auth application-default login' or set $GOOGLE_APPLICATION_CREDENTIALS)")
}

func loadTokenFromCredentialsFile(path string) (string, time.Time, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", time.Time{}, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", time.Time{}, err
	}

	credType, _ := raw["type"].(string)
	switch credType {
	case "authorized_user":
		return refreshAuthorizedUserToken(raw)
	case "service_account":
		return exchangeServiceAccountJWT(raw)
	default:
		return "", time.Time{}, fmt.Errorf("unsupported credentials type: %s", credType)
	}
}

func refreshAuthorizedUserToken(creds map[string]interface{}) (string, time.Time, error) {
	clientID, _ := creds["client_id"].(string)
	clientSecret, _ := creds["client_secret"].(string)
	refreshToken, _ := creds["refresh_token"].(string)

	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return "", time.Time{}, fmt.Errorf("incomplete authorized_user credentials")
	}

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("token refresh error (status %d): %s", resp.StatusCode, string(body))
	}

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", time.Time{}, err
	}

	expiry := time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
	return res.AccessToken, expiry, nil
}

func exchangeServiceAccountJWT(creds map[string]interface{}) (string, time.Time, error) {
	clientEmail, _ := creds["client_email"].(string)
	privateKeyPEM, _ := creds["private_key"].(string)
	tokenURI, _ := creds["token_uri"].(string)
	if tokenURI == "" {
		tokenURI = "https://oauth2.googleapis.com/token"
	}

	if clientEmail == "" || privateKeyPEM == "" {
		return "", time.Time{}, fmt.Errorf("missing service account client_email or private_key")
	}

	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", time.Time{}, fmt.Errorf("failed to parse private key PEM")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Fallback to PKCS1
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	rsaKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return "", time.Time{}, fmt.Errorf("private key is not RSA")
	}

	now := time.Now().Unix()
	headerJSON, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	claimsJSON, _ := json.Marshal(map[string]interface{}{
		"iss":   clientEmail,
		"scope": "https://www.googleapis.com/auth/generative-language https://www.googleapis.com/auth/cloud-platform",
		"aud":   tokenURI,
		"exp":   now + 3600,
		"iat":   now,
	})

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := headerB64 + "." + claimsB64

	hashed := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign JWT: %w", err)
	}

	jwtStr := signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", jwtStr)

	resp, err := http.PostForm(tokenURI, form)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("service account token exchange error (status %d): %s", resp.StatusCode, string(body))
	}

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", time.Time{}, err
	}

	expiry := time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
	return res.AccessToken, expiry, nil
}

func runGcloudPrintAccessToken() (string, error) {
	cmd := exec.Command("gcloud", "auth", "print-access-token")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(out))
	return token, nil
}
