package repository

import (
	"context"

	"avito-test-quest/internal/logger"

	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"
)

type PrRepository struct {
	db   DB
	psql sq.StatementBuilderType
}

// NewPrRepository создает новый экземпляр PrRepository
func NewPrRepository(db DB) Repository {
	return &PrRepository{
		db:   db,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// ==================== Team Repository Methods ====================

// CreateTeam создает новую команду
func (r *PrRepository) CreateTeam(ctx context.Context, teamName string) (*TeamModel, error) {
	sql, args, err := r.psql.Insert("teams").Columns("team_name").Values(teamName).
		Suffix("RETURNING id, team_name, created_at, updated_at").ToSql()
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to build sql for CreateTeam", zap.Error(err))
		return nil, err
	}
	var tm TeamModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&tm.ID, &tm.TeamName, &tm.CreatedAt, &tm.UpdatedAt); err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to scan CreateTeam", zap.Error(err))
		return nil, err
	}
	return &tm, nil
}

// GetTeamByName получает команду по имени
func (r *PrRepository) GetTeamByName(ctx context.Context, teamName string) (*TeamModel, error) {
	sql, args, err := r.psql.Select("id", "team_name", "created_at", "updated_at").From("teams").Where(sq.Eq{"team_name": teamName}).ToSql()
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to build sql for GetTeamByName", zap.Error(err))
		return nil, err
	}
	var tm TeamModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&tm.ID, &tm.TeamName, &tm.CreatedAt, &tm.UpdatedAt); err != nil {
		return nil, err
	}
	return &tm, nil
}

// GetTeamByID получает команду по ID
func (r *PrRepository) GetTeamByID(ctx context.Context, teamID int64) (*TeamModel, error) {
	sql, args, err := r.psql.Select("id", "team_name", "created_at", "updated_at").From("teams").Where(sq.Eq{"id": teamID}).ToSql()
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to build sql for GetTeamByID", zap.Error(err))
		return nil, err
	}
	var tm TeamModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&tm.ID, &tm.TeamName, &tm.CreatedAt, &tm.UpdatedAt); err != nil {
		return nil, err
	}
	return &tm, nil
}

