package handler

import (
	"github.com/gin-gonic/gin"
)

// Handler основной интерфейс для всех handlers
type Handler interface {
	// InitRoutes инициализирует все роуты
	InitRoutes()
}

// TeamHandler интерфейс для работы с командами
type TeamHandler interface {
	// CreateTeam POST /team/add
	// Создать команду с участниками (создаёт/обновляет пользователей)
	CreateTeam(c *gin.Context)

	// GetTeam GET /team/get
	// Получить команду с участниками по team_name (query param)
	GetTeam(c *gin.Context)
}

// UserHandler интерфейс для работы с пользователями
type UserHandler interface {
	// SetIsActive POST /users/setIsActive
	// Установить флаг активности пользователя (требует Admin токен)
	SetIsActive(c *gin.Context)

	// GetUserReviews GET /users/getReview
	// Получить PR'ы, где пользователь назначен ревьювером (query param: user_id)
	GetUserReviews(c *gin.Context)
}

// PullRequestHandler интерфейс для работы с Pull Request'ами
type PullRequestHandler interface {
	// CreatePullRequest POST /pullRequest/create
	// Создать PR и автоматически назначить до 2 ревьюверов из команды автора
	CreatePullRequest(c *gin.Context)

	// MergePullRequest POST /pullRequest/merge
	// Пометить PR как MERGED (идемпотентная операция)
	MergePullRequest(c *gin.Context)

	// ReassignReviewer POST /pullRequest/reassign
	// Переназначить конкретного ревьювера на другого из его команды
	ReassignReviewer(c *gin.Context)
}

// HealthHandler интерфейс для health check
type HealthHandler interface {
	// HealthCheck GET /health
	// Проверка состояния сервиса
	HealthCheck(c *gin.Context)
}

// AllHandlers объединяет все handler интерфейсы
type AllHandlers interface {
	Handler
	TeamHandler
	UserHandler
	PullRequestHandler
	HealthHandler
}
