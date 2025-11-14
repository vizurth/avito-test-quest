package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type StatsResult struct {
	ReviewerStats []interface{} `json:"reviewer_stats"`
	PRStats       []interface{} `json:"pr_stats"`
}

func TestStatsEndpoint(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("GetStats_EmptyDatabase", func(t *testing.T) {
		cleanupTestData(t)

		// отправляем запрос на получение статистики
		resp := makeRequest(t, "GET", "/stats", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result StatsResult
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Empty(t, result.ReviewerStats)
		assert.Empty(t, result.PRStats)
	})

	t.Run("GetStats_WithData", func(t *testing.T) {
		cleanupTestData(t)

		// создаем команды для тестов
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "David", "is_active": true},
		})

		// создаем PR
		createTestPR(t, "pr-1", "Test PR 1", "u1")
		createTestPR(t, "pr-2", "Test PR 2", "u1")
		createTestPR(t, "pr-3", "Test PR 3", "u2")

		// назначаем ревьюверов
		assignReviewer(t, "pr-1", "u2")
		assignReviewer(t, "pr-1", "u3")

		assignReviewer(t, "pr-2", "u2")
		assignReviewer(t, "pr-2", "u4")

		assignReviewer(t, "pr-3", "u1")
		assignReviewer(t, "pr-3", "u3")

		// отправляем запрос на получение статистики
		resp := makeRequest(t, "GET", "/stats", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Проверяем reviewer_stats
		reviewerStats := result["reviewer_stats"].([]interface{})
		assert.Len(t, reviewerStats, 4)

		// создаем мапу для проверки
		reviewerMap := make(map[string]int)
		for _, stat := range reviewerStats {
			s := stat.(map[string]interface{})
			userID := s["user_id"].(string)
			count := int(s["assigned_count"].(float64))
			reviewerMap[userID] = count
		}

		// проверяем количество назначений для каждого ревьювера
		assert.Equal(t, 2, reviewerMap["u2"], "u2 should have 2 assignments")
		assert.Equal(t, 2, reviewerMap["u3"], "u3 should have 2 assignments")
		assert.Equal(t, 1, reviewerMap["u4"], "u4 should have 1 assignment")
		assert.Equal(t, 1, reviewerMap["u1"], "u1 should have 1 assignment")

		prStats := result["pr_stats"].([]interface{})
		assert.Len(t, prStats, 3)

		prMap := make(map[string]int)
		for _, stat := range prStats {
			s := stat.(map[string]interface{})
			prID := s["pull_request_id"].(string)
			count := int(s["reviewer_count"].(float64))
			prMap[prID] = count
		}

		// проверяем количество ревьюверов на каждом PR
		assert.Equal(t, 2, prMap["pr-1"], "pr-1 should have 2 reviewers")
		assert.Equal(t, 2, prMap["pr-2"], "pr-2 should have 2 reviewers")
		assert.Equal(t, 2, prMap["pr-3"], "pr-3 should have 2 reviewers")
	})

	t.Run("GetStats_SortedByCount", func(t *testing.T) {
		cleanupTestData(t)

		// создаем команду с пользователями
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		})

		// создаем PR
		createTestPR(t, "pr-1", "Test PR 1", "u1")
		createTestPR(t, "pr-2", "Test PR 2", "u1")
		createTestPR(t, "pr-3", "Test PR 3", "u1")

		// u2 назначен на все PR (3 раза)
		assignReviewer(t, "pr-1", "u2")
		assignReviewer(t, "pr-2", "u2")
		assignReviewer(t, "pr-3", "u2")

		// u3 назначен только на pr-1 (1 раз)
		assignReviewer(t, "pr-1", "u3")

		// отправляем запрос на получение статистики
		resp := makeRequest(t, "GET", "/stats", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		// Проверяем, что reviewer_stats отсортирована по убыванию assigned_count
		reviewerStats := result["reviewer_stats"].([]interface{})

		// Первый должен быть u2 с 3 назначениями
		firstStat := reviewerStats[0].(map[string]interface{})
		assert.Equal(t, "u2", firstStat["user_id"])
		assert.Equal(t, float64(3), firstStat["assigned_count"])

		// Проверяем pr_stats отсортирована по убыванию reviewer_count
		prStats := result["pr_stats"].([]interface{})

		// pr-1 должен быть первым (2 ревьювера)
		firstPR := prStats[0].(map[string]interface{})
		assert.Equal(t, "pr-1", firstPR["pull_request_id"])
		assert.Equal(t, float64(2), firstPR["reviewer_count"])
	})

	t.Run("GetStats_WithUsersWithoutAssignments", func(t *testing.T) {
		cleanupTestData(t)

		// создаем команду с пользователями, которые не будут назначены
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		})

		createTestPR(t, "pr-1", "Test PR", "u1")
		assignReviewer(t, "pr-1", "u2")

		// отправляем запрос на получение статистики
		resp := makeRequest(t, "GET", "/stats", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		reviewerStats := result["reviewer_stats"].([]interface{})

		assert.Len(t, reviewerStats, 3)

		reviewerMap := make(map[string]int)
		for _, stat := range reviewerStats {
			s := stat.(map[string]interface{})
			userID := s["user_id"].(string)
			count := int(s["assigned_count"].(float64))
			reviewerMap[userID] = count
		}

		assert.Equal(t, 1, reviewerMap["u2"])
		assert.Equal(t, 0, reviewerMap["u1"])
		assert.Equal(t, 0, reviewerMap["u3"])
	})

	t.Run("GetStats_MultiplePRsWithDifferentReviewerCounts", func(t *testing.T) {
		cleanupTestData(t)

		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		})

		// pr с 2 ревьюверами
		createTestPR(t, "pr-1", "Test PR 1", "u1")
		assignReviewer(t, "pr-1", "u2")
		assignReviewer(t, "pr-1", "u3")

		// pr с 1 ревьювером
		createTestPR(t, "pr-2", "Test PR 2", "u1")
		assignReviewer(t, "pr-2", "u2")

		// pr без ревьюверов
		createTestPR(t, "pr-3", "Test PR 3", "u1")

		// отправляем запрос на получение статистики
		resp := makeRequest(t, "GET", "/stats", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		prStats := result["pr_stats"].([]interface{})
		assert.Len(t, prStats, 3)

		prMap := make(map[string]int)
		for _, stat := range prStats {
			s := stat.(map[string]interface{})
			prID := s["pull_request_id"].(string)
			count := int(s["reviewer_count"].(float64))
			prMap[prID] = count
		}

		assert.Equal(t, 2, prMap["pr-1"])
		assert.Equal(t, 1, prMap["pr-2"])
		assert.Equal(t, 0, prMap["pr-3"])
	})
}
