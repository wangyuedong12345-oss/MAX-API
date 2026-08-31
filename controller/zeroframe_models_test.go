package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MAX-API-Next/MAX-API/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZeroFrameOfficialModelsIncludeAICanvasProfiles(t *testing.T) {
	models := zeroFrameOfficialModels()
	require.Len(t, models, 3)

	byID := map[string]zeroFrameModel{}
	for _, model := range models {
		byID[model.Id] = model
		assert.Equal(t, model.Id, model.Name)
		require.NotEmpty(t, model.Type)
		require.NotEmpty(t, model.AICanvas)
	}

	assert.Equal(t, "text", byID["DeepSeek-V4-flash"].Type)
	assert.Equal(t, "image", byID["Doubao-Seedream-5.0-lite"].Type)
	assert.Equal(t, "video", byID["Doubao-Seedance-2.0-mini"].Type)

	seedance := byID["Doubao-Seedance-2.0-mini"].AICanvas
	profile, ok := seedance["executionProfile"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "custom", profile["preset"])
	protocol, ok := profile["protocol"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(2), numericJSONValue(t, protocol["version"]))
	assert.Equal(t, "async", protocol["mode"])

	capability, ok := seedance["videoCapability"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, capability["durations"])
	assert.Equal(t, 3, capability["maxVideoReferences"])
}

func TestGetZeroFrameModelsResponseShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/zeroframe/models", GetZeroFrameModels)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/zeroframe/models", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var payload struct {
		Data []zeroFrameModel `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Len(t, payload.Data, 3)
}

func numericJSONValue(t *testing.T, value any) float64 {
	t.Helper()
	switch typed := value.(type) {
	case int:
		return float64(typed)
	case float64:
		return typed
	default:
		t.Fatalf("unexpected numeric value type %T", value)
		return 0
	}
}
