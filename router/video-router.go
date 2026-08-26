package router

import (
	"strings"

	"github.com/MAX-API-Next/MAX-API/controller"
	"github.com/MAX-API-Next/MAX-API/middleware"
	relayconstant "github.com/MAX-API-Next/MAX-API/relay/constant"

	"github.com/gin-gonic/gin"
)

func relayTaskSubmitOrFetch(c *gin.Context) {
	if c.GetInt("relay_mode") == relayconstant.RelayModeVideoFetchByID {
		controller.RelayTaskFetch(c)
		return
	}
	controller.RelayTask(c)
}

func relayTaskFetchFromVideoGenerations(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		taskID = c.Param("video_id")
	}
	if taskID == "" {
		taskID = c.Query("task_id")
	}
	if taskID == "" {
		taskID = c.Query("id")
	}
	if taskID == "" {
		taskID = c.Query("video_id")
	}
	if taskID == "" {
		taskPath := strings.Trim(c.Param("task_path"), "/")
		if taskPath != "" {
			taskID = strings.Split(taskPath, "/")[0]
		}
	}
	if taskID != "" {
		c.Set("task_id", taskID)
	}
	controller.RelayTaskFetch(c)
}

func SetVideoRouter(router *gin.Engine) {
	// Video proxy: accepts either session auth (dashboard) or token auth (API clients)
	videoProxyRouter := router.Group("/v1")
	videoProxyRouter.Use(middleware.RouteTag("relay"))
	videoProxyRouter.Use(middleware.TokenOrUserAuth())
	{
		videoProxyRouter.GET("/videos/:task_id/content", controller.VideoProxy)
	}

	videoV1Router := router.Group("/v1")
	videoV1Router.Use(middleware.RouteTag("relay"))
	videoV1Router.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		videoV1Router.GET("/tasks/:task_id", relayTaskFetchFromVideoGenerations)
		videoV1Router.POST("/tasks/:task_id", relayTaskFetchFromVideoGenerations)
		videoV1Router.POST("/video/generations", relayTaskSubmitOrFetch)
		videoV1Router.POST("/video/generations/:task_id", controller.RelayTaskFetch)
		videoV1Router.GET("/video/generations/:task_id", controller.RelayTaskFetch)
		videoV1Router.POST("/videos/:video_id/remix", controller.RelayTask)
	}
	// openai compatible API video routes
	// docs: https://platform.openai.com/docs/api-reference/videos/create
	{
		videoV1Router.POST("/videos/generations", relayTaskSubmitOrFetch)
		videoV1Router.POST("/videos/generations/*task_path", relayTaskFetchFromVideoGenerations)
		videoV1Router.POST("/videos", relayTaskSubmitOrFetch)
		videoV1Router.POST("/videos/:video_id", relayTaskFetchFromVideoGenerations)
		videoV1Router.GET("/videos/generations", relayTaskFetchFromVideoGenerations)
		videoV1Router.GET("/videos/generations/*task_path", relayTaskFetchFromVideoGenerations)
		videoV1Router.GET("/videos/:task_id", controller.RelayTaskFetch)
	}

	agnesRouter := router.Group("/")
	agnesRouter.Use(middleware.RouteTag("relay"))
	agnesRouter.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		agnesRouter.GET("/agnesapi", relayTaskFetchFromVideoGenerations)
		agnesRouter.POST("/agnesapi", relayTaskFetchFromVideoGenerations)
	}

	klingV1Router := router.Group("/kling/v1")
	klingV1Router.Use(middleware.RouteTag("relay"))
	klingV1Router.Use(middleware.KlingRequestConvert(), middleware.TokenAuth(), middleware.Distribute())
	{
		klingV1Router.POST("/videos/text2video", controller.RelayTask)
		klingV1Router.POST("/videos/image2video", controller.RelayTask)
		klingV1Router.POST("/videos/omni-video", controller.RelayTask)
		klingV1Router.GET("/videos/text2video/:task_id", controller.RelayTaskFetch)
		klingV1Router.GET("/videos/image2video/:task_id", controller.RelayTaskFetch)
		klingV1Router.GET("/videos/omni-video/:task_id", controller.RelayTaskFetch)
	}

	// Jimeng official API routes - direct mapping to official API format
	jimengOfficialGroup := router.Group("jimeng")
	jimengOfficialGroup.Use(middleware.RouteTag("relay"))
	jimengOfficialGroup.Use(middleware.JimengRequestConvert(), middleware.TokenAuth(), middleware.Distribute())
	{
		// Maps to: /?Action=CVSync2AsyncSubmitTask&Version=2022-08-31 and /?Action=CVSync2AsyncGetResult&Version=2022-08-31
		jimengOfficialGroup.POST("/", controller.RelayTask)
	}
}