// TeamExists проверяет существование команды по имени
func (r *PrRepository) TeamExists(ctx context.Context, teamName string) (bool, error) {
	sql, args, err := r.psql.Select("count(1)").From("teams").Where(sq.Eq{"team_name": teamName}).ToSql()
	if err != nil {
		return false, err
	}
	var cnt int
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// ==================== User Repository Methods ====================

// CreateUser создает нового пользователя
func (r *PrRepository) CreateUser(ctx context.Context, userID, username string, teamID int64, isActive bool) (*UserModel, error) {
	sql, args, err := r.psql.Insert("users").Columns("user_id", "username", "team_id", "is_active").Values(userID, username, teamID, isActive).
		Suffix("RETURNING id, user_id, username, team_id, is_active, created_at, updated_at").ToSql()
	if err != nil {
		return nil, err
	}
	var u UserModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateUser обновляет данные пользователя
func (r *PrRepository) UpdateUser(ctx context.Context, userID, username string, teamID int64, isActive bool) (*UserModel, error) {
	sql, args, err := r.psql.Update("users").Set("username", username).Set("team_id", teamID).Set("is_active", isActive).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).Where(sq.Eq{"user_id": userID}).Suffix("RETURNING id, user_id, username, team_id, is_active, created_at, updated_at").ToSql()
	if err != nil {
		return nil, err
	}
	var u UserModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByID получает пользователя по userID
func (r *PrRepository) GetUserByID(ctx context.Context, userID string) (*UserModel, error) {
	sql, args, err := r.psql.Select("id", "user_id", "username", "team_id", "is_active", "created_at", "updated_at").From("users").Where(sq.Eq{"user_id": userID}).ToSql()
	if err != nil {
		return nil, err
	}
	var u UserModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserWithTeam получает пользователя с названием команды
func (r *PrRepository) GetUserWithTeam(ctx context.Context, userID string) (*UserWithTeam, error) {
	sb := r.psql.Select("u.user_id", "u.username", "t.team_name", "u.is_active").From("users u").Join("teams t ON u.team_id = t.id").Where(sq.Eq{"u.user_id": userID})
	sql, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var uwt UserWithTeam
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&uwt.UserID, &uwt.Username, &uwt.TeamName, &uwt.IsActive); err != nil {
		return nil, err
	}
	return &uwt, nil
}

// GetUsersByTeamID получает всех пользователей команды
func (r *PrRepository) GetUsersByTeamID(ctx context.Context, teamID int64) ([]UserModel, error) {
	sql, args, err := r.psql.Select("id", "user_id", "username", "team_id", "is_active", "created_at", "updated_at").From("users").Where(sq.Eq{"team_id": teamID}).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []UserModel
	for rows.Next() {
		var u UserModel
		if err := rows.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

// SetIsActive обновляет флаг активности пользователя
func (r *PrRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*UserModel, error) {
	sql, args, err := r.psql.Update("users").Set("is_active", isActive).Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).Where(sq.Eq{"user_id": userID}).Suffix("RETURNING id, user_id, username, team_id, is_active, created_at, updated_at").ToSql()
	if err != nil {
		return nil, err
	}
	var u UserModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// UserExists проверяет существование пользователя
func (r *PrRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	sql, args, err := r.psql.Select("count(1)").From("users").Where(sq.Eq{"user_id": userID}).ToSql()
	if err != nil {
		return false, err
	}
	var cnt int
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// GetActiveUsersInTeam получает активных пользователей команды
func (r *PrRepository) GetActiveUsersInTeam(ctx context.Context, teamID int64) ([]UserModel, error) {
	sql, args, err := r.psql.Select("id", "user_id", "username", "team_id", "is_active", "created_at", "updated_at").From("users").Where(sq.Eq{"team_id": teamID, "is_active": true}).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []UserModel
	for rows.Next() {
		var u UserModel
		if err := rows.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

// ==================== Pull Request Repository Methods ====================

// CreatePullRequest создает новый Pull Request
func (r *PrRepository) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*PullRequestModel, error) {
	sql, args, err := r.psql.Insert("pull_requests").Columns("pull_request_id", "pull_request_name", "author_id").Values(prID, prName, authorID).
		Suffix("RETURNING id, pull_request_id, pull_request_name, author_id, status, created_at, merged_at, updated_at").ToSql()
	if err != nil {
		return nil, err
	}
	var pr PullRequestModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&pr.ID, &pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
		return nil, err
	}

	return &pr, nil
}

// GetPullRequestByID получает Pull Request по его ID
func (r *PrRepository) GetPullRequestByID(ctx context.Context, prID string) (*PullRequestModel, error) {
	sql, args, err := r.psql.Select("id", "pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at", "updated_at").From("pull_requests").Where(sq.Eq{"pull_request_id": prID}).ToSql()
	if err != nil {
		return nil, err
	}
	var pr PullRequestModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&pr.ID, &pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
		return nil, err
	}

	return &pr, nil
}

// GetPullRequestWithReviewers получает Pull Request вместе с назначенными ревьюверами
func (r *PrRepository) GetPullRequestWithReviewers(ctx context.Context, prID string) (*PRWithReviewers, error) {
	pr, err := r.GetPullRequestByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	reviewers, err := r.GetReviewersByPRID(ctx, prID)
	if err != nil {
		return nil, err
	}

	return &PRWithReviewers{PullRequest: pr, Reviewers: reviewers}, nil
}

// MergePullRequest помечает Pull Request как замерженный
func (r *PrRepository) MergePullRequest(ctx context.Context, prID string) (*PullRequestModel, error) {
	sql, args, err := r.psql.Update("pull_requests").Set("status", "MERGED").Set("merged_at", sq.Expr("CURRENT_TIMESTAMP")).Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).Where(sq.Eq{"pull_request_id": prID}).Suffix("RETURNING id, pull_request_id, pull_request_name, author_id, status, created_at, merged_at, updated_at").ToSql()
	if err != nil {
		return nil, err
	}
	var pr PullRequestModel
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&pr.ID, &pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
		return nil, err
	}

	return &pr, nil
}

// PullRequestExists проверяет существование Pull Request по его ID
func (r *PrRepository) PullRequestExists(ctx context.Context, prID string) (bool, error) {
	sql, args, err := r.psql.Select("count(1)").From("pull_requests").Where(sq.Eq{"pull_request_id": prID}).ToSql()
	if err != nil {
		return false, err
	}
	var cnt int
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&cnt); err != nil {
		return false, err
	}

	return cnt > 0, nil
}

// GetPullRequestsByAuthor получает все Pull Request, созданные автором
func (r *PrRepository) GetPullRequestsByAuthor(ctx context.Context, authorID string) ([]PullRequestModel, error) {
	sql, args, err := r.psql.Select("id", "pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at", "updated_at").From("pull_requests").Where(sq.Eq{"author_id": authorID}).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []PullRequestModel
	for rows.Next() {
		var pr PullRequestModel
		if err := rows.Scan(&pr.ID, &pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, pr)
	}

	return res, nil
}

// ==================== PR Reviewer Repository Methods ====================

// AssignReviewer назначает ревьювера на Pull Request
func (r *PrRepository) AssignReviewer(ctx context.Context, prID, reviewerUserID string) error {
	sql, args, err := r.psql.Insert("pr_reviewers").Columns("pull_request_id", "reviewer_user_id").Values(prID, reviewerUserID).Suffix("ON CONFLICT DO NOTHING").ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

// RemoveReviewer удаляет ревьювера с Pull Request
func (r *PrRepository) RemoveReviewer(ctx context.Context, prID, reviewerUserID string) error {
	sql, args, err := r.psql.Delete("pr_reviewers").Where(sq.Eq{"pull_request_id": prID, "reviewer_user_id": reviewerUserID}).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)

	return err
}

// GetReviewersByPRID получает всех ревьюверов Pull Request по его ID
func (r *PrRepository) GetReviewersByPRID(ctx context.Context, prID string) ([]string, error) {
	sql, args, err := r.psql.Select("reviewer_user_id").From("pr_reviewers").Where(sq.Eq{"pull_request_id": prID}).OrderBy("assigned_at").ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		res = append(res, uid)
	}

	return res, nil
}

