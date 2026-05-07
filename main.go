package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/spf13/viper"

	"CquptFunAnnihilator/analyzer"
	"CquptFunAnnihilator/client"
	"CquptFunAnnihilator/fetcher"
	"CquptFunAnnihilator/llm"
	"CquptFunAnnihilator/logger"
	"CquptFunAnnihilator/models"
	"CquptFunAnnihilator/probe"
	"CquptFunAnnihilator/scraper"
	"CquptFunAnnihilator/solver"
)

var (
	configPath     string
	cookie         string
	problemID      int64
	listOnly       bool
	solveAll       bool
	probeOnly      bool
	scrapeOnly     bool
	analyzeOnly    bool
	dashboardOnly  bool
	parseDashboard bool
	classroomID    string
	taskID         string
)

func init() {
	flag.StringVar(&configPath, "config", "config.yaml", "Path to config file")
	flag.StringVar(&cookie, "cookie", "", "Cookie for authentication")
	flag.Int64Var(&problemID, "problem-id", 0, "Specific problem ID to solve")
	flag.BoolVar(&listOnly, "list", false, "List all problems only")
	flag.BoolVar(&solveAll, "solve-all", false, "Solve all problems")
	flag.BoolVar(&probeOnly, "probe", false, "Probe for valid API endpoints")
	flag.BoolVar(&scrapeOnly, "scrape", false, "Scrape frontend for API endpoints")
	flag.BoolVar(&analyzeOnly, "analyze", false, "Analyze homepage for API clues")
	flag.BoolVar(&dashboardOnly, "dashboard", false, "Analyze dashboard for embedded data")
	flag.BoolVar(&parseDashboard, "parse-dashboard", false, "Parse tasks from dashboard HTML")
	flag.StringVar(&classroomID, "classroom-id", "", "Classroom ID for task detail")
	flag.StringVar(&taskID, "task-id", "", "Task ID for task detail")
}

