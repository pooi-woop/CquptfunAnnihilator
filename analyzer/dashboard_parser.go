package analyzer

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"regexp"
	"time"

	"github.com/go-resty/resty/v2"
)

type ClassroomItem struct {
	Classroom struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		CourseID string `json:"courseId"`
		Status   string `json:"status"`
	} `json:"classroom"`
	Tasks []TaskItem `json:"tasks"`
}

type TaskItem struct {
	ClassroomTaskID    string `json:"classroomTaskId"`
	TaskID             string `json:"taskId"`
	Title              string `json:"title"`
	PublishedAt        string `json:"publishedAt"`
	DueAt              string `json:"dueAt"`
	MyLatestSubmission *struct {
		SubmissionID string `json:"submissionId"`
		AttemptNo    int    `json:"attemptNo"`
		CreatedAt    string `json:"createdAt"`
	} `json:"myLatestSubmission,omitempty"`
}

type DashboardData struct {
	Items []ClassroomItem `json:"items"`
}

func ParseDashboardHTML(htmlContent string) ([]TaskItem, error) {
	pattern := regexp.MustCompile(`<pre class="mt-3 overflow-auto text-xs text-zinc-700">({[\s\S]*?})</pre>`)
	matches := pattern.FindAllStringSubmatch(htmlContent, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("未找到原始数据")
	}

	unescaped := html.UnescapeString(matches[0][1])

	var dashboard DashboardData
	err := json.Unmarshal([]byte(unescaped), &dashboard)
	if err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	var allTasks []TaskItem
	for _, item := range dashboard.Items {
		allTasks = append(allTasks, item.Tasks...)
	}

	return allTasks, nil
}

func FetchDashboardTasks(baseURL, cookie string) ([]TaskItem, error) {
	classrooms, err := FetchDashboardClassrooms(baseURL, cookie)
	if err != nil {
		return nil, err
	}

	var allTasks []TaskItem
	for _, classroom := range classrooms {
		allTasks = append(allTasks, classroom.Tasks...)
	}

	return allTasks, nil
}

func FetchDashboardClassrooms(baseURL, cookie string) ([]ClassroomItem, error) {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	resp, err := client.R().Get(baseURL + "/student/dashboard")
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode())
	}

	return ParseDashboardClassrooms(string(resp.Body()))
}

func ParseDashboardClassrooms(htmlContent string) ([]ClassroomItem, error) {
	pattern := regexp.MustCompile(`<pre class="mt-3 overflow-auto text-xs text-zinc-700">({[\s\S]*?})</pre>`)
	matches := pattern.FindAllStringSubmatch(htmlContent, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("未找到原始数据")
	}

	unescaped := html.UnescapeString(matches[0][1])

	var dashboard DashboardData
	err := json.Unmarshal([]byte(unescaped), &dashboard)
	if err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	return dashboard.Items, nil
}

func SaveDashboardHTML(baseURL, cookie, filename string) error {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	resp, err := client.R().Get(baseURL + "/student/dashboard")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, resp.Body(), 0644)
}

type TaskDetail struct {
	ClassroomID      string          `json:"classroomId"`
	ClassroomTaskID  string          `json:"classroomTaskId"`
	TaskID           string          `json:"taskId"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	KnowledgeModule  string          `json:"knowledgeModule"`
	Stage            int             `json:"stage"`
	AllowLate        bool            `json:"allowLate"`
	DueAt            string          `json:"dueAt"`
	PublishedAt      string          `json:"publishedAt"`
	LatestSubmission *SubmissionInfo `json:"latestSubmission,omitempty"`
}

type SubmissionInfo struct {
	SubmissionID     string           `json:"submissionId"`
	AttemptNo        int              `json:"attemptNo"`
	AIFeedbackStatus string           `json:"aiFeedbackStatus"`
	FeedbackSummary  *FeedbackSummary `json:"feedbackSummary,omitempty"`
}

type FeedbackSummary struct {
	TotalItems int `json:"totalItems"`
}

func ParseTaskDetailHTML(htmlContent string) (*TaskDetail, error) {
	pattern := regexp.MustCompile(`<pre class="mt-3 overflow-auto text-xs text-zinc-700">({[\s\S]*?})</pre>`)
	matches := pattern.FindAllStringSubmatch(htmlContent, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("未找到原始数据")
	}

	unescaped := html.UnescapeString(matches[0][1])

	var taskData struct {
		Classroom struct {
			ID string `json:"id"`
		} `json:"classroom"`
		ClassroomTask struct {
			ID          string `json:"id"`
			ClassroomID string `json:"classroomId"`
			TaskID      string `json:"taskId"`
			Settings    struct {
				AllowLate bool `json:"allowLate"`
			} `json:"settings"`
			DueAt       string `json:"dueAt"`
			PublishedAt string `json:"publishedAt"`
		} `json:"classroomTask"`
		Task struct {
			ID              string `json:"id"`
			Title           string `json:"title"`
			Description     string `json:"description"`
			KnowledgeModule string `json:"knowledgeModule"`
			Stage           int    `json:"stage"`
		} `json:"task"`
		Submissions []struct {
			ID               string `json:"id"`
			AttemptNo        int    `json:"attemptNo"`
			AIFeedbackStatus string `json:"aiFeedbackStatus"`
			FeedbackSummary  struct {
				TotalItems int `json:"totalItems"`
			} `json:"feedbackSummary"`
		} `json:"submissions"`
	}

	err := json.Unmarshal([]byte(unescaped), &taskData)
	if err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	detail := &TaskDetail{
		ClassroomID:     taskData.Classroom.ID,
		ClassroomTaskID: taskData.ClassroomTask.ID,
		TaskID:          taskData.Task.ID,
		Title:           taskData.Task.Title,
		Description:     taskData.Task.Description,
		KnowledgeModule: taskData.Task.KnowledgeModule,
		Stage:           taskData.Task.Stage,
		AllowLate:       taskData.ClassroomTask.Settings.AllowLate,
		DueAt:           taskData.ClassroomTask.DueAt,
		PublishedAt:     taskData.ClassroomTask.PublishedAt,
	}

	if len(taskData.Submissions) > 0 {
		sub := taskData.Submissions[0]
		detail.LatestSubmission = &SubmissionInfo{
			SubmissionID:     sub.ID,
			AttemptNo:        sub.AttemptNo,
			AIFeedbackStatus: sub.AIFeedbackStatus,
			FeedbackSummary: &FeedbackSummary{
				TotalItems: sub.FeedbackSummary.TotalItems,
			},
		}
	}

	return detail, nil
}

func FetchTaskDetail(baseURL, cookie, classroomID, taskID string) (*TaskDetail, error) {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if cookie != "" {
		client.SetHeader("Cookie", cookie)
	}

	url := fmt.Sprintf("%s/student/classrooms/%s/tasks/%s", baseURL, classroomID, taskID)

	resp, err := client.R().Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP状态码: %d", resp.StatusCode())
	}

	return ParseTaskDetailHTML(string(resp.Body()))
}
