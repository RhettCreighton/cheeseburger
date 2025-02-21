package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type apiResponse struct {
	Page  int `json:"page"`
	Posts []struct {
		ID      int    `json:"ID"`
		Title   string `json:"Title"`
		Content string `json:"Content"`
	} `json:"posts"`
}

func TestAPIRoutes(t *testing.T) {
	// Set up test environment.
	db := setupTestDB(t)
	_ = setupTestTemplates(t)
	router, _, _ := setupTestRouter(t, db)

	t.Run("GET /api/posts returns list with pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/posts", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var res apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)

		require.Equal(t, 1, res.Page)
		require.Len(t, res.Posts, 1)
		require.Equal(t, 1, res.Posts[0].ID)
		require.Equal(t, "Test Post", res.Posts[0].Title)
		require.Equal(t, "This is a test post", res.Posts[0].Content)
	})
}
