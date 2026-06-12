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

    "git.ierislabs.dev/update-link/data"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type TagFetcher interface {
    FetchTags(image data.ImageInfo) ([]string, error)
}

func selectLatestTag(tags []string, currentExtra string) (data.SemanticVersion, error) {
    var candidates []data.SemanticVersion
    for _, tag := range tags {
        if currentExtra == "" {
            if strings.Contains(tag, "-") {
                continue
            }
        } else if !strings.Contains(tag, "-"+currentExtra) {
            continue
        }
        v, err := data.ParseSemanticVersion(tag)
        if err != nil || v.Extra != currentExtra {
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
        if b.GreaterThan(a) {
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
        return "", err
    }
    req.Header.Set("Authorization", "Bearer "+token)
    resp, err := httpClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("failed to fetch URL %s: status code %d", targetURL, resp.StatusCode)
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}

func fetch(targetURL string) (string, error) {
    resp, err := httpClient.Get(targetURL)
    if err != nil {
        return "", err
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
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}
