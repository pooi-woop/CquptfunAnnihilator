package scraper

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type ScrapedEndpoint struct {
	Path       string
	Source     string
	Confidence int
}

func ScrapeDashboard(baseURL, cookie string) ([]ScrapedEndpoint, error) {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	resp, err := client.R().Get(baseURL + "/student/dashboard")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dashboard: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("dashboard returned status %d", resp.StatusCode())
	}

	htmlContent := string(resp.Body())
	return extractEndpoints(htmlContent), nil
}

func extractEndpoints(htmlContent string) []ScrapedEndpoint {
	var endpoints []ScrapedEndpoint

	patterns := []struct {
		re         *regexp.Regexp
		source     string
		confidence int
	}{
		{regexp.MustCompile(`["']/api/[^"']+["']`), "Direct API path", 100},
		{regexp.MustCompile(`["']/student/api/[^"']+["']`), "Student API path", 100},
		{regexp.MustCompile(`["']/v1/[^"']+["']`), "v1 API path", 80},
		{regexp.MustCompile(`["']/v2/[^"']+["']`), "v2 API path", 80},
		{regexp.MustCompile(`["']/problems[^"']*["']`), "Problems path", 90},
		{regexp.MustCompile(`["']/questions[^"']*["']`), "Questions path", 90},
		{regexp.MustCompile(`["']/exercises[^"']*["']`), "Exercises path", 90},
		{regexp.MustCompile(`["']/assignments[^"']*["']`), "Assignments path", 90},
		{regexp.MustCompile(`["']/submissions[^"']*["']`), "Submissions path", 90},
		{regexp.MustCompile(`["']/quiz[^"']*["']`), "Quiz path", 80},
		{regexp.MustCompile(`["']/homework[^"']*["']`), "Homework path", 80},
	}

	foundPaths := make(map[string]bool)

	for _, pattern := range patterns {
		matches := pattern.re.FindAllString(htmlContent, -1)
		for _, match := range matches {
			path := strings.Trim(match, `"'`)
			if !foundPaths[path] && len(path) > 3 && !strings.Contains(path, " ") {
				foundPaths[path] = true
				endpoints = append(endpoints, ScrapedEndpoint{
					Path:       path,
					Source:     pattern.source,
					Confidence: pattern.confidence,
				})
			}
		}
	}

	return endpoints
}

func FetchAndParseScript(baseURL, cookie, scriptURL string) ([]ScrapedEndpoint, error) {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	resp, err := client.R().Get(scriptURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch script %s: %w", scriptURL, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("script returned status %d", resp.StatusCode())
	}

	scriptContent := string(resp.Body())
	return extractEndpoints(scriptContent), nil
}

func GetScriptURLs(htmlContent string) []string {
	re := regexp.MustCompile(`<script[^>]+src=["']([^"']+)["']`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var urls []string
	for _, match := range matches {
		if len(match) == 2 {
			urls = append(urls, match[1])
		}
	}

	return urls
}

func ComprehensiveScrape(baseURL, cookie string) ([]ScrapedEndpoint, error) {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	resp, err := client.R().Get(baseURL + "/student/dashboard")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dashboard: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("dashboard returned status %d", resp.StatusCode())
	}

	htmlContent := string(resp.Body())

	var allEndpoints []ScrapedEndpoint
	foundPaths := make(map[string]bool)

	pageEndpoints := extractEndpoints(htmlContent)
	for _, ep := range pageEndpoints {
		if !foundPaths[ep.Path] {
			foundPaths[ep.Path] = true
			allEndpoints = append(allEndpoints, ep)
		}
	}

	scriptURLs := GetScriptURLs(htmlContent)
	fmt.Printf("\nFound %d script URLs to analyze\n", len(scriptURLs))

	for _, scriptURL := range scriptURLs {
		if !strings.HasPrefix(scriptURL, "http") {
			scriptURL = baseURL + scriptURL
		}

		fmt.Printf("Analyzing script: %s\n", scriptURL)
		scriptEndpoints, err := FetchAndParseScript(baseURL, cookie, scriptURL)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		for _, ep := range scriptEndpoints {
			if !foundPaths[ep.Path] {
				foundPaths[ep.Path] = true
				allEndpoints = append(allEndpoints, ep)
			}
		}
	}

	return allEndpoints, nil
}

func ProbeScrapedEndpoints(baseURL, cookie string, endpoints []ScrapedEndpoint) []ScrapedEndpoint {
	client := resty.New().
		SetTimeout(10 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	var validEndpoints []ScrapedEndpoint

	fmt.Println("\n=== Probing scraped endpoints ===")
	for _, ep := range endpoints {
		url := baseURL + ep.Path
		resp, err := client.R().Get(url)

		if err != nil {
			continue
		}

		status := resp.StatusCode()
		contentType := resp.Header().Get("Content-Type")

		fmt.Printf("%-40s -> Status: %d, Content-Type: %s\n", ep.Path, status, contentType)

		if status == http.StatusOK || status == http.StatusUnauthorized || status == http.StatusForbidden {
			validEndpoints = append(validEndpoints, ep)
		}
	}

	return validEndpoints
}
