package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"CquptFunAnnihilator/logger"
	"CquptFunAnnihilator/models"
)

type LLMClient struct {
	apiKey    string
	baseURL   string
	model     string
	maxTokens int
	client    *http.Client
}

func NewLLMClient(apiKey, baseURL, model string, maxTokens int) *LLMClient {
	logger.Info("LLM client initialized",
		zap.String("baseURL", baseURL),
		zap.String("model", model),
		zap.Int("maxTokens", maxTokens),
	)

	return &LLMClient{
		apiKey:    apiKey,
		baseURL:   baseURL,
		model:     model,
		maxTokens: maxTokens,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *LLMClient) Chat(messages []models.LLMMessage) (string, error) {
	return c.ChatWithContext(context.Background(), messages)
}

func (c *LLMClient) ChatWithContext(ctx context.Context, messages []models.LLMMessage) (string, error) {
	logger.Debug("Sending chat request to LLM",
		zap.String("model", c.model),
		zap.Int("messageCount", len(messages)),
	)

	req := models.LLMRequest{
		Model:     c.model,
		Messages:  messages,
		MaxTokens: c.maxTokens,
	}

	body, err := json.Marshal(req)
	if err != nil {
		logger.Error("Failed to marshal LLM request",
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.Debug("LLM request body",
		zap.String("body", string(body)),
	)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		logger.Error("Failed to create HTTP request for LLM",
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	logger.Debug("Sending request to LLM API",
		zap.String("url", c.baseURL+"/chat/completions"),
	)

	startTime := time.Now()
	resp, err := c.client.Do(httpReq)
	elapsed := time.Since(startTime)

	if err != nil {
		logger.Error("LLM API request failed",
			zap.Duration("elapsed", elapsed),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	logger.Debug("LLM API response received",
		zap.Int("statusCode", resp.StatusCode),
		zap.Duration("elapsed", elapsed),
	)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read LLM response body",
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debug("LLM response body",
		zap.String("body", string(respBody)),
	)

	if resp.StatusCode != http.StatusOK {
		logger.Error("LLM API request failed with non-OK status",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var llmResp models.LLMResponse
	if err := json.Unmarshal(respBody, &llmResp); err != nil {
		logger.Error("Failed to unmarshal LLM response",
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		logger.Error("LLM response has no choices")
		return "", fmt.Errorf("no choices in response")
	}

	content := llmResp.Choices[0].Message.Content

	logger.Info("LLM chat completed successfully",
		zap.Int("promptTokens", llmResp.Usage.PromptTokens),
		zap.Int("completionTokens", llmResp.Usage.CompletionTokens),
		zap.Int("totalTokens", llmResp.Usage.TotalTokens),
		zap.Duration("elapsed", elapsed),
		zap.Int("responseLength", len(content)),
	)

	return content, nil
}

func (c *LLMClient) SolveProblem(problem *models.Problem) (string, error) {
	logger.Info("Solving problem with LLM",
		zap.Int64("problemID", problem.ID),
		zap.String("title", problem.Title),
		zap.String("type", problem.Type),
	)

	systemPrompt := `你是一个智能解题助手，专门帮助学生解答题目。请仔细阅读题目要求，给出准确、完整的答案。

回答格式要求：
- 直接给出答案，不要添加解释（除非题目要求解释）
- 如果是编程题，只输出代码，不要添加markdown标记
- 如果是问答题，给出简洁明了的答案
- 确保答案准确无误`

	userPrompt := fmt.Sprintf(`题目：%s

题目描述：%s

请解答这道题目，直接给出答案。`, problem.Title, problem.Description)

	if len(problem.TestCases) > 0 {
		userPrompt += "\n\n测试用例：\n"
		for i, tc := range problem.TestCases {
			userPrompt += fmt.Sprintf("用例 %d:\n输入: %s\n预期输出: %s\n", i+1, tc.Input, tc.Expected)
		}
	}

	logger.Debug("LLM problem solving prompt prepared",
		zap.String("systemPrompt", systemPrompt),
		zap.String("userPrompt", userPrompt),
	)

	messages := []models.LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	answer, err := c.Chat(messages)
	if err != nil {
		logger.Error("Failed to solve problem with LLM",
			zap.Int64("problemID", problem.ID),
			zap.Error(err),
		)
		return "", err
	}

	logger.Info("Successfully got solution from LLM",
		zap.Int64("problemID", problem.ID),
		zap.Int("answerLength", len(answer)),
		zap.String("answerPreview", truncateString(answer, 100)),
	)

	return answer, nil
}

func (c *LLMClient) SolveCodingProblem(problem *models.Problem) (string, error) {
	logger.Info("Solving coding problem with LLM",
		zap.Int64("problemID", problem.ID),
		zap.String("title", problem.Title),
	)

	systemPrompt := `你是一个编程题解题助手。请仔细阅读编程题目，写出正确、高效的代码解决方案。

要求：
- 只输出代码，不要任何解释
- 代码要完整可运行
- 注意语言和题目的具体要求（Python/Java/C++等）
- 如果有多个测试用例，确保代码能通过所有用例`

	userPrompt := fmt.Sprintf(`编程题：%s

题目描述：%s
`, problem.Title, problem.Description)

	if len(problem.TestCases) > 0 {
		userPrompt += "测试用例：\n"
		for i, tc := range problem.TestCases {
			userPrompt += fmt.Sprintf("用例 %d:\n输入: %s\n预期输出: %s\n", i+1, tc.Input, tc.Expected)
		}
	}

	logger.Debug("LLM coding problem prompt prepared",
		zap.String("systemPrompt", systemPrompt),
		zap.String("userPrompt", userPrompt),
	)

	messages := []models.LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	answer, err := c.Chat(messages)
	if err != nil {
		logger.Error("Failed to solve coding problem with LLM",
			zap.Int64("problemID", problem.ID),
			zap.Error(err),
		)
		return "", err
	}

	cleanedAnswer := cleanCode(answer)

	logger.Info("Successfully got coding solution from LLM",
		zap.Int64("problemID", problem.ID),
		zap.Int("codeLength", len(cleanedAnswer)),
		zap.String("codePreview", truncateString(cleanedAnswer, 100)),
	)

	return cleanedAnswer, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func cleanCode(content string) string {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "```java\n") && strings.HasSuffix(content, "\n```") {
		content = content[7 : len(content)-3]
	} else if strings.HasPrefix(content, "```\n") && strings.HasSuffix(content, "\n```") {
		content = content[4 : len(content)-3]
	}

	return strings.TrimSpace(content)
}
