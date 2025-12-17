package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/crush/internal/app"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/db"
	"github.com/stretchr/testify/require"
)

func setupTestApp(t *testing.T) (*app.App, func()) {
	ctx := context.Background()

	cfg, err := config.Init(t.TempDir(), t.TempDir(), false)
	require.NoError(t, err)

	conn, err := db.Connect(ctx, cfg.Options.DataDirectory)
	require.NoError(t, err)

	appInstance, err := app.New(ctx, conn, cfg)
	require.NoError(t, err)

	cleanup := func() {
		appInstance.Shutdown()
	}

	return appInstance, cleanup
}

func TestCreateSession(t *testing.T) {
	appInstance, cleanup := setupTestApp(t)
	defer cleanup()

	h := &handlers{app: appInstance}

	body := createSessionRequest{Title: "Test Session"}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/sessions", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.createSession(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.NotEmpty(t, resp["id"])
	require.Equal(t, "Test Session", resp["title"])
}

func TestListSessions(t *testing.T) {
	appInstance, cleanup := setupTestApp(t)
	defer cleanup()

	h := &handlers{app: appInstance}

	// Create a session first
	sess, err := appInstance.Sessions.Create(context.Background(), "Test Session")
	require.NoError(t, err)
	require.NotEmpty(t, sess.ID)

	req := httptest.NewRequest("GET", "/sessions", nil)
	w := httptest.NewRecorder()

	h.listSessions(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var sessions []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &sessions)
	require.NoError(t, err)

	require.Len(t, sessions, 1)
	require.Equal(t, "Test Session", sessions[0]["title"])
}

func TestSessionNotFound(t *testing.T) {
	appInstance, cleanup := setupTestApp(t)
	defer cleanup()

	h := &handlers{app: appInstance}

	req := httptest.NewRequest("GET", "/sessions/invalid-id", nil)
	req.SetPathValue("id", "invalid-id")
	w := httptest.NewRecorder()

	h.getSession(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}
