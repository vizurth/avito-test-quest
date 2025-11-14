package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestE2EWorkflow тестирует полный жизненный цикл работы с сервисом
func TestE2EWorkflow(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cleanupTestData(t)

	// 1. создаем команды с пользователями
	t.Log("Step 1: Creating teams")

	backendPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	resp := makeRequest(t, "POST", "/team/add", backendPayload, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	frontendPayload := map[string]interface{}{
		"team_name": "frontend",
		"members": []map[string]interface{}{
			{"user_id": "u4", "username": "David", "is_active": true},
			{"user_id": "u5", "username": "Eve", "is_active": true},
		},
	}
	resp = makeRequest(t, "POST", "/team/add", frontendPayload, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// 2. проверяем, что команды созданы
	t.Log("Step 2: Verifying teams")

	resp = makeRequest(t, "GET", "/team/get?team_name=backend", nil, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var backendTeam map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&backendTeam)
	resp.Body.Close()
	assert.Equal(t, "backend", backendTeam["team_name"])
	assert.Len(t, backendTeam["members"].([]interface{}), 3)

	// 3. создаем несколько PR от разных авторов
	t.Log("Step 3: Creating pull requests")

	pr1Payload := map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add search functionality",
		"author_id":         "u1",
	}
	resp = makeRequest(t, "POST", "/pullRequest/create", pr1Payload, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var pr1Result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&pr1Result)
	resp.Body.Close()

	pr1 := pr1Result["pr"].(map[string]interface{})
	assert.Equal(t, "OPEN", pr1["status"])
	reviewers1 := pr1["assigned_reviewers"].([]interface{})
	assert.Len(t, reviewers1, 2, "Should assign 2 reviewers from backend team")

	pr2Payload := map[string]interface{}{
		"pull_request_id":   "pr-1002",
		"pull_request_name": "Fix login bug",
		"author_id":         "u3",
	}
	resp = makeRequest(t, "POST", "/pullRequest/create", pr2Payload, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	pr3Payload := map[string]interface{}{
		"pull_request_id":   "pr-1003",
		"pull_request_name": "Update frontend components",
		"author_id":         "u4",
	}
	resp = makeRequest(t, "POST", "/pullRequest/create", pr3Payload, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// 4. проверяем список PR для ревьювера
	t.Log("Step 4: Checking reviewer's PRs")

	firstReviewer := reviewers1[0].(string)
	resp = makeRequest(t, "GET", "/users/getReview?user_id="+firstReviewer, nil, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var reviewerPRs map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&reviewerPRs)
	resp.Body.Close()

	prs := reviewerPRs["pull_requests"].([]interface{})
	assert.GreaterOrEqual(t, len(prs), 1, "Reviewer should have at least 1 PR assigned")

	// 5. деактивируем пользователя
	t.Log("Step 5: Deactivating user")

	deactivatePayload := map[string]interface{}{
		"user_id":   "u2",
		"is_active": false,
	}
	resp = makeRequest(t, "POST", "/users/setIsActive", deactivatePayload, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// 6. создаем новый PR - u2 не должен быть назначен
	t.Log("Step 6: Creating PR with inactive user")

	pr4Payload := map[string]interface{}{
		"pull_request_id":   "pr-1004",
		"pull_request_name": "Add new feature",
		"author_id":         "u1",
	}
	resp = makeRequest(t, "POST", "/pullRequest/create", pr4Payload, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var pr4Result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&pr4Result)
	resp.Body.Close()

	pr4 := pr4Result["pr"].(map[string]interface{})
	reviewers4 := pr4["assigned_reviewers"].([]interface{})

	for _, r := range reviewers4 {
		assert.NotEqual(t, "u2", r, "Inactive user u2 should not be assigned")
	}

	// 7. переназначаем ревьювера на pr-1001
	t.Log("Step 7: Reassigning reviewer")

	if len(reviewers1) > 0 {
		reassignPayload := map[string]interface{}{
			"pull_request_id": "pr-1001",
			"old_user_id":     reviewers1[0].(string),
		}
		resp = makeRequest(t, "POST", "/pullRequest/reassign", reassignPayload, nil)

		if resp.StatusCode == http.StatusOK {
			var reassignResult map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&reassignResult)

			newPR := reassignResult["pr"].(map[string]interface{})
			newReviewers := newPR["assigned_reviewers"].([]interface{})

			oldReviewer := reviewers1[0].(string)
			for _, r := range newReviewers {
				if r == oldReviewer {
					t.Errorf("Old reviewer %s should not be in the list", oldReviewer)
				}
			}
		}
		resp.Body.Close()
	}

	// 8. мержим PR
	t.Log("Step 8: Merging PR")

	mergePayload := map[string]interface{}{
		"pull_request_id": "pr-1001",
	}
	resp = makeRequest(t, "POST", "/pullRequest/merge", mergePayload, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var mergeResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&mergeResult)
	resp.Body.Close()

	mergedPR := mergeResult["pr"].(map[string]interface{})
	assert.Equal(t, "MERGED", mergedPR["status"])
	assert.NotNil(t, mergedPR["mergedAt"])

	// 9. пытаемся переназначить ревьювера на уже замерженный PR
	t.Log("Step 9: Trying to reassign on merged PR")

	reassignMergedPayload := map[string]interface{}{
		"pull_request_id": "pr-1001",
		"old_user_id":     reviewers1[0].(string),
	}
	resp = makeRequest(t, "POST", "/pullRequest/reassign", reassignMergedPayload, nil)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var errorResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&errorResult)
	resp.Body.Close()

	errObj := errorResult["error"].(map[string]interface{})
	assert.Equal(t, "PR_MERGED", errObj["code"])

	// 10. проверяем статистику
	t.Log("Step 10: Checking stats")

	resp = makeRequest(t, "GET", "/stats", nil, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&stats)
	resp.Body.Close()

	reviewerStats := stats["reviewer_stats"].([]interface{})
	prStats := stats["pr_stats"].([]interface{})

	assert.GreaterOrEqual(t, len(reviewerStats), 3, "Should have stats for all users")
	assert.Equal(t, 4, len(prStats), "Should have stats for all PRs")

	// 11. повторно мержим уже замерженный PR для проверки идемпотентности
	t.Log("Step 11: Testing merge idempotency")

	resp = makeRequest(t, "POST", "/pullRequest/merge", mergePayload, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	t.Log("E2E workflow completed successfully")
}
