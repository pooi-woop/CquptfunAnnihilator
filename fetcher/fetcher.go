package fetcher

import (
	"fmt"

	"go.uber.org/zap"

	"CquptFunAnnihilator/analyzer"
	"CquptFunAnnihilator/client"
	"CquptFunAnnihilator/logger"
	"CquptFunAnnihilator/models"
)

type Fetcher struct {
	httpClient *client.HttpClient
	baseURL    string
	cookie     string
}

func NewFetcher(httpClient *client.HttpClient) *Fetcher {
	logger.Info("Problem fetcher initialized")
	return &Fetcher{
		httpClient: httpClient,
		baseURL:    httpClient.GetBaseURL(),
		cookie:     httpClient.GetCookie(),
	}
}

func (f *Fetcher) GetProblemList(page, pageSize int) (*models.ProblemListData, error) {
	logger.Debug("Fetching problem list from dashboard",
		zap.Int("page", page),
		zap.Int("pageSize", pageSize),
	)

	classrooms, err := analyzer.FetchDashboardClassrooms(f.baseURL, f.cookie)
	if err != nil {
		logger.Error("Failed to fetch problem list from dashboard",
			zap.Error(err),
		)
		return nil, err
	}

	var problems []models.Problem
	problemIndex := 0

	for _, classroom := range classrooms {
		for _, task := range classroom.Tasks {
			problemIndex++

			detail, err := analyzer.FetchTaskDetail(f.baseURL, f.cookie, classroom.Classroom.ID, task.ClassroomTaskID)
			if err != nil {
				logger.Warn("Failed to fetch task detail",
					zap.String("taskID", task.ClassroomTaskID),
					zap.Error(err),
				)
				continue
			}

			problems = append(problems, models.Problem{
				ID:              int64(problemIndex),
				Title:           task.Title,
				Description:     detail.Description,
				Type:            "coding",
				Difficulty:      "medium",
				ClassroomID:     classroom.Classroom.ID,
				TaskID:          task.TaskID,
				ClassroomTaskID: task.ClassroomTaskID,
			})
		}
	}

	logger.Info("Successfully fetched problem list from dashboard",
		zap.Int("totalProblems", len(problems)),
	)

	return &models.ProblemListData{
		Problems: problems,
		Total:    len(problems),
		Page:     1,
		PageSize: len(problems),
	}, nil
}

func (f *Fetcher) GetProblem(problemID int64) (*models.Problem, error) {
	logger.Debug("Fetching problem details",
		zap.Int64("problemID", problemID),
	)

	classrooms, err := analyzer.FetchDashboardClassrooms(f.baseURL, f.cookie)
	if err != nil {
		logger.Error("Failed to fetch classrooms",
			zap.Int64("problemID", problemID),
			zap.Error(err),
		)
		return nil, err
	}

	problemIndex := 0
	for _, classroom := range classrooms {
		for _, task := range classroom.Tasks {
			problemIndex++
			if int64(problemIndex) == problemID {
				detail, err := analyzer.FetchTaskDetail(f.baseURL, f.cookie, classroom.Classroom.ID, task.ClassroomTaskID)
				if err != nil {
					logger.Error("Failed to fetch task detail",
						zap.String("taskID", task.ClassroomTaskID),
						zap.Error(err),
					)
					return nil, err
				}

				problem := &models.Problem{
					ID:              problemID,
					Title:           task.Title,
					Description:     detail.Description,
					Type:            "coding",
					Difficulty:      "medium",
					ClassroomID:     classroom.Classroom.ID,
					TaskID:          task.TaskID,
					ClassroomTaskID: task.ClassroomTaskID,
				}

				logger.Info("Successfully fetched problem",
					zap.Int64("problemID", problem.ID),
					zap.String("title", problem.Title),
					zap.String("type", problem.Type),
				)

				return problem, nil
			}
		}
	}

	return nil, fmt.Errorf("problem ID %d out of range", problemID)
}

func (f *Fetcher) GetAllProblems() ([]models.Problem, error) {
	logger.Info("Starting to fetch all problems from dashboard")

	classrooms, err := analyzer.FetchDashboardClassrooms(f.baseURL, f.cookie)
	if err != nil {
		logger.Error("Failed to fetch classrooms from dashboard",
			zap.Error(err),
		)
		return nil, err
	}

	var allProblems []models.Problem
	problemIndex := 0

	for _, classroom := range classrooms {
		for _, task := range classroom.Tasks {
			problemIndex++

			detail, err := analyzer.FetchTaskDetail(f.baseURL, f.cookie, classroom.Classroom.ID, task.ClassroomTaskID)
			if err != nil {
				logger.Warn("Failed to fetch task detail, skipping",
					zap.String("taskID", task.ClassroomTaskID),
					zap.Error(err),
				)
				continue
			}

			problem := models.Problem{
				ID:              int64(problemIndex),
				Title:           task.Title,
				Description:     detail.Description,
				Type:            "coding",
				Difficulty:      "medium",
				ClassroomID:     classroom.Classroom.ID,
				TaskID:          task.TaskID,
				ClassroomTaskID: task.ClassroomTaskID,
			}
			allProblems = append(allProblems, problem)

			logger.Info("Fetched problem",
				zap.Int("index", problemIndex),
				zap.String("title", task.Title),
			)
		}
	}

	logger.Info("Finished fetching all problems",
		zap.Int("totalProblems", len(allProblems)),
	)

	return allProblems, nil
}

func (f *Fetcher) GetProblemsByType(problemType string) ([]models.Problem, error) {
	logger.Info("Filtering problems by type",
		zap.String("problemType", problemType),
	)

	allProblems, err := f.GetAllProblems()
	if err != nil {
		return nil, err
	}

	var filtered []models.Problem
	for _, p := range allProblems {
		if p.Type == problemType {
			filtered = append(filtered, p)
		}
	}

	logger.Info("Filtered problems by type",
		zap.String("problemType", problemType),
		zap.Int("filteredCount", len(filtered)),
	)

	return filtered, nil
}

func (f *Fetcher) GetProblemsByDifficulty(difficulty string) ([]models.Problem, error) {
	logger.Info("Filtering problems by difficulty",
		zap.String("difficulty", difficulty),
	)

	allProblems, err := f.GetAllProblems()
	if err != nil {
		return nil, err
	}

	var filtered []models.Problem
	for _, p := range allProblems {
		if p.Difficulty == difficulty {
			filtered = append(filtered, p)
		}
	}

	logger.Info("Filtered problems by difficulty",
		zap.String("difficulty", difficulty),
		zap.Int("filteredCount", len(filtered)),
	)

	return filtered, nil
}
