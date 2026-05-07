package solver

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"CquptFunAnnihilator/analyzer"
	"CquptFunAnnihilator/client"
	"CquptFunAnnihilator/llm"
	"CquptFunAnnihilator/logger"
	"CquptFunAnnihilator/models"
)

type Solver struct {
	httpClient       *client.HttpClient
	llmClient        *llm.LLMClient
	delayMs          int
	maxRetries       int
	baseURL          string
	cookie           string
	submissionDelayS int
}

func NewSolver(httpClient *client.HttpClient, llmClient *llm.LLMClient, delayMs, maxRetries, submissionDelayS int) *Solver {
	logger.Info("Solver initialized",
		zap.Int("delayMs", delayMs),
		zap.Int("maxRetries", maxRetries),
		zap.Int("submissionDelayS", submissionDelayS),
	)

	return &Solver{
		httpClient:       httpClient,
		llmClient:        llmClient,
		delayMs:          delayMs,
		maxRetries:       maxRetries,
		baseURL:          httpClient.GetBaseURL(),
		cookie:           httpClient.GetCookie(),
		submissionDelayS: submissionDelayS,
	}
}

func (s *Solver) SubmitAnswer(problem *models.Problem, answer interface{}, attemptNo int) (*models.SubmissionResult, error) {
	logger.Debug("Submitting answer",
		zap.Int64("problemID", problem.ID),
		zap.String("answerType", fmt.Sprintf("%T", answer)),
	)

	if problem.ClassroomID == "" || problem.TaskID == "" || problem.ClassroomTaskID == "" {
		return nil, fmt.Errorf("problem missing classroom, task ID or classroom task ID")
	}

	code, ok := answer.(string)
	if !ok {
		return nil, fmt.Errorf("answer is not a string (code)")
	}

	resp, err := analyzer.SubmitCode(s.baseURL, s.cookie, problem.ClassroomID, problem.TaskID, problem.ClassroomTaskID, code, "java", attemptNo)
	if err != nil {
		logger.Error("Failed to submit answer",
			zap.Int64("problemID", problem.ID),
			zap.Error(err),
		)
		return nil, err
	}

	if resp.Success {
		logger.Info("Answer submitted successfully",
			zap.Int64("problemID", problem.ID),
			zap.String("message", resp.Message),
		)

		return &models.SubmissionResult{
			Correct:     true,
			Score:       100,
			Feedback:    resp.Message,
			PassedCases: 1,
			TotalCases:  1,
		}, nil
	}

	return nil, fmt.Errorf("submission failed: %s", resp.Message)
}

