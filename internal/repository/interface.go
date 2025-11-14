package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// TeamModel представляет команду в БД
type TeamModel struct {
	ID        int64     `db:"id"`
	TeamName  string    `db:"team_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// UserModel представляет пользователя в БД
type UserModel struct {
	ID        int64     `db:"id"`
	UserID    string    `db:"user_id"`
	Username  string    `db:"username"`
	TeamID    int64     `db:"team_id"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// PullRequestModel представляет PR в БД
type PullRequestModel struct {
	ID              int64      `db:"id"`
	PullRequestID   string     `db:"pull_request_id"`
	PullRequestName string     `db:"pull_request_name"`
	AuthorID        string     `db:"author_id"`
	Status          string     `db:"status"` // OPEN, MERGED
	CreatedAt       time.Time  `db:"created_at"`
	MergedAt        *time.Time `db:"merged_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

// PRReviewerModel представляет связь PR и ревьювера в БД
type PRReviewerModel struct {
	ID             int64     `db:"id"`
	PullRequestID  string    `db:"pull_request_id"`
	ReviewerUserID string    `db:"reviewer_user_id"`
	AssignedAt     time.Time `db:"assigned_at"`
}

// ReviewerAssignmentHistoryModel представляет историю переназначений
type ReviewerAssignmentHistoryModel struct {
	ID                int64     `db:"id"`
	PullRequestID     string    `db:"pull_request_id"`
	OldReviewerUserID string    `db:"old_reviewer_user_id"`
	NewReviewerUserID string    `db:"new_reviewer_user_id"`
	ReassignedAt      time.Time `db:"reassigned_at"`
	Reason            *string   `db:"reason"`
}

// UserWithTeam расширенная модель пользователя с названием команды
type UserWithTeam struct {
	UserID   string `db:"user_id"`
	Username string `db:"username"`
	TeamName string `db:"team_name"`
	IsActive bool   `db:"is_active"`
}

// PRWithReviewers PR с назначенными ревьюверами
type PRWithReviewers struct {
	PullRequest *PullRequestModel
	Reviewers   []string // список user_id ревьюверов
}

// TeamRepository интерфейс для работы с командами
type TeamRepository interface {
	// CreateTeam создает новую команду
	CreateTeam(ctx context.Context, teamName string) (*TeamModel, error)

	// GetTeamByName получает команду по имени
	GetTeamByName(ctx context.Context, teamName string) (*TeamModel, error)

	// GetTeamByID получает команду по ID
	GetTeamByID(ctx context.Context, teamID int64) (*TeamModel, error)

	// TeamExists проверяет существование команды по имени
	TeamExists(ctx context.Context, teamName string) (bool, error)
}

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	// CreateUser создает нового пользователя
	CreateUser(ctx context.Context, userID, username string, teamID int64, isActive bool) (*UserModel, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, userID, username string, teamID int64, isActive bool) (*UserModel, error)

	// GetUserByID получает пользователя по user_id
	GetUserByID(ctx context.Context, userID string) (*UserModel, error)

	// GetUserWithTeam получает пользователя с названием команды
	GetUserWithTeam(ctx context.Context, userID string) (*UserWithTeam, error)

	// GetUsersByTeamID получает всех пользователей команды
	GetUsersByTeamID(ctx context.Context, teamID int64) ([]UserModel, error)

	// SetIsActive обновляет флаг активности пользователя
	SetIsActive(ctx context.Context, userID string, isActive bool) (*UserModel, error)

	// UserExists проверяет существование пользователя
	UserExists(ctx context.Context, userID string) (bool, error)

	// GetActiveUsersInTeam получает активных пользователей команды
	GetActiveUsersInTeam(ctx context.Context, teamID int64) ([]UserModel, error)
}

// PullRequestRepository интерфейс для работы с Pull Request
type PullRequestRepository interface {
	// CreatePullRequest создает новый PR
	CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*PullRequestModel, error)

	// GetPullRequestByID получает PR по pull_request_id
	GetPullRequestByID(ctx context.Context, prID string) (*PullRequestModel, error)

	// GetPullRequestWithReviewers получает PR со списком ревьюверов
	GetPullRequestWithReviewers(ctx context.Context, prID string) (*PRWithReviewers, error)

	// MergePullRequest обновляет статус PR на MERGED
	MergePullRequest(ctx context.Context, prID string) (*PullRequestModel, error)

	// PullRequestExists проверяет существование PR
	PullRequestExists(ctx context.Context, prID string) (bool, error)

	// GetPullRequestsByAuthor получает все PR автора
	GetPullRequestsByAuthor(ctx context.Context, authorID string) ([]PullRequestModel, error)
}

// PRReviewerRepository интерфейс для работы с ревьюверами PR
type PRReviewerRepository interface {
	// AssignReviewer назначает ревьювера на PR
	AssignReviewer(ctx context.Context, prID, reviewerUserID string) error

	// RemoveReviewer удаляет ревьювера с PR
	RemoveReviewer(ctx context.Context, prID, reviewerUserID string) error

	// GetReviewersByPRID получает всех ревьюверов PR
	GetReviewersByPRID(ctx context.Context, prID string) ([]string, error)

	// GetPRsByReviewerID получает все PR, где пользователь назначен ревьювером
	GetPRsByReviewerID(ctx context.Context, reviewerUserID string) ([]PullRequestModel, error)

	// IsReviewerAssigned проверяет, назначен ли пользователь ревьювером на PR
	IsReviewerAssigned(ctx context.Context, prID, reviewerUserID string) (bool, error)

	// ReplaceReviewer заменяет одного ревьювера на другого
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error

	// CountReviewersByPRID подсчитывает количество ревьюверов на PR
	CountReviewersByPRID(ctx context.Context, prID string) (int, error)
}

// Repository объединяет все репозиторные интерфейсы
type Repository interface {
	TeamRepository
	UserRepository
	PullRequestRepository
	PRReviewerRepository
}
