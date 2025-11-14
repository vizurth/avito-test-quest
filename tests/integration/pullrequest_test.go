package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PullRequest struct {
	PullRequestID     string        `json:"pull_request_id"`
	PullRequestName   string        `json:"pull_request_name"`
	AuthorID          string        `json:"author_id"`
	AssignedReviewers []interface{} `json:"assigned_reviewers"`
}

type CreatePRResponse struct {
	PR PullRequest `json:"pr"`
}

func TestPullRequestEndpoints(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("CreatePR_SuccessWithTwoReviewers", func(t *testing.T) {
		cleanupTestData(t)

		// cоздаем команду с 3 активными пользователями
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		})

		payload := map[string]interface{}{
			"pull_request_id":   "pr-1",
			"pull_request_name": "Add feature",
			"author_id":         "u1",
		}

		// отправляем запрос на создание PR
		resp := makeRequest(t, "POST", "/pullRequest/create", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		pr := result["pr"].(map[string]interface{})
		assert.Equal(t, "pr-1", pr["pull_request_id"])
		assert.Equal(t, "Add feature", pr["pull_request_name"])
		assert.Equal(t, "u1", pr["author_id"])
		assert.Equal(t, "OPEN", pr["status"])

		reviewers := pr["assigned_reviewers"].([]interface{})
		assert.Len(t, reviewers, 2, "should assign 2 reviewers")

		for _, r := range reviewers {
			assert.NotEqual(t, "u1", r)
		}
	})

	t.Run("CreatePR_OneReviewerAvailable", func(t *testing.T) {
		cleanupTestData(t)

		// cоздаем команду с 2 пользователями (автор + 1 активный)
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		})

		payload := map[string]interface{}{
			"pull_request_id":   "pr-1",
			"pull_request_name": "Add feature",
			"author_id":         "u1",
		}

		// отправляем запрос на создание PR
		resp := makeRequest(t, "POST", "/pullRequest/create", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		pr := result["pr"].(map[string]interface{})
		reviewers := pr["assigned_reviewers"].([]interface{})
		assert.Len(t, reviewers, 1, "should assign only 1 reviewer")
		assert.Equal(t, "u2", reviewers[0])
	})

	t.Run("CreatePR_NoActiveReviewers", func(t *testing.T) {
		cleanupTestData(t)

		// cоздаем команду только с автором
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		})

		payload := map[string]interface{}{
			"pull_request_id":   "pr-1",
			"pull_request_name": "Add feature",
			"author_id":         "u1",
		}

		// отправляем запрос на создание PR
		resp := makeRequest(t, "POST", "/pullRequest/create", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result CreatePRResponse
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Len(t, result.PR.AssignedReviewers, 0, "should assign 0 reviewers")
	})

	t.Run("CreatePR_SkipInactiveUsers", func(t *testing.T) {
		cleanupTestData(t)

		// cоздаем команду с неактивными пользователями
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": false},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "David", "is_active": false},
		})

		payload := map[string]interface{}{
			"pull_request_id":   "pr-1",
			"pull_request_name": "Add feature",
			"author_id":         "u1",
		}

		// отправляем запрос на создание PR
		resp := makeRequest(t, "POST", "/pullRequest/create", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		pr := result["pr"].(map[string]interface{})
		reviewers := pr["assigned_reviewers"].([]interface{})

		// Должен быть назначен только u3 (единственный активный, кроме автора)
		assert.Len(t, reviewers, 1)
		assert.Equal(t, "u3", reviewers[0])
	})

	t.Run("CreatePR_DuplicateID", func(t *testing.T) {
		cleanupTestData(t)

		// cоздаем команду с одним пользователем
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		})

		payload := map[string]interface{}{
			"pull_request_id":   "pr-1",
			"pull_request_name": "Add feature",
			"author_id":         "u1",
		}

		// отправляем первый запрос на создание PR
		resp1 := makeRequest(t, "POST", "/pullRequest/create", payload, nil)
		resp1.Body.Close()
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		// отправляем запрос на создание второго PR с тем же ID
		resp2 := makeRequest(t, "POST", "/pullRequest/create", payload, nil)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusConflict, resp2.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "PR_EXISTS", errObj["code"])
	})

	t.Run("CreatePR_AuthorNotFound", func(t *testing.T) {
		cleanupTestData(t)

		payload := map[string]interface{}{
			"pull_request_id":   "pr-1",
			"pull_request_name": "Add feature",
			"author_id":         "nonexistent",
		}

		// отправляем запрос на создание PR
		resp := makeRequest(t, "POST", "/pullRequest/create", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "NOT_FOUND", errObj["code"])
	})

	t.Run("MergePR_Success", func(t *testing.T) {
		cleanupTestData(t)

		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		})

		createTestPR(t, "pr-1", "Test PR", "u1")

		payload := map[string]interface{}{
			"pull_request_id": "pr-1",
		}

		// отправляем запрос на мердж PR
		resp := makeRequest(t, "POST", "/pullRequest/merge", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		pr := result["pr"].(map[string]interface{})
		assert.Equal(t, "MERGED", pr["status"])
		assert.NotNil(t, pr["mergedAt"])
	})

	t.Run("MergePR_Idempotent", func(t *testing.T) {
		cleanupTestData(t)

		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		})

		createTestPR(t, "pr-1", "Test PR", "u1")

		payload := map[string]interface{}{
			"pull_request_id": "pr-1",
		}

		// отправляем первый запрос на мерж PR
		resp1 := makeRequest(t, "POST", "/pullRequest/merge", payload, nil)
		resp1.Body.Close()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)

		// отправляем второй запрос на мерж того же PR
		resp2 := makeRequest(t, "POST", "/pullRequest/merge", payload, nil)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusOK, resp2.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&result)

		pr := result["pr"].(map[string]interface{})
		assert.Equal(t, "MERGED", pr["status"])
	})

	t.Run("MergePR_NotFound", func(t *testing.T) {
		cleanupTestData(t)

		payload := map[string]interface{}{
			"pull_request_id": "nonexistent",
		}

		// отправляем запрос на мерж несуществующего PR
		resp := makeRequest(t, "POST", "/pullRequest/merge", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "NOT_FOUND", errObj["code"])
	})
}

