package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserEndpoints(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("SetIsActive_Success", func(t *testing.T) {
		cleanupTestData(t)

		// создаем команду
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		})

		payload := map[string]interface{}{
			"user_id":   "u1",
			"is_active": false,
		}

		// отправляем запрос на изменение статуса активности
		resp := makeRequest(t, "POST", "/users/setIsActive", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		user := result["user"].(map[string]interface{})
		assert.Equal(t, "u1", user["user_id"])
		assert.Equal(t, false, user["is_active"])
		assert.Equal(t, "backend", user["team_name"])
	})

	t.Run("SetIsActive_UserNotFound", func(t *testing.T) {
		cleanupTestData(t)

		payload := map[string]interface{}{
			"user_id":   "nonexistent",
			"is_active": false,
		}

		// отправляем запрос на изменение статуса активности несуществующего пользователя
		resp := makeRequest(t, "POST", "/users/setIsActive", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "NOT_FOUND", errObj["code"])
	})

	t.Run("GetUserReviews_Success", func(t *testing.T) {
		cleanupTestData(t)

		// создаем команду с двумя пользователями
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		})

		// создаем два PR от u1
		createTestPR(t, "pr-1", "Test PR 1", "u1")
		createTestPR(t, "pr-2", "Test PR 2", "u1")

		// назначаем u2 ревьювером обоих PR
		assignReviewer(t, "pr-1", "u2")
		assignReviewer(t, "pr-2", "u2")

		// запрашиваем ревью для u2
		resp := makeRequest(t, "GET", "/users/getReview?user_id=u2", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "u2", result["user_id"])

		prs := result["pull_requests"].([]interface{})
		assert.Len(t, prs, 2)

		// Проверяем, что оба PR присутствуют
		prIDs := make(map[string]bool)
		for _, pr := range prs {
			prObj := pr.(map[string]interface{})
			prIDs[prObj["pull_request_id"].(string)] = true
		}
		assert.True(t, prIDs["pr-1"])
		assert.True(t, prIDs["pr-2"])
	})

	t.Run("GetUserReviews_NoReviews", func(t *testing.T) {
		cleanupTestData(t)

		// создаем команду
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		})

		resp := makeRequest(t, "GET", "/users/getReview?user_id=u1", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "u1", result["user_id"])

		prsRaw, ok := result["pull_requests"]
		require.True(t, ok, "ключ pull_requests должен существовать")

		prs, ok := prsRaw.([]interface{})
		require.False(t, ok, "pull_requests должен быть массивом")

		assert.Len(t, prs, 0)
	})

	t.Run("GetUserReviews_UserNotFound", func(t *testing.T) {
		cleanupTestData(t)

		// запрашиваем ревью для несуществующего пользователя
		resp := makeRequest(t, "GET", "/users/getReview?user_id=nonexistent", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "NOT_FOUND", errObj["code"])
	})

	t.Run("SetIsActive_ToggleMultipleTimes", func(t *testing.T) {
		cleanupTestData(t)

		// создаем команду
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		})

		payload1 := map[string]interface{}{
			"user_id":   "u1",
			"is_active": false,
		}
		resp1 := makeRequest(t, "POST", "/users/setIsActive", payload1, nil)
		resp1.Body.Close()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)

		payload2 := map[string]interface{}{
			"user_id":   "u1",
			"is_active": true,
		}
		resp2 := makeRequest(t, "POST", "/users/setIsActive", payload2, nil)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusOK, resp2.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&result)

		user := result["user"].(map[string]interface{})
		assert.Equal(t, true, user["is_active"])
	})
}