// GetPRsByReviewerID получает все Pull Request, где пользователь назначен ревьювером
func (r *PrRepository) GetPRsByReviewerID(ctx context.Context, reviewerUserID string) ([]PullRequestModel, error) {
	sb := r.psql.Select("p.id", "p.pull_request_id", "p.pull_request_name", "p.author_id", "p.status", "p.created_at", "p.merged_at", "p.updated_at").From("pull_requests p").Join("pr_reviewers r ON p.pull_request_id = r.pull_request_id").Where(sq.Eq{"r.reviewer_user_id": reviewerUserID})
	sql, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []PullRequestModel
	for rows.Next() {
		var pr PullRequestModel
		if err := rows.Scan(&pr.ID, &pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, pr)
	}

	return res, nil
}

// IsReviewerAssigned проверяет, назначен ли ревьювер на Pull Request
func (r *PrRepository) IsReviewerAssigned(ctx context.Context, prID, reviewerUserID string) (bool, error) {
	sql, args, err := r.psql.Select("count(1)").From("pr_reviewers").Where(sq.Eq{"pull_request_id": prID, "reviewer_user_id": reviewerUserID}).ToSql()
	if err != nil {
		return false, err
	}
	var cnt int
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&cnt); err != nil {
		return false, err
	}

	return cnt > 0, nil
}

// ReplaceReviewer заменяет одного ревьювера на другого для заданного Pull Request
func (r *PrRepository) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	// простая реализация: удалить старого и назначить нового.
	if err := r.RemoveReviewer(ctx, prID, oldReviewerID); err != nil {
		return err
	}
	if err := r.AssignReviewer(ctx, prID, newReviewerID); err != nil {
		// попытаться откатить удаление старого в случае ошибки назначения
		_ = r.AssignReviewer(ctx, prID, oldReviewerID)
		return err
	}

	return nil
}

// CountReviewersByPRID подсчитывает количество ревьюверов, назначенных на Pull Request
func (r *PrRepository) CountReviewersByPRID(ctx context.Context, prID string) (int, error) {
	sql, args, err := r.psql.Select("count(1)").From("pr_reviewers").Where(sq.Eq{"pull_request_id": prID}).ToSql()
	if err != nil {
		return 0, err
	}
	var cnt int
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&cnt); err != nil {
		return 0, err
	}

	return cnt, nil
}

// ==================== Stats Repository Methods ====================

// GetReviewerStats получает статистику по ревьюверам (кол-во назначений)
func (r *PrRepository) GetReviewerStats(ctx context.Context) ([]ReviewerStatRow, error) {
	sql, args, err := r.psql.Select("u.user_id", "u.username", "COUNT(pr.id) as assigned_count").
		From("users u").
		LeftJoin("pr_reviewers pr ON u.user_id = pr.reviewer_user_id").
		GroupBy("u.user_id", "u.username").
		OrderBy("assigned_count DESC").
		ToSql()
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to build sql for GetReviewerStats", zap.Error(err))
		return nil, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to query reviewer stats", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var stats []ReviewerStatRow
	for rows.Next() {
		var stat ReviewerStatRow
		if err := rows.Scan(&stat.UserID, &stat.Username, &stat.AssignedCount); err != nil {
			logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to scan reviewer stat", zap.Error(err))
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// GetPRStats получает статистику по Pull Requests (кол-во назначенных ревьюверов)
func (r *PrRepository) GetPRStats(ctx context.Context) ([]PRStatRow, error) {
	sql, args, err := r.psql.Select("p.pull_request_id", "p.pull_request_name", "p.author_id", "p.status", "COUNT(pr.id) as reviewer_count").
		From("pull_requests p").
		LeftJoin("pr_reviewers pr ON p.pull_request_id = pr.pull_request_id").
		GroupBy("p.id", "p.pull_request_id", "p.pull_request_name", "p.author_id", "p.status").
		OrderBy("reviewer_count DESC").
		ToSql()
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to build sql for GetPRStats", zap.Error(err))
		return nil, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to query pr stats", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var stats []PRStatRow
	for rows.Next() {
		var stat PRStatRow
		if err := rows.Scan(&stat.PullRequestID, &stat.PullRequestName, &stat.AuthorID, &stat.Status, &stat.ReviewerCount); err != nil {
			logger.GetOrCreateLoggerFromCtx(ctx).Error(ctx, "failed to scan pr stat", zap.Error(err))
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// проверка реализации интерфейса Repository
var _ Repository = (*PrRepository)(nil)
