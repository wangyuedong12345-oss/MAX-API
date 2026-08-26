package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MAX-API-Next/MAX-API/common"
	"github.com/MAX-API-Next/MAX-API/constant"
	"github.com/MAX-API-Next/MAX-API/model"
	relayconstant "github.com/MAX-API-Next/MAX-API/relay/constant"
	"github.com/MAX-API-Next/MAX-API/service"
	"github.com/MAX-API-Next/MAX-API/setting"
	"github.com/MAX-API-Next/MAX-API/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

func TestGetModelFromJSONBodyPreservesProviderRoutingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{
		"model":"gpt-test",
		"routing_strategy":"cost",
		"provider":{"sort":"price","order":["openai"]}
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	t.Cleanup(func() { common.CleanupBodyStorage(ctx) })

	request, err := getModelFromJSONBody(ctx)
	require.NoError(t, err)
	require.Equal(t, "gpt-test", request.Model)

	storage, err := common.GetBodyStorage(ctx)
	require.NoError(t, err)
	body, err := storage.Bytes()
	require.NoError(t, err)
	require.Equal(t, "cost", gjson.GetBytes(body, "routing_strategy").String())
	require.Equal(t, "price", gjson.GetBytes(body, "provider.sort").String())
	require.Equal(t, "openai", gjson.GetBytes(body, "provider.order.0").String())
}

func TestOpenAIVideoPostWithTaskIDRoutesAsFetch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name        string
		path        string
		body        string
		contentType string
		params      gin.Params
	}{
		{
			name:        "body video_id",
			path:        "/v1/videos/generations",
			body:        `{"video_id":"task_123"}`,
			contentType: "application/json",
		},
		{
			name:   "path task_id",
			path:   "/v1/videos/generations/task_123",
			params: gin.Params{{Key: "task_path", Value: "/task_123"}},
		},
		{
			name:   "single video path",
			path:   "/v1/videos/task_123",
			params: gin.Params{{Key: "video_id", Value: "task_123"}},
		},
		{
			name:   "legacy single video path",
			path:   "/v1/video/generations/task_123",
			params: gin.Params{{Key: "task_id", Value: "task_123"}},
		},
		{
			name:   "agnes query path",
			path:   "/agnesapi?video_id=task_123",
			params: gin.Params{},
		},
		{
			name:   "agnes v1 tasks path",
			path:   "/v1/tasks/task_123?language=zh",
			params: gin.Params{{Key: "task_id", Value: "task_123"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest(http.MethodPost, test.path, bytes.NewBufferString(test.body))
			if test.contentType != "" {
				ctx.Request.Header.Set("Content-Type", test.contentType)
			}
			ctx.Params = test.params
			t.Cleanup(func() { common.CleanupBodyStorage(ctx) })

			modelRequest, shouldSelectChannel, err := getModelRequest(ctx)
			require.NoError(t, err)

			require.False(t, shouldSelectChannel)
			require.Empty(t, modelRequest.Model)
			require.Equal(t, relayconstant.RelayModeVideoFetchByID, ctx.GetInt("relay_mode"))
			require.Equal(t, "task_123", ctx.GetString("task_id"))
		})
	}
}

func TestOpenAIVideoPostWithModelRoutesAsSubmit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/videos/generations", bytes.NewBufferString(`{"model":"Doubao-Seedance-2.0-mini","prompt":"test"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	t.Cleanup(func() { common.CleanupBodyStorage(ctx) })

	modelRequest, shouldSelectChannel, err := getModelRequest(ctx)
	require.NoError(t, err)

	require.True(t, shouldSelectChannel)
	require.Equal(t, "Doubao-Seedance-2.0-mini", modelRequest.Model)
	require.Equal(t, relayconstant.RelayModeVideoSubmit, ctx.GetInt("relay_mode"))
	require.Empty(t, ctx.GetString("task_id"))
}

func TestPlaygroundExplicitGroupBypassesStoredTokenRoutePlan(t *testing.T) {
	previousDB := model.DB
	previousMemoryCache := common.MemoryCacheEnabled
	usableGroups := setting.GetUserUsableGroupsCopy()
	usableGroupsJSON, err := common.Marshal(usableGroups)
	require.NoError(t, err)
	groupRatiosJSON := ratio_setting.GroupRatio2JSONString()
	t.Cleanup(func() {
		model.DB = previousDB
		common.MemoryCacheEnabled = previousMemoryCache
		require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(string(usableGroupsJSON)))
		require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(groupRatiosJSON))
	})

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Channel{}, &model.Ability{}))
	model.DB = db
	common.MemoryCacheEnabled = false
	require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(`{"default":"Default","base":"Base","deluxe":"Deluxe"}`))
	require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(`{"default":1,"base":1,"deluxe":2}`))

	modelName := "gpt-playground-route"
	for index, group := range []string{"base", "deluxe"} {
		channelID := index + 1
		priority := int64(0)
		require.NoError(t, db.Create(&model.Channel{
			Id:     channelID,
			Type:   1,
			Key:    fmt.Sprintf("key-%d", channelID),
			Status: common.ChannelStatusEnabled,
			Name:   group,
			Models: modelName,
			Group:  group,
		}).Error)
		require.NoError(t, db.Create(&model.Ability{
			Group:     group,
			Model:     modelName,
			ChannelId: channelID,
			Enabled:   true,
			Priority:  &priority,
			Weight:    100,
		}).Error)
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/pg/chat/completions", bytes.NewBufferString(fmt.Sprintf(`{"model":%q,"group":"deluxe"}`, modelName)))
	ctx.Request.Header.Set("Content-Type", "application/json")
	common.SetContextKey(ctx, constant.ContextKeyUserGroup, "default")
	common.SetContextKey(ctx, constant.ContextKeyUsingGroup, "base")
	common.SetContextKey(ctx, constant.ContextKeyTokenGroup, "base")
	common.SetContextKey(ctx, constant.ContextKeyTokenRoutingPolicy, model.TokenRoutingPolicy{
		Version:        model.TokenRoutingPolicyVersion,
		Mode:           model.TokenRoutingModeManual,
		Groups:         []string{"base"},
		RetryOnFailure: true,
	})
	t.Cleanup(func() { common.CleanupBodyStorage(ctx) })

	Distribute()(ctx)

	require.False(t, ctx.IsAborted())
	require.Equal(t, "deluxe", common.GetContextKeyString(ctx, constant.ContextKeyUsingGroup))
	require.Equal(t, 2, common.GetContextKeyInt(ctx, constant.ContextKeyChannelId))
	_, hasPlan := common.GetContextKey(ctx, constant.ContextKeyTokenRoutePlan)
	require.False(t, hasPlan)
}

func TestFindAffinityRouteGroupUsesFirstMatchingRouteGroup(t *testing.T) {
	group, index, ok := findAffinityRouteGroup(&service.TokenRoutePlan{
		OrderedGroups: []string{"base", "deluxe"},
	}, "gpt-affinity-route", 7, func(group string, modelName string, channelID int) bool {
		return group == "deluxe" && modelName == "gpt-affinity-route" && channelID == 7
	})
	require.True(t, ok)
	require.Equal(t, "deluxe", group)
	require.Equal(t, 1, index)
}
