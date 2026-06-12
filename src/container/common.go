package container

import (
    "fmt"
    "io"
    "net/http"
    "slices"
    "strings"
    "time"

    "git.ierislabs.dev/update-link/data"
)

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

func fetch(url string) (string, error) {
    var httpClient = &http.Client{Timeout: 10 * time.Second}
    resp, err := httpClient.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("failed to fetch URL %s: status code %d", url, resp.StatusCode)
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}