func TestReassignReviewer(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("ReassignReviewer_Success", func(t *testing.T) {
		cleanupTestData(t)

		// Создаем команду с достаточным количеством пользователей
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "David", "is_active": true},
		})

		createTestPR(t, "pr-1", "Test PR", "u1")
		assignReviewer(t, "pr-1", "u2")

		payload := map[string]interface{}{
			"pull_request_id": "pr-1",
			"old_user_id":     "u2",
		}

		// отправляем запрос на переназначение ревьювера
		resp := makeRequest(t, "POST", "/pullRequest/reassign", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		pr := result["pr"].(map[string]interface{})
		replacedBy := result["replaced_by"].(string)

		reviewers := pr["assigned_reviewers"].([]interface{})

		for _, r := range reviewers {
			assert.NotEqual(t, "u2", r)
		}

		assert.Contains(t, []string{"u3", "u4"}, replacedBy)
	})

	t.Run("ReassignReviewer_NotAssigned", func(t *testing.T) {
		cleanupTestData(t)

		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		})

		createTestPR(t, "pr-1", "Test PR", "u1")
		payload := map[string]interface{}{
			"pull_request_id": "pr-1",
			"old_user_id":     "u2",
		}

		// отправляем запрос на переназначение ревьювера, который не назначен
		resp := makeRequest(t, "POST", "/pullRequest/reassign", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "NOT_ASSIGNED", errObj["code"])
	})

	t.Run("ReassignReviewer_OnMergedPR", func(t *testing.T) {
		cleanupTestData(t)

		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		})

		createTestPR(t, "pr-1", "Test PR", "u1")
		assignReviewer(t, "pr-1", "u2")

		ctx := context.Background()
		_, err := testDB.Exec(ctx, "UPDATE pull_requests SET status = 'MERGED', merged_at = CURRENT_TIMESTAMP WHERE pull_request_id = 'pr-1'")
		require.NoError(t, err)

		payload := map[string]interface{}{
			"pull_request_id": "pr-1",
			"old_user_id":     "u2",
		}

		// отправляем запрос на переназначение ревьювера на замерженном PR
		resp := makeRequest(t, "POST", "/pullRequest/reassign", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "PR_MERGED", errObj["code"])
	})

	t.Run("ReassignReviewer_NoCandidate", func(t *testing.T) {
		cleanupTestData(t)

		// Создаем команду с только 2 пользователями
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		})

		createTestPR(t, "pr-1", "Test PR", "u1")
		assignReviewer(t, "pr-1", "u2")

		payload := map[string]interface{}{
			"pull_request_id": "pr-1",
			"old_user_id":     "u2",
		}

		// отправляем запрос на переназначение ревьювера, когда нет кандидатов
		resp := makeRequest(t, "POST", "/pullRequest/reassign", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "NO_CANDIDATE", errObj["code"])
	})

	t.Run("ReassignReviewer_PRNotFound", func(t *testing.T) {
		cleanupTestData(t)

		payload := map[string]interface{}{
			"pull_request_id": "nonexistent",
			"old_user_id":     "u1",
		}

		// отправляем запрос на переназначение ревьювера для несуществующего PR
		resp := makeRequest(t, "POST", "/pullRequest/reassign", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "NOT_FOUND", errObj["code"])
	})
}
