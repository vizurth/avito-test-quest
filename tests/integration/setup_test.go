package integration

import (
	"avito-test-quest/internal/app"
	"avito-test-quest/internal/config"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var (
	testBaseURL string
	testDB      *pgxpool.Pool
)

// setupTestEnvironment инициализирует тестовое окружение
func setupTestEnvironment(t *testing.T) (*app.App, func()) {
	ctx := context.Background()

	// загружаем конфигурацию
	cfg, err := config.New()
	require.NoError(t, err, "failed to load config")

	// переопределяем хост для подключения к БД
	if cfg.Postgres.Host == "postgres" || cfg.Postgres.Host == "" {
		cfg.Postgres.Host = "localhost"
	}

	// переопределяем порт для тестового сервера
	cfg.PR.Port = "8081"
	testBaseURL = "http://localhost:8081"

	application, err := app.New(ctx, cfg)
	require.NoError(t, err, "failed to create app")

	go func() {
		_ = application.Run(ctx)
	}()

	// ждем, пока сервер запустится
	waitForServer(t, testBaseURL+"/health", 10*time.Second)

	// получаем подключение к тестовой БД
	testDB = getDBConnection(t, cfg)

	// функция очистки после тестов
	cleanup := func() {
		cleanupTestData(t)
		application.Shutdown(ctx)
		testDB.Close()
	}

	return application, cleanup
}

// waitForServer ожидает запуска сервера
func waitForServer(t *testing.T, url string, timeout time.Duration) {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("server did not start within %v", timeout)
}

// getDBConnection создает подключение к тестовой БД
func getDBConnection(t *testing.T, cfg *config.Config) *pgxpool.Pool {
	connString := cfg.Postgres.GetConnString()
	pool, err := pgxpool.New(context.Background(), connString)
	require.NoError(t, err, "failed to connect to test database")
	return pool
}

// cleanupTestData очищает тестовые данные из БД
func cleanupTestData(t *testing.T) {
	ctx := context.Background()

	// удаляем данные из всех таблиц
	queries := []string{
		"TRUNCATE TABLE pr_reviewers CASCADE",
		"TRUNCATE TABLE pull_requests CASCADE",
		"TRUNCATE TABLE users CASCADE",
		"TRUNCATE TABLE teams CASCADE",
	}

	for _, query := range queries {
		_, err := testDB.Exec(ctx, query)
		if err != nil {
			t.Logf("Warning: failed to cleanup table: %v", err)
		}
	}
}

// createTestTeam вспомогательная функция для создания команды в тестах
func createTestTeam(t *testing.T, teamName string, members []map[string]interface{}) {
	ctx := context.Background()

	var teamID int64
	err := testDB.QueryRow(ctx, "INSERT INTO teams (team_name) VALUES ($1) RETURNING id", teamName).Scan(&teamID)
	require.NoError(t, err, "failed to insert test team")

	// добавляем участников команды в таблицу users
	for _, member := range members {
		userID := member["user_id"].(string)
		username := member["username"].(string)
		isActive := member["is_active"].(bool)

		_, err := testDB.Exec(ctx,
			"INSERT INTO users (user_id, username, team_id, is_active) VALUES ($1, $2, $3, $4)",
			userID, username, teamID, isActive)
		require.NoError(t, err, fmt.Sprintf("failed to insert test user %s", userID))
	}
}

// createTestPR вспомогательная функция для создания PR в тестах
func createTestPR(t *testing.T, prID, prName, authorID string) {
	ctx := context.Background()
	_, err := testDB.Exec(ctx,
		"INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id) VALUES ($1, $2, $3)",
		prID, prName, authorID)
	require.NoError(t, err, "failed to insert test PR")
}

// assignReviewer вспомогательная функция для назначения ревьювера в тестах
func assignReviewer(t *testing.T, prID, reviewerID string) {
	ctx := context.Background()
	_, err := testDB.Exec(ctx,
		"INSERT INTO pr_reviewers (pull_request_id, reviewer_user_id) VALUES ($1, $2)",
		prID, reviewerID)
	require.NoError(t, err, "failed to assign reviewer")
}
