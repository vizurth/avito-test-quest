package service

import (
	"avito-test-quest/internal/models"
	"context"
)

// TeamService интерфейс для работы с командами
type TeamService interface {
	// CreateTeam создает команду с участниками
	// Возвращает созданную команду или ошибку TEAM_EXISTS
	CreateTeam(ctx context.Context, input models.CreateTeamInput) (*models.Team, error)

	// GetTeam получает команду по имени
	// Возвращает команду или ошибку NOT_FOUND
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
}

// UserService интерфейс для работы с пользователями
type UserService interface {
	// SetIsActive устанавливает флаг активности пользователя
	// Возвращает обновленного пользователя или ошибку NOT_FOUND
	SetIsActive(ctx context.Context, input models.SetIsActiveInput) (*models.User, error)

	// GetUserReviews получает список PR, где пользователь назначен ревьювером
	// Возвращает список PR или ошибку NOT_FOUND
	GetUserReviews(ctx context.Context, userID string) (*models.UserReviewsOutput, error)
}

// PullRequestService интерфейс для работы с Pull Request
type PullRequestService interface {
	// CreatePullRequest создает PR и назначает до 2 ревьюверов
	// Возвращает созданный PR или ошибки: NOT_FOUND, PR_EXISTS
	CreatePullRequest(ctx context.Context, input models.CreatePullRequestInput) (*models.PullRequest, error)

	// MergePullRequest помечает PR как MERGED (идемпотентно)
	// Возвращает обновленный PR или ошибку NOT_FOUND
	MergePullRequest(ctx context.Context, input models.MergePullRequestInput) (*models.PullRequest, error)

	// ReassignReviewer переназначает ревьювера на другого из команды
	// Возвращает обновленный PR и ID нового ревьювера
	// Ошибки: NOT_FOUND, PR_MERGED, NOT_ASSIGNED, NO_CANDIDATE
	ReassignReviewer(ctx context.Context, input models.ReassignReviewerInput) (*models.ReassignReviewerOutput, error)
}

// StatsService интерфейс для получения статистики
type StatsService interface {
	// GetStats получает статистику по ревьюверам и PR
	GetStats(ctx context.Context) (*models.StatsOutput, error)
}

// Service объединяет все сервисные интерфейсы
type Service interface {
	TeamService
	UserService
	PullRequestService
	StatsService
}
