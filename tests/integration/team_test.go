package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamEndpoints(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("CreateTeam_Success", func(t *testing.T) {
		cleanupTestData(t)

		payload := map[string]interface{}{
			"team_name": "backend",
			"members": []map[string]interface{}{
				{"user_id": "u1", "username": "Alice", "is_active": true},
				{"user_id": "u2", "username": "Bob", "is_active": true},
			},
		}

		// отправляем запрос на создание команды
		resp := makeRequest(t, "POST", "/team/add", payload, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		team := result["team"].(map[string]interface{})
		assert.Equal(t, "backend", team["team_name"])

		members := team["members"].([]interface{})
		assert.Len(t, members, 2)
	})

	t.Run("CreateTeam_DuplicateName", func(t *testing.T) {
		cleanupTestData(t)

		payload := map[string]interface{}{
			"team_name": "backend",
			"members": []map[string]interface{}{
				{"user_id": "u1", "username": "Alice", "is_active": true},
			},
		}

		// отправляем первый запрос на создание команды
		resp1 := makeRequest(t, "POST", "/team/add", payload, nil)
		resp1.Body.Close()
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		// отправляем второй запрос с тем же именем команды
		resp2 := makeRequest(t, "POST", "/team/add", payload, nil)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "TEAM_EXISTS", errObj["code"])
	})

	t.Run("GetTeam_Success", func(t *testing.T) {
		cleanupTestData(t)

		// создаем команду для теста
		createTestTeam(t, "backend", []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": false},
		})

		// отправляем запрос на получение команды
		resp := makeRequest(t, "GET", "/team/get?team_name=backend", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "backend", result["team_name"])
		members := result["members"].([]interface{})
		assert.Len(t, members, 2)
	})

	t.Run("GetTeam_NotFound", func(t *testing.T) {
		cleanupTestData(t)

		// отправляем запрос на получение несуществующей команды
		resp := makeRequest(t, "GET", "/team/get?team_name=nonexistent", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		errObj := result["error"].(map[string]interface{})
		assert.Equal(t, "NOT_FOUND", errObj["code"])
	})

	t.Run("CreateTeam_UpdateExistingUser", func(t *testing.T) {
		cleanupTestData(t)

		payload1 := map[string]interface{}{
			"team_name": "backend",
			"members": []map[string]interface{}{
				{"user_id": "u1", "username": "Alice", "is_active": true},
			},
		}

		// создаем первую команду
		resp1 := makeRequest(t, "POST", "/team/add", payload1, nil)
		resp1.Body.Close()
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		payload2 := map[string]interface{}{
			"team_name": "frontend",
			"members": []map[string]interface{}{
				{"user_id": "u1", "username": "Alice Updated", "is_active": false},
			},
		}

		// создаем вторую команду с тем же пользователем, но обновленными данными
		resp2 := makeRequest(t, "POST", "/team/add", payload2, nil)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusCreated, resp2.StatusCode)

		// проверяем, что данные пользователя были обновлены
		respGet := makeRequest(t, "GET", "/team/get?team_name=frontend", nil, nil)
		defer respGet.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(respGet.Body).Decode(&result)

		members := result["members"].([]interface{})
		member := members[0].(map[string]interface{})
		assert.Equal(t, "Alice Updated", member["username"])
		assert.Equal(t, false, member["is_active"])
	})
}

// makeRequest вспомогательная функция для выполнения HTTP запросов
func makeRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) *http.Response {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, testBaseURL+path, reqBody)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}
