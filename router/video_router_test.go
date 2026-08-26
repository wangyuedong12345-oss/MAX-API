package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestVideoRouterRegistersOpenAIVideoGenerations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetVideoRouter(engine)

	foundSubmit := false
	foundLegacySubmit := false
	foundLegacyPostFetchPath := false
	foundAgnesQuery := false
	foundAgnesV1TaskQuery := false
	foundPostFetchPath := false
	foundSinglePostFetchPath := false
	foundFetchQuery := false
	foundFetchPath := false
	for _, route := range engine.Routes() {
		if route.Method == http.MethodPost && route.Path == "/v1/video/generations" {
			foundLegacySubmit = true
		}
		if route.Method == http.MethodPost && route.Path == "/v1/video/generations/:task_id" {
			foundLegacyPostFetchPath = true
		}
		if route.Method == http.MethodGet && route.Path == "/agnesapi" {
			foundAgnesQuery = true
		}
		if route.Method == http.MethodGet && route.Path == "/v1/tasks/:task_id" {
			foundAgnesV1TaskQuery = true
		}
		if route.Method == http.MethodPost && route.Path == "/v1/videos/generations" {
			foundSubmit = true
		}
		if route.Method == http.MethodPost && route.Path == "/v1/videos/generations/*task_path" {
			foundPostFetchPath = true
		}
		if route.Method == http.MethodPost && route.Path == "/v1/videos/:video_id" {
			foundSinglePostFetchPath = true
		}
		if route.Method == http.MethodGet && route.Path == "/v1/videos/generations" {
			foundFetchQuery = true
		}
		if route.Method == http.MethodGet && route.Path == "/v1/videos/generations/*task_path" {
			foundFetchPath = true
		}
	}

	require.True(t, foundLegacySubmit, "expected POST /v1/video/generations route to be registered")
	require.True(t, foundLegacyPostFetchPath, "expected POST /v1/video/generations/:task_id route to be registered")
	require.True(t, foundAgnesQuery, "expected GET /agnesapi route to be registered")
	require.True(t, foundAgnesV1TaskQuery, "expected GET /v1/tasks/:task_id route to be registered")
	require.True(t, foundSubmit, "expected POST /v1/videos/generations route to be registered")
	require.True(t, foundPostFetchPath, "expected POST /v1/videos/generations/*task_path route to be registered")
	require.True(t, foundSinglePostFetchPath, "expected POST /v1/videos/:video_id route to be registered")
	require.True(t, foundFetchQuery, "expected GET /v1/videos/generations route to be registered")
	require.True(t, foundFetchPath, "expected GET /v1/videos/generations/*task_path route to be registered")
}

func TestVideoGenerationsFetchCompatibilityDoesNotFallThroughNoRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetVideoRouter(engine)

	tests := []string{
		"/agnesapi?video_id=task_123",
		"/v1/videos/generations?task_id=task_123",
		"/v1/videos/generations/task_123",
		"/v1/videos/generations/task_123/status",
		"/v1/videos/task_123",
		"/v1/tasks/task_123?language=zh",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, path, nil)
			engine.ServeHTTP(recorder, request)

			require.NotEqual(t, http.StatusNotFound, recorder.Code)
		})
	}
}

func TestVideoGenerationsPostFetchCompatibilityDoesNotFallThroughNoRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetVideoRouter(engine)

	tests := []string{
		"/v1/video/generations/task_123",
		"/v1/videos/generations/task_123",
		"/v1/videos/task_123",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, path, nil)
			engine.ServeHTTP(recorder, request)

			require.NotEqual(t, http.StatusNotFound, recorder.Code)
		})
	}
}
