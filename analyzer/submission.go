package analyzer

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type SubmissionRequest struct {
	Language string `json:"language"`
	CodeText string `json:"codeText"`
}

type SubmissionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SubmissionPayload struct {
	TaskId          string `json:"taskId"`
	ClassroomTaskId string `json:"classroomTaskId"`
	Code            string `json:"code"`
	AttemptNo       int    `json:"attemptNo"`
}

func SubmitCode(baseURL, cookie, classroomID, taskID, classroomTaskID, code, language string, attemptNo int) (*SubmissionResponse, error) {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	url := fmt.Sprintf("%s/api/proxy/classrooms/%s/tasks/%s/submissions", baseURL, classroomID, classroomTaskID)

	fmt.Printf("=== 提交信息 ===\n")
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Language: %s\n", language)
	fmt.Printf("Code length: %d\n", len(code))
	fmt.Printf("Cookie present: %s\n", cookie != "")
	fmt.Printf("================\n")

	resp, err := client.R().
		SetHeader("Referer", fmt.Sprintf("%s/student/classrooms/%s/tasks/%s", baseURL, classroomID, classroomTaskID)).
		SetHeader("Origin", baseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Priority", "u=1, i").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36 Edg/147.0.0.0").
		SetBody(map[string]interface{}{
			"content": map[string]interface{}{
				"language": "auto",
				"codeText": code,
			},
		}).
		Post(url)

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	fmt.Printf("=== 响应信息 ===\n")
	fmt.Printf("状态码: %d\n", resp.StatusCode())
	fmt.Printf("Content-Type: %s\n", resp.Header().Get("Content-Type"))
	fmt.Printf("X-RSC: %s\n", resp.Header().Get("X-RSC"))

	body := string(resp.Body())
	if len(body) > 2000 {
		fmt.Printf("响应内容（前2000字符）: %s\n", body[:2000])
	} else {
		fmt.Printf("响应内容: %s\n", body)
	}
	fmt.Printf("====================\n")

	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		contentType := resp.Header().Get("Content-Type")

		if strings.Contains(contentType, "application/json") {
			var response SubmissionResponse
			if err := json.Unmarshal(resp.Body(), &response); err == nil {
				if response.Success {
					return &response, nil
				}
			}
			if strings.Contains(body, "\"status\":\"SUBMITTED\"") || strings.Contains(body, "\"id\":\"") {
				return &SubmissionResponse{
					Success: true,
					Message: "提交成功（服务器返回JSON）",
				}, nil
			}
		}

		if strings.Contains(contentType, "text/x-component") {
			if strings.Contains(body, "success") || strings.Contains(body, "Success") || strings.Contains(body, "成功") {
				return &SubmissionResponse{
					Success: true,
					Message: "提交成功（RSC响应）",
				}, nil
			}
		}

		if strings.Contains(body, "submissions") && strings.Contains(body, "attemptNo") {
			return &SubmissionResponse{
				Success: true,
				Message: "提交成功（已更新任务页面）",
			}, nil
		}

		return &SubmissionResponse{
			Success: true,
			Message: "提交成功（服务器返回HTML）",
		}, nil
	}

	if resp.StatusCode() == 500 {
		return nil, fmt.Errorf("服务器内部错误")
	}

	return nil, fmt.Errorf("提交失败，状态码: %d", resp.StatusCode())
}
