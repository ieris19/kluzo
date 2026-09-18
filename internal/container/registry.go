package container

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"git.ierislabs.dev/ieris19/kluzo/internal/data"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type TagFetcher interface {
	FetchTags(image data.ImageInfo) ([]string, error)
}

var bearerParamRegexp = regexp.MustCompile(`(\w+)="([^"]*)"`)

type bearerChallenge struct {
	Realm   string
	Service string
	Scope   string
}

func parseBearerChallenge(header string) (bearerChallenge, bool) {
	if !strings.HasPrefix(header, "Bearer ") {
		return bearerChallenge{}, false
	}
	var c bearerChallenge
	for _, match := range bearerParamRegexp.FindAllStringSubmatch(header, -1) {
		switch match[1] {
		case "realm":
			c.Realm = match[2]
		case "service":
			c.Service = match[2]
		case "scope":
			c.Scope = match[2]
		}
	}
	if c.Realm == "" {
		return bearerChallenge{}, false
	}
	return c, true
}

type tokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
}

func fetchToken(challenge bearerChallenge) (string, error) {
	params := url.Values{}
	if challenge.Service != "" {
		params.Set("service", challenge.Service)
	}
	if challenge.Scope != "" {
		params.Set("scope", challenge.Scope)
	}
	tokenURL := challenge.Realm
	if len(params) > 0 {
		tokenURL += "?" + params.Encode()
	}
	resp, err := httpClient.Get(tokenURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch auth token: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch auth token: status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read auth token: %v", err)
	}
	var tr tokenResponse
	if err = json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("failed to parse auth token: %v", err)
	}
	if tr.Token != "" {
		return tr.Token, nil
	}
	if tr.AccessToken != "" {
		return tr.AccessToken, nil
	}
	return "", fmt.Errorf("no token in auth response")
}

// registrySession carries the bearer token across requests. Tokens are scoped
// to a single repository, so a session covers one image's requests: the pages
// of a tag listing share one token instead of re-running the 401 challenge.
type registrySession struct {
	token string
}

func (s *registrySession) fetchAuthorization(targetURL string, challenge bearerChallenge) (string, http.Header, error) {
	token, err := fetchToken(challenge)
	if err != nil {
		return "", nil, fmt.Errorf("failed to authenticate for %s: %v", targetURL, err)
	}
	s.token = token
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build request for %s: %v", targetURL, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch URL %s: %v", targetURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to fetch URL %s: status code %d", targetURL, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body from %s: %v", targetURL, err)
	}
	return string(body), resp.Header, nil
}

// Fetch the response body along with its headers
func (s *registrySession) fetch(targetURL string) (string, http.Header, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build request for %s: %v", targetURL, err)
	}
	req.Header.Set("Accept", "application/json")
	// Reuse the token from an earlier request in this session, if any
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch URL %s: %v", targetURL, err)
	}
	defer resp.Body.Close()

	// A cached token that has expired lands here too, and is replaced
	if resp.StatusCode == http.StatusUnauthorized {
		challenge, ok := parseBearerChallenge(resp.Header.Get("Www-Authenticate"))
		if !ok {
			return "", nil, fmt.Errorf("failed to fetch URL %s: status code 401", targetURL)
		}
		return s.fetchAuthorization(targetURL, challenge)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to fetch URL %s: status code %d", targetURL, resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); strings.Contains(ct, "text/html") {
		return "", nil, fmt.Errorf("failed to fetch URL %s: server returned HTML instead of JSON", targetURL)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body from %s: %v", targetURL, err)
	}
	return string(body), resp.Header, nil
}
