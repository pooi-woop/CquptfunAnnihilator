package analyzer

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"time"

	"github.com/go-resty/resty/v2"
)

func ExtractEmbeddedJSON(baseURL, cookie string) error {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	resp, err := client.R().Get(baseURL + "/student/dashboard")
	if err != nil {
		return fmt.Errorf("failed to fetch dashboard: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("dashboard returned status %d", resp.StatusCode())
	}

	html := string(resp.Body())

	fmt.Println("=== Extracting Embedded JSON ===")

	// 查找window.__INITIAL_STATE__
	initStatePattern := regexp.MustCompile(`window\.__INITIAL_STATE__\s*=\s*({[\s\S]*?});`)
	matches := initStatePattern.FindAllStringSubmatch(html, -1)
	if len(matches) > 0 {
		fmt.Printf("\n--- Found __INITIAL_STATE__ (first 500 chars) ---\n")
		fmt.Println(truncate(matches[0][1], 500))
	}

	// 查找window.data
	dataPattern := regexp.MustCompile(`window\.data\s*=\s*({[\s\S]*?});`)
	matches = dataPattern.FindAllStringSubmatch(html, -1)
	if len(matches) > 0 {
		fmt.Printf("\n--- Found window.data (first 500 chars) ---\n")
		fmt.Println(truncate(matches[0][1], 500))
	}

	// 查找<script id="__NEXT_DATA__"
	nextDataPattern := regexp.MustCompile(`<script[^>]*id="__NEXT_DATA__"[^>]*>({[\s\S]*?})</script>`)
	matches = nextDataPattern.FindAllStringSubmatch(html, -1)
	if len(matches) > 0 {
		fmt.Printf("\n--- Found NEXT_DATA (first 500 chars) ---\n")
		fmt.Println(truncate(matches[0][1], 500))
	}

	// 查找JSON-LD
	jsonLdPattern := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>({[\s\S]*?})</script>`)
	matches = jsonLdPattern.FindAllStringSubmatch(html, -1)
	if len(matches) > 0 {
		fmt.Printf("\n--- Found JSON-LD (first 500 chars) ---\n")
		fmt.Println(truncate(matches[0][1], 500))
	}

	// 查找问题列表模式
	problemsPattern := regexp.MustCompile(`"problems?"\s*:\s*(\[[^\]]+\])`)
	matches = problemsPattern.FindAllStringSubmatch(html, -1)
	if len(matches) > 0 {
		fmt.Printf("\n--- Found problems array (first 500 chars) ---\n")
		fmt.Println(truncate(matches[0][1], 500))
	}

	// 查找问题对象模式
	questionPattern := regexp.MustCompile(`({[^}]*"question"[^}]*})`)
	matches = questionPattern.FindAllStringSubmatch(html, -5)
	if len(matches) > 0 {
		fmt.Printf("\n--- Found question objects ---\n")
		for i, match := range matches {
			fmt.Printf("  [%d] %s\n", i+1, truncate(match[1], 100))
		}
	}

	// 查找id和title模式
	idTitlePattern := regexp.MustCompile(`"id"\s*:\s*(\d+)[^}]*"title"\s*:\s*"([^"]+)"`)
	matches = idTitlePattern.FindAllStringSubmatch(html, -10)
	if len(matches) > 0 {
		fmt.Printf("\n--- Found id-title pairs ---\n")
		for i, match := range matches {
			fmt.Printf("  [%d] ID: %s, Title: %s\n", i+1, match[1], match[2])
		}
	}

	return nil
}
