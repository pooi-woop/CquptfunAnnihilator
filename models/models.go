package models

type Config struct {
	Platform PlatformConfig `mapstructure:"platform"`
	Auth     AuthConfig     `mapstructure:"auth"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Solver   SolverConfig   `mapstructure:"solver"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type PlatformConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Timeout int    `mapstructure:"timeout"`
}

type AuthConfig struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
	Cookie   string `mapstructure:"cookie"`
}

type LLMConfig struct {
	APIKey    string `mapstructure:"api_key"`
	BaseURL   string `mapstructure:"base_url"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type SolverConfig struct {
	DelayMS           int    `mapstructure:"delay_ms"`
	MaxRetries        int    `mapstructure:"max_retries"`
	UserAgent         string `mapstructure:"user_agent"`
	SubmissionDelayS  int    `mapstructure:"submission_delay_s"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    LoginData `json:"data"`
}

type LoginData struct {
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expires_at"`
	UserInfo  interface{} `json:"user_info"`
}

type Problem struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Difficulty  string      `json:"difficulty"`
	Options     []Option    `json:"options,omitempty"`
	Answer      interface{} `json:"answer,omitempty"`
	TestCases   []TestCase  `json:"test_cases,omitempty"`
	ClassroomID     string      `json:"classroom_id,omitempty"`
	TaskID          string      `json:"task_id,omitempty"`
	ClassroomTaskID string      `json:"classroom_task_id,omitempty"`
}

type Option struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	IsSample bool   `json:"is_sample"`
}

type ProblemListResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    ProblemListData `json:"data"`
}

type ProblemListData struct {
	Problems []Problem `json:"problems"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

type SubmissionRequest struct {
	ProblemID int64       `json:"problem_id"`
	Answer    interface{} `json:"answer"`
}

type SubmissionResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    SubmissionResult `json:"data"`
}

type SubmissionResult struct {
	Correct     bool    `json:"correct"`
	Score       float64 `json:"score"`
	Feedback    string  `json:"feedback"`
	PassedCases int     `json:"passed_cases"`
	TotalCases  int     `json:"total_cases"`
}

type LLMRequest struct {
	Model     string       `json:"model"`
	Messages  []LLMMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens,omitempty"`
}

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int        `json:"index"`
	Message      LLMMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
