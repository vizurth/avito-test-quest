package service

import (
	"avito-test-quest/internal/logger"
	"avito-test-quest/internal/models"
	"avito-test-quest/internal/repository"
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
)

type PrService struct {
	repo repository.Repository
}

// NewPrService создает новый экземпляр PrService
func NewPrService(repo repository.Repository) Service {
	return &PrService{
		repo: repo,
	}
}

// ==================== Team Service Methods ====================

func (s *PrService) CreateTeam(ctx context.Context, input models.CreateTeamInput) (*models.Team, error) {
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	exists, err := s.repo.TeamExists(ctx, input.TeamName)
	if err != nil {
		log.Error(ctx, "failed to check team exists", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, errors.New("TEAM_EXISTS")
	}
	teamModel, err := s.repo.CreateTeam(ctx, input.TeamName)
	if err != nil {
		log.Error(ctx, "failed to create team", zap.Error(err))
		return nil, err
	}
	// create or update users
	for _, m := range input.Members {
		ue, err := s.repo.UserExists(ctx, m.UserID)
		if err != nil {
			log.Error(ctx, "failed to check user exists", zap.Error(err))
			return nil, err
		}
		if ue {
			if _, err := s.repo.UpdateUser(ctx, m.UserID, m.Username, teamModel.ID, m.IsActive); err != nil {
				log.Error(ctx, "failed to update user", zap.Error(err))
				return nil, err
			}
		} else {
			if _, err := s.repo.CreateUser(ctx, m.UserID, m.Username, teamModel.ID, m.IsActive); err != nil {
				log.Error(ctx, "failed to create user", zap.Error(err))
				return nil, err
			}
		}
	}
	users, err := s.repo.GetUsersByTeamID(ctx, teamModel.ID)
	if err != nil {
		log.Error(ctx, "failed to fetch team users", zap.Error(err))
		return nil, err
	}
	var members []models.TeamMember
	for _, u := range users {
		members = append(members, models.TeamMember{UserID: u.UserID, Username: u.Username, IsActive: u.IsActive})
	}
	return &models.Team{TeamName: teamModel.TeamName, Members: members}, nil
}

func (s *PrService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	tm, err := s.repo.GetTeamByName(ctx, teamName)
	if err != nil {
		log.Info(ctx, "team not found", zap.String("team", teamName), zap.Error(err))
		return nil, errors.New("NOT_FOUND")
	}
	users, err := s.repo.GetUsersByTeamID(ctx, tm.ID)
	if err != nil {
		log.Error(ctx, "failed to get users by team id", zap.Error(err))
		return nil, err
	}
	var members []models.TeamMember
	for _, u := range users {
		members = append(members, models.TeamMember{UserID: u.UserID, Username: u.Username, IsActive: u.IsActive})
	}
	return &models.Team{TeamName: tm.TeamName, Members: members}, nil
}

// ==================== User Service Methods ====================

func (s *PrService) SetIsActive(ctx context.Context, input models.SetIsActiveInput) (*models.User, error) {
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	exists, err := s.repo.UserExists(ctx, input.UserID)
	if err != nil {
		log.Error(ctx, "failed to check user exists", zap.Error(err))
		return nil, err
	}
	if !exists {
		return nil, errors.New("NOT_FOUND")
	}
	u, err := s.repo.SetIsActive(ctx, input.UserID, input.IsActive)
	if err != nil {
		log.Error(ctx, "failed to set is_active", zap.Error(err))
		return nil, err
	}
	team, err := s.repo.GetTeamByID(ctx, u.TeamID)
	if err != nil {
		log.Error(ctx, "failed to get team for user", zap.Error(err))
		return nil, err
	}
	return &models.User{UserID: u.UserID, Username: u.Username, TeamName: team.TeamName, IsActive: u.IsActive}, nil
}

func (s *PrService) GetUserReviews(ctx context.Context, userID string) (*models.UserReviewsOutput, error) {
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	exists, err := s.repo.UserExists(ctx, userID)
	if err != nil {
		log.Error(ctx, "failed to check user exists", zap.Error(err))
		return nil, err
	}
	if !exists {
		return nil, errors.New("NOT_FOUND")
	}
	prs, err := s.repo.GetPRsByReviewerID(ctx, userID)
	if err != nil {
		log.Error(ctx, "failed to get prs by reviewer", zap.Error(err))
		return nil, err
	}
	var out []models.PullRequestShort
	for _, p := range prs {
		out = append(out, models.PullRequestShort{PullRequestID: p.PullRequestID, PullRequestName: p.PullRequestName, AuthorID: p.AuthorID, Status: p.Status})
	}
	return &models.UserReviewsOutput{UserID: userID, PullRequests: out}, nil
}

// ==================== Pull Request Service Methods ====================

func (s *PrService) CreatePullRequest(ctx context.Context, input models.CreatePullRequestInput) (*models.PullRequest, error) {
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	author, err := s.repo.GetUserByID(ctx, input.AuthorID)
	if err != nil {
		log.Info(ctx, "author not found", zap.String("author", input.AuthorID), zap.Error(err))
		return nil, errors.New("NOT_FOUND")
	}
	exists, err := s.repo.PullRequestExists(ctx, input.PullRequestID)
	if err != nil {
		log.Error(ctx, "failed to check pr exists", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, errors.New("PR_EXISTS")
	}
	prModel, err := s.repo.CreatePullRequest(ctx, input.PullRequestID, input.PullRequestName, input.AuthorID)
	if err != nil {
		log.Error(ctx, "failed to create pr", zap.Error(err))
		return nil, err
	}
	// pick up to 2 active reviewers from team
	candidates, err := s.repo.GetActiveUsersInTeam(ctx, author.TeamID)
	if err != nil {
		log.Error(ctx, "failed to get active users in team", zap.Error(err))
		return nil, err
	}
	var assigned []string
	for _, c := range candidates {
		if c.UserID == input.AuthorID {
			continue
		}
		if len(assigned) >= 2 {
			break
		}
		if err := s.repo.AssignReviewer(ctx, input.PullRequestID, c.UserID); err != nil {
			log.Warn(ctx, "failed to assign reviewer", zap.String("user", c.UserID), zap.Error(err))
			continue
		}
		assigned = append(assigned, c.UserID)
	}
	// build output
	var createdAt *string
	if !prModel.CreatedAt.IsZero() {
		t := prModel.CreatedAt.UTC().Format(time.RFC3339)
		createdAt = &t
	}
	var mergedAt *string
	if prModel.MergedAt != nil {
		t := prModel.MergedAt.UTC().Format(time.RFC3339)
		mergedAt = &t
	}
	return &models.PullRequest{PullRequestID: prModel.PullRequestID, PullRequestName: prModel.PullRequestName, AuthorID: prModel.AuthorID, Status: prModel.Status, AssignedReviewers: assigned, CreatedAt: createdAt, MergedAt: mergedAt}, nil
}

func (s *PrService) MergePullRequest(ctx context.Context, input models.MergePullRequestInput) (*models.PullRequest, error) {
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	pr, err := s.repo.GetPullRequestByID(ctx, input.PullRequestID)
	if err != nil {
		log.Info(ctx, "pr not found", zap.String("pr", input.PullRequestID), zap.Error(err))
		return nil, errors.New("NOT_FOUND")
	}
	if pr.Status == "MERGED" {
		// build response
		reviewers, _ := s.repo.GetReviewersByPRID(ctx, input.PullRequestID)
		var createdAt *string
		if !pr.CreatedAt.IsZero() {
			t := pr.CreatedAt.UTC().Format(time.RFC3339)
			createdAt = &t
		}
		var mergedAt *string
		if pr.MergedAt != nil {
			t := pr.MergedAt.UTC().Format(time.RFC3339)
			mergedAt = &t
		}
		return &models.PullRequest{PullRequestID: pr.PullRequestID, PullRequestName: pr.PullRequestName, AuthorID: pr.AuthorID, Status: pr.Status, AssignedReviewers: reviewers, CreatedAt: createdAt, MergedAt: mergedAt}, nil
	}
	updated, err := s.repo.MergePullRequest(ctx, input.PullRequestID)
	if err != nil {
		log.Error(ctx, "failed to merge pr", zap.Error(err))
		return nil, err
	}
	reviewers, _ := s.repo.GetReviewersByPRID(ctx, input.PullRequestID)
	var createdAt *string
	if !updated.CreatedAt.IsZero() {
		t := updated.CreatedAt.UTC().Format(time.RFC3339)
		createdAt = &t
	}
	var mergedAt *string
	if updated.MergedAt != nil {
		t := updated.MergedAt.UTC().Format(time.RFC3339)
		mergedAt = &t
	}
	return &models.PullRequest{PullRequestID: updated.PullRequestID, PullRequestName: updated.PullRequestName, AuthorID: updated.AuthorID, Status: updated.Status, AssignedReviewers: reviewers, CreatedAt: createdAt, MergedAt: mergedAt}, nil
}

func (s *PrService) ReassignReviewer(ctx context.Context, input models.ReassignReviewerInput) (*models.ReassignReviewerOutput, error) {
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	prWith, err := s.repo.GetPullRequestWithReviewers(ctx, input.PullRequestID)
	if err != nil {
		log.Info(ctx, "pr not found for reassign", zap.String("pr", input.PullRequestID), zap.Error(err))
		return nil, errors.New("NOT_FOUND")
	}
	if prWith.PullRequest.Status == "MERGED" {
		return nil, errors.New("PR_MERGED")
	}
	assigned, err := s.repo.IsReviewerAssigned(ctx, input.PullRequestID, input.OldReviewerID)
	if err != nil {
		log.Error(ctx, "failed to check reviewer assigned", zap.Error(err))
		return nil, err
	}
	if !assigned {
		return nil, errors.New("NOT_ASSIGNED")
	}
	oldUser, err := s.repo.GetUserByID(ctx, input.OldReviewerID)
	if err != nil {
		log.Error(ctx, "failed to get old reviewer user", zap.Error(err))
		return nil, err
	}
	candidates, err := s.repo.GetActiveUsersInTeam(ctx, oldUser.TeamID)
	if err != nil {
		log.Error(ctx, "failed to fetch candidates", zap.Error(err))
		return nil, err
	}
	excluded := map[string]struct{}{}
	excluded[prWith.PullRequest.AuthorID] = struct{}{}
	for _, r := range prWith.Reviewers {
		excluded[r] = struct{}{}
	}
	// try find new candidate
	var chosen *string
	for _, c := range candidates {
		if _, ok := excluded[c.UserID]; ok {
			continue
		}
		chosen = &c.UserID
		break
	}
	if chosen == nil {
		return nil, errors.New("NO_CANDIDATE")
	}
	if err := s.repo.ReplaceReviewer(ctx, input.PullRequestID, input.OldReviewerID, *chosen); err != nil {
		log.Error(ctx, "failed to replace reviewer", zap.Error(err))
		return nil, err
	}
	// history persistence not implemented in repository interface/migrations
	updated, err := s.repo.GetPullRequestWithReviewers(ctx, input.PullRequestID)
	if err != nil {
		log.Error(ctx, "failed to fetch updated pr after reassign", zap.Error(err))
		return nil, err
	}
	var createdAt *string
	if !updated.PullRequest.CreatedAt.IsZero() {
		t := updated.PullRequest.CreatedAt.UTC().Format(time.RFC3339)
		createdAt = &t
	}
	var mergedAt *string
	if updated.PullRequest.MergedAt != nil {
		t := updated.PullRequest.MergedAt.UTC().Format(time.RFC3339)
		mergedAt = &t
	}
	outPR := &models.PullRequest{PullRequestID: updated.PullRequest.PullRequestID, PullRequestName: updated.PullRequest.PullRequestName, AuthorID: updated.PullRequest.AuthorID, Status: updated.PullRequest.Status, AssignedReviewers: updated.Reviewers, CreatedAt: createdAt, MergedAt: mergedAt}
	return &models.ReassignReviewerOutput{PR: outPR, ReplacedBy: *chosen}, nil
}

// Compile-time check that PrService implements Service
var _ Service = (*PrService)(nil)
