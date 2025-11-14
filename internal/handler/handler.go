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

	// Teams endpoints
	teamGroup := h.router.Group("/team")
	{
		teamGroup.POST("/add", h.CreateTeam)
		teamGroup.GET("/get", h.GetTeam)
	}

	// Users endpoints
	usersGroup := h.router.Group("/users")
	{
		usersGroup.POST("/setIsActive", h.SetIsActive) // Admin only
		usersGroup.GET("/getReview", h.GetUserReviews)
	}

	// Pull Requests endpoints
	prGroup := h.router.Group("/pullRequest")
	{
		prGroup.POST("/create", h.CreatePullRequest)  // Admin only
		prGroup.POST("/merge", h.MergePullRequest)    // Admin only
		prGroup.POST("/reassign", h.ReassignReviewer) // Admin only
	}
}

// ==================== Team Handlers ====================

func (h *PrHandler) CreateTeam(c *gin.Context) {
	var input models.CreateTeamInput
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid create team request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

func (h *PrHandler) GetTeam(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "team_name required"}})
		return
	}
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

func (h *PrHandler) SetIsActive(c *gin.Context) {
	var input models.SetIsActiveInput
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid setIsActive request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

func (h *PrHandler) GetUserReviews(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "user_id required"}})
		return
	}
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

func (h *PrHandler) CreatePullRequest(c *gin.Context) {
	var input models.CreatePullRequestInput
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid create pr request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

func (h *PrHandler) MergePullRequest(c *gin.Context) {
	var input models.MergePullRequestInput
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid merge pr request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

func (h *PrHandler) ReassignReviewer(c *gin.Context) {
	var input models.ReassignReviewerInput
	ctx := c.Request.Context()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Warn(ctx, "invalid reassign request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

func (h *PrHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Compile-time check that PrHandler implements AllHandlers
var _ AllHandlers = (*PrHandler)(nil)
