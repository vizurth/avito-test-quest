package models

// Team представляет команду с участниками
type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

// TeamMember представляет участника команды
type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// User представляет пользователя
type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// PullRequest представляет pull request с полной информацией
type PullRequest struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"` // OPEN, MERGED
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         *string  `json:"createdAt,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

// PullRequestShort представляет краткую информацию о PR
type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

// CreateTeamInput входные данные для создания команды
type CreateTeamInput struct {
	TeamName string       `json:"team_name" binding:"required"`
	Members  []TeamMember `json:"members" binding:"required,dive"`
}

// SetIsActiveInput входные данные для изменения статуса активности
type SetIsActiveInput struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// CreatePullRequestInput входные данные для создания PR
type CreatePullRequestInput struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

// MergePullRequestInput входные данные для мерджа PR
type MergePullRequestInput struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

// ReassignReviewerInput входные данные для переназначения ревьювера
type ReassignReviewerInput struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldReviewerID string `json:"old_user_id" binding:"required"`
}

// ReassignReviewerOutput результат переназначения ревьювера
type ReassignReviewerOutput struct {
	PR         *PullRequest `json:"pr"`
	ReplacedBy string       `json:"replaced_by"`
}

// UserReviewsOutput список PR для пользователя
type UserReviewsOutput struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

// ReviewerStat статистика ревьювера
type ReviewerStat struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	AssignedCount int    `json:"assigned_count"`
}

// PRStat статистика PR
type PRStat struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
	ReviewerCount   int    `json:"reviewer_count"`
}

// StatsOutput выходные данные статистики
type StatsOutput struct {
	ReviewerStats []ReviewerStat `json:"reviewer_stats"`
	PRStats       []PRStat       `json:"pr_stats"`
}