func (s *Solver) SolveAndSubmit(problem *models.Problem) (*models.SubmissionResult, error) {
	logger.Info("Solving and submitting problem",
		zap.Int64("problemID", problem.ID),
		zap.String("title", problem.Title),
		zap.String("type", problem.Type),
	)

	startTime := time.Now()

	answer, err := s.llmClient.SolveProblem(problem)
	if err != nil {
		logger.Error("LLM failed to solve problem",
			zap.Int64("problemID", problem.ID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("LLM failed to solve problem: %w", err)
	}

	logger.Debug("LLM returned answer",
		zap.Int64("problemID", problem.ID),
		zap.String("answer", truncateString(answer, 200)),
	)

	for i := 0; i < s.maxRetries; i++ {
		logger.Debug("Submission attempt",
			zap.Int64("problemID", problem.ID),
			zap.Int("attempt", i+1),
			zap.Int("maxRetries", s.maxRetries),
		)

		result, err := s.SubmitAnswer(problem, answer, i+1)
		if err != nil {
			logger.Warn("Submission attempt failed",
				zap.Int64("problemID", problem.ID),
				zap.Int("attempt", i+1),
				zap.Error(err),
			)
			time.Sleep(time.Duration(s.delayMs) * time.Millisecond)
			continue
		}

		elapsed := time.Since(startTime)
		logger.Info("Problem solved and submitted successfully",
			zap.Int64("problemID", problem.ID),
			zap.Bool("correct", result.Correct),
			zap.Float64("score", result.Score),
			zap.Duration("totalTime", elapsed),
		)

		return result, nil
	}

	return nil, fmt.Errorf("failed to submit after %d retries", s.maxRetries)
}

func (s *Solver) SolveAndSubmitCoding(problem *models.Problem) (*models.SubmissionResult, error) {
	logger.Info("Solving and submitting coding problem",
		zap.Int64("problemID", problem.ID),
		zap.String("title", problem.Title),
	)

	startTime := time.Now()

	answer, err := s.llmClient.SolveCodingProblem(problem)
	if err != nil {
		logger.Error("LLM failed to solve coding problem",
			zap.Int64("problemID", problem.ID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("LLM failed to solve coding problem: %w", err)
	}

	logger.Debug("LLM returned coding answer",
		zap.Int64("problemID", problem.ID),
		zap.String("code", truncateString(answer, 200)),
	)

	for i := 0; i < s.maxRetries; i++ {
		logger.Debug("Coding submission attempt",
			zap.Int64("problemID", problem.ID),
			zap.Int("attempt", i+1),
		)

		result, err := s.SubmitAnswer(problem, answer, i+1)
		if err != nil {
			logger.Warn("Coding submission attempt failed",
				zap.Int64("problemID", problem.ID),
				zap.Int("attempt", i+1),
				zap.Error(err),
			)
			if strings.Contains(err.Error(), "429") {
				logger.Info("Rate limited, waiting for submission delay",
					zap.Int("delayS", s.submissionDelayS),
				)
				s.waitWithCountdown(s.submissionDelayS, "提交过于频繁，正在等待重提交（此限制由重庆邮电大学官方设定）")
			} else {
				time.Sleep(time.Duration(s.delayMs) * time.Millisecond)
			}
			continue
		}

		elapsed := time.Since(startTime)
		logger.Info("Coding problem solved and submitted successfully",
			zap.Int64("problemID", problem.ID),
			zap.Bool("correct", result.Correct),
			zap.Float64("score", result.Score),
			zap.Duration("totalTime", elapsed),
		)

		return result, nil
	}

	return nil, fmt.Errorf("failed to submit after %d retries", s.maxRetries)
}

func (s *Solver) SolveWithDelay(problem *models.Problem) (*models.SubmissionResult, error) {
	logger.Debug("Waiting before solving",
		zap.Int64("problemID", problem.ID),
		zap.Int("delayMs", s.delayMs),
	)

	time.Sleep(time.Duration(s.delayMs) * time.Millisecond)

	if problem.Type == "coding" {
		return s.SolveAndSubmitCoding(problem)
	}

	return s.SolveAndSubmit(problem)
}

func (s *Solver) BatchSolve(problems []models.Problem) map[int64]*models.SubmissionResult {
	logger.Info("Starting batch solve",
		zap.Int("totalProblems", len(problems)),
	)

	results := make(map[int64]*models.SubmissionResult)
	startTime := time.Now()

	for i, problem := range problems {
		logger.Info("Processing problem in batch",
			zap.Int("current", i+1),
			zap.Int("total", len(problems)),
			zap.Int64("problemID", problem.ID),
			zap.String("title", problem.Title),
			zap.String("type", problem.Type),
		)

		var result *models.SubmissionResult
		var err error

		if problem.Type == "coding" {
			result, err = s.SolveAndSubmitCoding(&problem)
		} else {
			result, err = s.SolveAndSubmit(&problem)
		}

		if err != nil {
			logger.Error("Failed to solve problem in batch",
				zap.Int64("problemID", problem.ID),
				zap.Int("current", i+1),
				zap.Error(err),
			)
			results[problem.ID] = nil
		} else {
			results[problem.ID] = result
			logger.Info("Problem solved in batch",
				zap.Int64("problemID", problem.ID),
				zap.Bool("correct", result.Correct),
				zap.Float64("score", result.Score),
			)
		}

		if i < len(problems)-1 {
			if s.submissionDelayS > 0 {
				s.waitWithCountdown(s.submissionDelayS, "")
			} else {
				logger.Debug("Waiting before next problem",
					zap.Int("delayMs", s.delayMs),
					zap.Int("nextProblem", i+2),
				)
				time.Sleep(time.Duration(s.delayMs) * time.Millisecond)
			}
		}
	}

	elapsed := time.Since(startTime)

	correctCount := 0
	for _, r := range results {
		if r != nil && r.Correct {
			correctCount++
		}
	}

	logger.Info("Batch solve completed",
		zap.Int("totalProblems", len(problems)),
		zap.Int("successfulCount", len(results)),
		zap.Int("correctCount", correctCount),
		zap.Duration("totalTime", elapsed),
	)

	return results
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (s *Solver) waitWithCountdown(seconds int, reason string) {
	if seconds <= 0 {
		return
	}

	if reason == "" {
		fmt.Printf("\n=== 等待提交间隔 (%d秒) ===\n", seconds)
	} else {
		fmt.Printf("\n=== %s ===\n", reason)
		fmt.Printf("等待时间: %d秒\n", seconds)
	}

	for i := seconds; i > 0; i-- {
		fmt.Printf("\r剩余时间: %d秒", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n==========================\n")
}
