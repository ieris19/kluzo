package container

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"git.ierislabs.dev/ieris19/kluzo/internal/data"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type TagFetcher interface {
	FetchTags(image data.ImageInfo) ([]string, error)
}

func selectLatestTag(tags []string, definition data.ContainerDefinition) (data.SemanticVersion, error) {
	current := definition.Version
	tagPattern := definition.TagPattern
	var candidates []data.SemanticVersion
	for _, tag := range tags {
		if tagPattern == nil {
			if current.Extra == "" {
				if strings.Contains(tag, "-") {
					continue
				}
			} else if !strings.Contains(tag, "-"+current.Extra) {
				continue
			}
		}
		v, err := data.ParseSemanticVersion(tag, tagPattern)
		if err != nil || v.Extra != current.Extra {
			continue
		}
		if definition.Pin == data.PinMajor && v.Major != current.Major {
			continue
		}
		if definition.Pin == data.PinMinor && (v.Major != current.Major || v.Minor != current.Minor) {
			continue
		}
		candidates = append(candidates, v)
	}
	if len(candidates) == 0 {
		return data.SemanticVersion{}, fmt.Errorf("no matching tags found")
	}
	slices.SortFunc(candidates, func(a, b data.SemanticVersion) int {
		if a.GreaterThan(b) {
			return -1
		}
		if a.LessThan(b) {
			return 1
		}
		return 0
	})
	return candidates[0], nil
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

func fetchAuthorization(targetURL string, challenge bearerChallenge) (string, error) {
	token, err := fetchToken(challenge)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate for %s: %v", targetURL, err)
	}
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request for %s: %v", targetURL, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL %s: %v", targetURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL %s: status code %d", targetURL, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from %s: %v", targetURL, err)
	}
	return string(body), nil
}

func fetch(targetURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request for %s: %v", targetURL, err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL %s: %v", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		challenge, ok := parseBearerChallenge(resp.Header.Get("Www-Authenticate"))
		if !ok {
			return "", fmt.Errorf("failed to fetch URL %s: status code 401", targetURL)
		}
		return fetchAuthorization(targetURL, challenge)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL %s: status code %d", targetURL, resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); strings.Contains(ct, "text/html") {
		return "", fmt.Errorf("failed to fetch URL %s: server returned HTML instead of JSON", targetURL)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from %s: %v", targetURL, err)
	}
	return string(body), nil
}
