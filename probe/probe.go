package probe

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type ProbeResult struct {
	Path          string
	StatusCode    int
	ContentLength int
	ContentType   string
	IsJSON        bool
	Error         error
}

func ProbeAPIEndpoints(baseURL string, cookie string) []ProbeResult {
	client := resty.New().
		SetTimeout(10 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	possiblePaths := []string{
		"/",
		"/student/dashboard",
		"/api",
		"/api/",
		"/api/problems",
		"/api/v1/problems",
		"/api/v2/problems",
		"/api/v3/problems",
		"/student/api/problems",
		"/student/v1/problems",
		"/student/v2/problems",
		"/api/questions",
		"/api/v1/questions",
		"/course/api/problems",
		"/api/course/problems",
		"/backend/api/problems",
		"/api/tasks",
		"/problems",
		"/v1/problems",
		"/api/graphql",
		"/graphql",
		"/api/quiz",
		"/api/v1/quiz",
		"/api/exercises",
		"/api/v1/exercises",
		"/api/assignments",
		"/api/v1/assignments",
		"/api/courses",
		"/api/v1/courses",
		"/api/lesson",
		"/api/v1/lesson",
		"/api/chapter",
		"/api/v1/chapter",
		"/api/homework",
		"/api/v1/homework",
		"/api/exam",
		"/api/v1/exam",
		"/api/test",
		"/api/v1/test",
		"/api/challenges",
		"/api/v1/challenges",
		"/api/activities",
		"/api/v1/activities",
		"/api/items",
		"/api/v1/items",
		"/api/list/problems",
		"/api/get/problems",
		"/api/fetch/problems",
		"/api/session",
		"/api/user",
		"/api/me",
		"/api/auth",
		"/api/login",
		"/api/validate",
		"/api/check",
		"/.well-known/openid-configuration",
		"/api/assessment",
		"/api/v1/assessment",
		"/api/quiz/questions",
		"/api/v1/quiz/questions",
		"/api/course/problems",
		"/api/v1/course/problems",
		"/api/submissions",
		"/api/v1/submissions",
		"/student/submissions",
		"/api/activity",
		"/api/v1/activity",
		"/api/course/activity",
		"/api/task/list",
		"/api/v1/task/list",
		"/api/exercise/list",
		"/api/v1/exercise/list",
		"/api/data/problems",
		"/api/data/questions",
	}

	postPaths := []string{
		"/api/graphql",
		"/graphql",
		"/api/query",
		"/api/v1/query",
		"/api/execute",
		"/api/v1/execute",
	}

	results := make([]ProbeResult, 0)

	fmt.Println("\n--- Testing POST endpoints ---")
	for _, path := range postPaths {
		url := baseURL + path
		resp, err := client.R().SetBody(map[string]interface{}{}).Post(url)

		result := ProbeResult{
			Path:       path,
			StatusCode: resp.StatusCode(),
			Error:      err,
		}

		if err == nil {
			result.ContentLength = int(resp.Size())
			result.ContentType = resp.Header().Get("Content-Type")
			result.IsJSON = strings.Contains(result.ContentType, "application/json")
		}

		results = append(results, result)
		if result.StatusCode != 404 {
			fmt.Printf("POST %-25s -> Status: %d, Content-Type: %s\n",
				path, result.StatusCode, truncateContentType(result.ContentType))
		}
	}

	for _, path := range possiblePaths {
		url := baseURL + path
		resp, err := client.R().Get(url)

		result := ProbeResult{
			Path:       path,
			StatusCode: resp.StatusCode(),
			Error:      err,
		}

		if err == nil {
			result.ContentLength = int(resp.Size())
			result.ContentType = resp.Header().Get("Content-Type")
			result.IsJSON = strings.Contains(result.ContentType, "application/json")
		}

		results = append(results, result)

		if result.StatusCode != 404 {
			fmt.Printf("Probing %-35s -> Status: %d, Content-Type: %s, IsJSON: %t\n",
				path, result.StatusCode, truncateContentType(result.ContentType), result.IsJSON)
		}
	}

	fmt.Println("\n--- All 404 responses ---")
	for _, r := range results {
		if r.StatusCode == 404 {
			fmt.Printf("  %s\n", r.Path)
		}
	}

	return results
}

func truncateContentType(ct string) string {
	if len(ct) > 40 {
		return ct[:37] + "..."
	}
	return ct
}

func FindValidEndpoints(results []ProbeResult) []string {
	var valid []string
	for _, r := range results {
		if r.StatusCode == http.StatusOK && r.IsJSON {
			valid = append(valid, r.Path)
		}
	}
	return valid
}
