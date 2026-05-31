package container

import (
    "fmt"
    "io"
    "net/http"
    "time"
)

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