func loadConfig() (*models.Config, error) {
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg models.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func main() {
	flag.Parse()

	cfg, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	if err := logger.InitWithFile(&cfg.Logging); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("=== CquptFunAnnihilator Started ===",
		zap.String("version", "1.0.0"),
		zap.String("configPath", configPath),
	)

	if cookie != "" {
		cfg.Auth.Cookie = cookie
	}

	if cfg.Auth.Cookie == "" {
		logger.Fatal("Cookie is required",
			zap.Bool("hasCookie", cfg.Auth.Cookie != ""),
		)
		fmt.Println("Error: Cookie is required")
		fmt.Println("Set it in config.yaml under auth.cookie or use -cookie flag")
		os.Exit(1)
	}

	if cfg.LLM.APIKey == "" || cfg.LLM.APIKey == "your_api_key_here" {
		logger.Fatal("LLM API key is required",
			zap.String("apiKey", cfg.LLM.APIKey),
		)
		fmt.Println("Error: LLM API key is required")
		fmt.Println("Set it in config.yaml under llm.api_key")
		os.Exit(1)
	}

	logger.Info("Configuration loaded successfully",
		zap.String("baseURL", cfg.Platform.BaseURL),
		zap.String("llmModel", cfg.LLM.Model),
		zap.Int("llmMaxTokens", cfg.LLM.MaxTokens),
		zap.Int("solverDelayMs", cfg.Solver.DelayMS),
		zap.Int("solverMaxRetries", cfg.Solver.MaxRetries),
		zap.Int("submissionDelayS", cfg.Solver.SubmissionDelayS),
	)
	
	fmt.Printf("提交间隔时间配置: %d秒\n", cfg.Solver.SubmissionDelayS)

	httpClient := client.NewHttpClient(
		cfg.Platform.BaseURL,
		cfg.Platform.Timeout,
		cfg.Solver.UserAgent,
	)

	httpClient.SetCookie(cfg.Auth.Cookie)

	logger.Info("Using cookie for authentication",
		zap.String("baseURL", cfg.Platform.BaseURL),
	)

	llmClient := llm.NewLLMClient(
		cfg.LLM.APIKey,
		cfg.LLM.BaseURL,
		cfg.LLM.Model,
		cfg.LLM.MaxTokens,
	)

	if probeOnly {
		fmt.Println("=== Probing API endpoints ===")
		fmt.Printf("Base URL: %s\n\n", cfg.Platform.BaseURL)

		results := probe.ProbeAPIEndpoints(cfg.Platform.BaseURL, cfg.Auth.Cookie)
		validEndpoints := probe.FindValidEndpoints(results)

		fmt.Println("\n=== Valid JSON endpoints found: ===")
		if len(validEndpoints) == 0 {
			fmt.Println("No valid JSON endpoints found")
		} else {
			for _, endpoint := range validEndpoints {
				fmt.Printf("  %s\n", endpoint)
			}
		}
		return
	}

	if scrapeOnly {
		fmt.Println("=== Scraping frontend for API endpoints ===")
		fmt.Printf("Base URL: %s\n\n", cfg.Platform.BaseURL)

		endpoints, err := scraper.ComprehensiveScrape(cfg.Platform.BaseURL, cfg.Auth.Cookie)
		if err != nil {
			fmt.Printf("Error scraping: %v\n", err)
			return
		}

		fmt.Println("\n=== Scraped endpoints found ===")
		if len(endpoints) == 0 {
			fmt.Println("No endpoints found")
		} else {
			for _, ep := range endpoints {
				fmt.Printf("  [%d] %-40s (Source: %s)\n", ep.Confidence, ep.Path, ep.Source)
			}
		}

		validEndpoints := scraper.ProbeScrapedEndpoints(cfg.Platform.BaseURL, cfg.Auth.Cookie, endpoints)

		fmt.Println("\n=== Valid endpoints ===")
		if len(validEndpoints) == 0 {
			fmt.Println("No valid endpoints found")
		} else {
			for _, ep := range validEndpoints {
				fmt.Printf("  %s\n", ep.Path)
			}
		}

		return
	}

	if analyzeOnly {
		fmt.Println("=== Analyzing homepage for API clues ===")
		fmt.Printf("Base URL: %s\n\n", cfg.Platform.BaseURL)

		if err := analyzer.AnalyzeHomepage(cfg.Platform.BaseURL, cfg.Auth.Cookie); err != nil {
			fmt.Printf("Error analyzing: %v\n", err)
			return
		}

		analyzer.TestSubdomains(cfg.Platform.BaseURL, cfg.Auth.Cookie)

		return
	}

	if dashboardOnly {
		fmt.Println("=== Analyzing dashboard for embedded data ===")
		fmt.Printf("Base URL: %s\n\n", cfg.Platform.BaseURL)

		if err := analyzer.AnalyzeDashboard(cfg.Platform.BaseURL, cfg.Auth.Cookie); err != nil {
			fmt.Printf("Error analyzing dashboard: %v\n", err)
			return
		}

		return
	}

	if parseDashboard {
		fmt.Println("=== Parsing tasks from dashboard HTML ===")
		fmt.Printf("Base URL: %s\n\n", cfg.Platform.BaseURL)

		tasks, err := analyzer.FetchDashboardTasks(cfg.Platform.BaseURL, cfg.Auth.Cookie)
		if err != nil {
			fmt.Printf("Error fetching tasks: %v\n", err)
			return
		}

		fmt.Printf("Found %d tasks:\n\n", len(tasks))
		for i, task := range tasks {
			fmt.Printf("%d. %s\n", i+1, task.Title)
			fmt.Printf("   ClassroomTaskID: %s\n", task.ClassroomTaskID)
			fmt.Printf("   TaskID: %s\n", task.TaskID)
			fmt.Printf("   Due: %s\n", task.DueAt)
			if task.MyLatestSubmission != nil {
				fmt.Printf("   Submitted: Yes (Attempt #%d)\n", task.MyLatestSubmission.AttemptNo)
			} else {
				fmt.Printf("   Submitted: No\n")
			}
			fmt.Println()
		}

		return
	}

	if classroomID != "" && taskID != "" {
		fmt.Println("=== Fetching task detail ===")
		fmt.Printf("Classroom: %s, Task: %s\n\n", classroomID, taskID)

		detail, err := analyzer.FetchTaskDetail(cfg.Platform.BaseURL, cfg.Auth.Cookie, classroomID, taskID)
		if err != nil {
			fmt.Printf("Error fetching task detail: %v\n", err)
			return
		}

		fmt.Printf("Title: %s\n", detail.Title)
		fmt.Printf("Knowledge Module: %s\n", detail.KnowledgeModule)
		fmt.Printf("Stage: %d\n", detail.Stage)
		fmt.Printf("Allow Late: %v\n", detail.AllowLate)
		fmt.Printf("Published: %s\n", detail.PublishedAt)
		fmt.Printf("Due: %s\n", detail.DueAt)
		if detail.LatestSubmission != nil {
			fmt.Printf("Latest Submission: %s (Attempt #%d, Status: %s)\n",
				detail.LatestSubmission.SubmissionID,
				detail.LatestSubmission.AttemptNo,
				detail.LatestSubmission.AIFeedbackStatus)
		}
		fmt.Println("\n=== Description ===")
		fmt.Println(detail.Description)

		return
	}

	problemFetcher := fetcher.NewFetcher(httpClient)
	problemSolver := solver.NewSolver(httpClient, llmClient, cfg.Solver.DelayMS, cfg.Solver.MaxRetries, cfg.Solver.SubmissionDelayS)

	if listOnly {
		listProblems(problemFetcher)
		return
	}

	if problemID > 0 {
		solveSingleProblem(problemFetcher, problemSolver, problemID)
		return
	}

	if solveAll {
		solveAllProblems(problemFetcher, problemSolver)
		return
	}

	fmt.Println("CquptFunAnnihilator - Auto Problem Solver with AI")
	fmt.Println("=================================================")
	fmt.Println("\nUsage:")
	fmt.Println("  -config <path>     Path to config file (default: config.yaml)")
	fmt.Println("  -cookie <cookie>   Authentication cookie")
	fmt.Println("  -problem-id <id>   Solve a specific problem")
	fmt.Println("  -list              List all available problems")
	fmt.Println("  -solve-all         Solve all problems")
	fmt.Println("\nExamples:")
	fmt.Println("  ./CquptFunAnnihilator -cookie \"session=xxx\" -list")
	fmt.Println("  ./CquptFunAnnihilator -cookie \"session=xxx\" -problem-id 123")
	fmt.Println("  ./CquptFunAnnihilator -cookie \"session=xxx\" -solve-all")

	logger.Info("No action specified, showing help")
}

func listProblems(fetcher *fetcher.Fetcher) {
	logger.Info("Fetching problem list...")

	problems, err := fetcher.GetAllProblems()
	if err != nil {
		logger.Fatal("Failed to fetch problems",
			zap.Error(err),
		)
	}

	logger.Info("Problem list fetched successfully",
		zap.Int("totalProblems", len(problems)),
	)

	fmt.Printf("\nTotal Problems: %d\n\n", len(problems))
	fmt.Printf("%-10s %-40s %-15s %-10s\n", "ID", "Title", "Type", "Difficulty")
	fmt.Println(strings.Repeat("-", 75))

	for _, p := range problems {
		fmt.Printf("%-10d %-40s %-15s %-10s\n", p.ID, truncate(p.Title, 38), p.Type, p.Difficulty)
	}
}

func solveSingleProblem(fetcher *fetcher.Fetcher, solv *solver.Solver, problemID int64) {
	logger.Info("Fetching problem details",
		zap.Int64("problemID", problemID),
	)

	problem, err := fetcher.GetProblem(problemID)
	if err != nil {
		logger.Fatal("Failed to fetch problem",
			zap.Int64("problemID", problemID),
			zap.Error(err),
		)
	}

	logger.Info("Problem fetched successfully",
		zap.Int64("problemID", problem.ID),
		zap.String("title", problem.Title),
		zap.String("type", problem.Type),
		zap.String("difficulty", problem.Difficulty),
	)

	fmt.Printf("\nProblem: %s\n", problem.Title)
	fmt.Printf("Type: %s\n", problem.Type)
	fmt.Printf("Difficulty: %s\n", problem.Difficulty)
	fmt.Printf("Description: %s\n\n", truncate(problem.Description, 100))

	var result *models.SubmissionResult
	if problem.Type == "coding" {
		result, err = solv.SolveAndSubmitCoding(problem)
	} else {
		result, err = solv.SolveAndSubmit(problem)
	}

	if err != nil {
		logger.Fatal("Failed to solve and submit problem",
			zap.Int64("problemID", problemID),
			zap.Error(err),
		)
	}

	logger.Info("Problem solved and submitted",
		zap.Int64("problemID", problemID),
		zap.Bool("correct", result.Correct),
		zap.Float64("score", result.Score),
		zap.String("feedback", result.Feedback),
	)

	fmt.Printf("Result: %v\n", result.Correct)
	fmt.Printf("Score: %.2f\n", result.Score)
	fmt.Printf("Feedback: %s\n", result.Feedback)
}

func solveAllProblems(fetcher *fetcher.Fetcher, solv *solver.Solver) {
	logger.Info("Fetching all problems for batch solving...")

	problems, err := fetcher.GetAllProblems()
	if err != nil {
		logger.Fatal("Failed to fetch problems for batch solving",
			zap.Error(err),
		)
	}

	logger.Info("All problems fetched successfully",
		zap.Int("totalProblems", len(problems)),
	)

	fmt.Printf("\nTotal problems to solve: %d\n", len(problems))

	results := solv.BatchSolve(problems)

	correctCount := 0
	totalCount := len(results)

	for pid, result := range results {
		if result != nil && result.Correct {
			correctCount++
		} else if result != nil {
			fmt.Printf("Problem %d: Failed (Score: %.2f)\n", pid, result.Score)
		} else {
			fmt.Printf("Problem %d: Error\n", pid)
		}
	}

	fmt.Printf("\n================== Summary ==================\n")
	fmt.Printf("Total: %d | Correct: %d | Failed: %d\n", totalCount, correctCount, totalCount-correctCount)
	fmt.Printf("Accuracy: %.2f%%\n", float64(correctCount)/float64(totalCount)*100)

	logger.Info("Batch solving completed",
		zap.Int("total", totalCount),
		zap.Int("correct", correctCount),
		zap.Int("failed", totalCount-correctCount),
		zap.Float64("accuracy", float64(correctCount)/float64(totalCount)*100),
	)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
