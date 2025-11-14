package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("HealthCheck_Success", func(t *testing.T) {
		// отправляем запрос на проверку здоровья сервиса
		resp := makeRequest(t, "GET", "/health", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "ok", result["status"])
	})

	t.Run("HealthCheck_MultipleRequests", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			// отправляем несколько запросов подряд на проверку здоровья сервиса
			resp := makeRequest(t, "GET", "/health", nil, nil)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}
