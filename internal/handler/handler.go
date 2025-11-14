package handler

import (
	"net/http"

	"avito-test-quest/internal/logger"
	"avito-test-quest/internal/models"
	"avito-test-quest/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PrHandler struct {
	service service.Service
	router  *gin.Engine
}

// NewPrHandler создает новый экземпляр PrHandler
func NewPrHandler(service service.Service, router *gin.Engine) AllHandlers {
	return &PrHandler{
		service: service,
		router:  router,
	}
}

// InitRoutes инициализирует все роуты приложения
func (h *PrHandler) InitRoutes() {
	// Health check
	h.router.GET("/health", h.HealthCheck)

	// ручки Teams
	teamGroup := h.router.Group("/team")
	{
		teamGroup.POST("/add", h.CreateTeam)
		teamGroup.GET("/get", h.GetTeam)
	}

	// ручки Users
	usersGroup := h.router.Group("/users")
	{
		usersGroup.POST("/setIsActive", h.SetIsActive) // только для админов (если будет аутентификация)
		usersGroup.GET("/getReview", h.GetUserReviews)
	}

	// ручки Pull Requests
	prGroup := h.router.Group("/pullRequest")
	{
		prGroup.POST("/create", h.CreatePullRequest)  // только для админов (если будет аутентификация)
		prGroup.POST("/merge", h.MergePullRequest)    // только для админов (если будет аутентификация)
		prGroup.POST("/reassign", h.ReassignReviewer) // только для админов (если будет аутентификация)
	}

	// Stats endpoint
	h.router.GET("/stats", h.GetStats)
}

// ==================== Team Handlers ====================

// CreateTeam создает команду с участниками
func (h *PrHandler) CreateTeam(c *gin.Context) {
	var input models.CreateTeamInput
	ctx := c.Request.Context()

	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid create team request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// создаем команду
	tm, err := h.service.CreateTeam(ctx, input)
	if err != nil {
		switch err.Error() {
		case "TEAM_EXISTS":
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "TEAM_EXISTS", "message": "team_name already exists"}})
			return
		default:
			log.Error(ctx, "create team failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"team": tm})
}

// GetTeam получает команду по имени
func (h *PrHandler) GetTeam(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "team_name required"}})
		return
	}
	// получаем команду
	tm, err := h.service.GetTeam(ctx, teamName)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "team not found"}})
			return
		}
		log.Error(ctx, "get team failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tm)
}

// ==================== User Handlers ====================

// SetIsActive устанавливает флаг активности пользователя
func (h *PrHandler) SetIsActive(c *gin.Context) {
	var input models.SetIsActiveInput
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid setIsActive request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// устанавливаем is_active
	u, err := h.service.SetIsActive(ctx, input)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "user not found"}})
			return
		}
		log.Error(ctx, "set is_active failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": u})
}

// GetUserReviews получает список PR, где пользователь назначен ревьювером
func (h *PrHandler) GetUserReviews(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "user_id required"}})
		return
	}
	// получаем список PR ревьювера
	out, err := h.service.GetUserReviews(ctx, userID)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "user not found"}})
			return
		}
		log.Error(ctx, "get user reviews failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}

// ==================== Pull Request Handlers ====================

// CreatePullRequest создает PR и назначает до 2 ревьюверов
func (h *PrHandler) CreatePullRequest(c *gin.Context) {
	var input models.CreatePullRequestInput
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid create pr request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// создаем PR
	pr, err := h.service.CreatePullRequest(ctx, input)
	if err != nil {
		switch err.Error() {
		case "NOT_FOUND":
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "author or team not found"}})
			return
		case "PR_EXISTS":
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "PR_EXISTS", "message": "PR id already exists"}})
			return
		default:
			log.Error(ctx, "create pr failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"pr": pr})
}

// MergePullRequest помечает PR как MERGED (идемпотентно)
func (h *PrHandler) MergePullRequest(c *gin.Context) {
	var input models.MergePullRequestInput
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid merge pr request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// мержим PR
	pr, err := h.service.MergePullRequest(ctx, input)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "pr not found"}})
			return
		}
		log.Error(ctx, "merge pr failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

// ReassignReviewer переназначает ревьювера на другого из команды
func (h *PrHandler) ReassignReviewer(c *gin.Context) {
	var input models.ReassignReviewerInput
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid reassign request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// переназначаем ревьювера
	out, err := h.service.ReassignReviewer(ctx, input)
	if err != nil {
		switch err.Error() {
		case "NOT_FOUND":
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "pr or user not found"}})
			return
		case "PR_MERGED":
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "PR_MERGED", "message": "cannot reassign on merged PR"}})
			return
		case "NOT_ASSIGNED":
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "NOT_ASSIGNED", "message": "reviewer is not assigned to this PR"}})
			return
		case "NO_CANDIDATE":
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "NO_CANDIDATE", "message": "no active replacement candidate in team"}})
			return
		default:
			log.Error(ctx, "reassign failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"pr": out.PR, "replaced_by": out.ReplacedBy})
}

// ==================== Health Handler ====================

// HealthCheck проверка здоровья сервиса
func (h *PrHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ==================== Stats Handler ====================

// GetStats получить статистику по ревьюверам и PR
func (h *PrHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)

	stats, err := h.service.GetStats(ctx)
	if err != nil {
		log.Error(ctx, "failed to get stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// проверка реализации интерфейса AllHandlers
var _ AllHandlers = (*PrHandler)(nil)
