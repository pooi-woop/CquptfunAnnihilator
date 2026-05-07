package analyzer

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

func AnalyzeHomepage(baseURL, cookie string) error {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	resp, err := client.R().Get(baseURL)
	if err != nil {
		return fmt.Errorf("failed to fetch homepage: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("homepage returned status %d", resp.StatusCode())
	}

	html := string(resp.Body())

	fmt.Println("=== Homepage Analysis ===")
	fmt.Printf("Content length: %d bytes\n", len(html))

	// 查找可能的API主机
	fmt.Println("\n--- Possible API Hosts ---")
	apiHostPatterns := []string{
		`api\.[a-zA-Z0-9]+\.[a-zA-Z]+`,
		`["']https?://[^"']+api[^"']*["']`,
		`["']//[^"']+api[^"']*["']`,
	}

	for _, pattern := range apiHostPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(html, -10)
		for _, match := range matches {
			fmt.Printf("  %s\n", strings.Trim(match, `"'`))
		}
	}

	// 查找可能的basePath
	fmt.Println("\n--- Possible Base Paths ---")
	basePathPatterns := []string{
		`baseURL\s*[=:]\s*["']([^"']+)["']`,
		`basePath\s*[=:]\s*["']([^"']+)["']`,
		`apiUrl\s*[=:]\s*["']([^"']+)["']`,
		`apiEndpoint\s*[=:]\s*["']([^"']+)["']`,
	}

	for _, pattern := range basePathPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(html, -10)
		for _, match := range matches {
			if len(match) == 2 {
				fmt.Printf("  %s\n", match[1])
			}
		}
	}

	// 查找配置变量
	fmt.Println("\n--- Environment Variables ---")
	envPattern := regexp.MustCompile(`window\._env_["']?\s*=\s*({[^}]+})`)
	matches := envPattern.FindAllStringSubmatch(html, -5)
	for _, match := range matches {
		if len(match) == 2 {
			fmt.Printf("  Found env config: %s\n", truncate(match[1], 100))
		}
	}

	// 查找script中的fetch调用
	fmt.Println("\n--- Fetch/Axios Patterns ---")
	fetchPatterns := []string{
		`fetch\s*\(\s*["']([^"']+)["']`,
		`axios\.get\s*\(\s*["']([^"']+)["']`,
		`axios\.post\s*\(\s*["']([^"']+)["']`,
	}

	for _, pattern := range fetchPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(html, -10)
		for _, match := range matches {
			if len(match) == 2 {
				fmt.Printf("  %s\n", match[1])
			}
		}
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func AnalyzeDashboard(baseURL, cookie string) error {
	return ExtractEmbeddedJSON(baseURL, cookie)
}

func TestSubdomains(baseURL, cookie string) {
	client := resty.New().
		SetTimeout(10 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	parts := strings.Split(strings.TrimPrefix(baseURL, "https://"), ".")
	if len(parts) < 2 {
		fmt.Println("Invalid base URL")
		return
	}

	domain := strings.Join(parts[len(parts)-2:], ".")

	subdomains := []string{
		"api",
		"backend",
		"api.v1",
		"api.v2",
		"student.api",
		"www",
		"app",
	}

	fmt.Println("\n=== Testing Subdomains ===")
	for _, sub := range subdomains {
		testURL := fmt.Sprintf("https://%s.%s", sub, domain)
		resp, err := client.R().Get(testURL + "/")
		if err != nil {
			fmt.Printf("  %-20s -> Error: %v\n", testURL, err)
			continue
		}
		fmt.Printf("  %-20s -> Status: %d\n", testURL, resp.StatusCode())
	}
}
