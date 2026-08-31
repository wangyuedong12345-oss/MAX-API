package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type zeroFrameModelResponse struct {
	Data []zeroFrameModel `json:"data"`
}

type zeroFrameModel struct {
	Id           string         `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	Capabilities map[string]any `json:"capabilities,omitempty"`
	AICanvas     map[string]any `json:"aiCanvas,omitempty"`
}

func zeroFrameOfficialModels() []zeroFrameModel {
	seedanceMiniDurations := integerRange(4, 15)
	return []zeroFrameModel{
		{
			Id:   "DeepSeek-V4-flash",
			Name: "DeepSeek-V4-flash",
			Type: "text",
			Capabilities: map[string]any{
				"input_modalities": []string{"text"},
			},
			AICanvas: map[string]any{
				"executionProfile": map[string]any{
					"preset": "openai-chat",
				},
				"inputModalities": []string{"text"},
			},
		},
		{
			Id:   "Doubao-Seedream-5.0-lite",
			Name: "Doubao-Seedream-5.0-lite",
			Type: "image",
			Capabilities: map[string]any{
				"input_modalities": []string{"text", "image"},
			},
			AICanvas: map[string]any{
				"executionProfile": map[string]any{
					"preset": "openai-image",
				},
				"imageReferenceRequestMode": "generation-json-image-data-urls",
				"inputModalities":           []string{"text", "image"},
			},
		},
		{
			Id:   "Doubao-Seedance-2.0-mini",
			Name: "Doubao-Seedance-2.0-mini",
			Type: "video",
			Capabilities: map[string]any{
				"input_modalities": []string{"text", "image"},
				"resolutions":      []string{"480p", "720p"},
				"ratios":           []string{"16:9", "9:16", "1:1", "4:3", "3:4", "21:9", "adaptive"},
				"durations":        seedanceMiniDurations,
			},
			AICanvas: map[string]any{
				"executionProfile": map[string]any{
					"preset":   "custom",
					"protocol": seedanceMiniProtocol(),
				},
				"videoCapability": map[string]any{
					"resolutions":        []string{"480p", "720p"},
					"defaultResolution":  "480p",
					"ratios":             []string{"16:9", "9:16", "1:1", "4:3", "3:4", "21:9", "adaptive"},
					"defaultRatio":       "16:9",
					"durations":          seedanceMiniDurations,
					"defaultDuration":    5,
					"frameRates":         []int{24},
					"defaultFrameRate":   24,
					"supportsAudio":      true,
					"maxImageReferences": 2,
					"maxVideoReferences": 3,
					"maxAudioReferences": 0,
				},
				"inputModalities": []string{"text", "image"},
			},
		},
	}
}

func integerRange(minValue int, maxValue int) []int {
	if maxValue < minValue {
		return nil
	}
	values := make([]int, 0, maxValue-minValue+1)
	for value := minValue; value <= maxValue; value++ {
		values = append(values, value)
	}
	return values
}

func seedanceMiniProtocol() map[string]any {
	return map[string]any{
		"version": 2,
		"mode":    "async",
		"auth": map[string]any{
			"type": "bearer",
		},
		"submit": map[string]any{
			"method":       "POST",
			"path":         "/videos/generations",
			"bodyEncoding": "json",
			"body": map[string]any{
				"model":          "{{model}}",
				"prompt":         "{{prompt}}",
				"content":        "{{seedanceContent}}",
				"duration":       "{{duration}}",
				"ratio":          "{{aspectRatio}}",
				"resolution":     "{{seedanceResolution}}",
				"generate_audio": "{{generateAudio}}",
			},
		},
		"response": map[string]any{
			"type":       "json",
			"taskIdPath": "video_id",
			"errorPath":  "error.message",
		},
		"poll": map[string]any{
			"method":   "GET",
			"path":     "/agnesapi",
			"pathMode": "origin",
			"query": map[string]any{
				"video_id": "{{submit.video_id}}",
			},
			"response": map[string]any{
				"statusPath":    "status",
				"successValues": []string{"completed", "succeeded", "success"},
				"failureValues": []string{"failed", "error", "canceled"},
				"result": map[string]any{
					"urlPath":  "url",
					"mimeType": "video/mp4",
				},
				"errorPath":    "error.message",
				"progressPath": "progress",
			},
			"intervalMs":    10000,
			"maxDurationMs": 3600000,
			"retry": map[string]any{
				"httpStatuses":       []int{408, 429, 500, 502, 503, 504},
				"maxRetries":         5,
				"backoff":            "exponential",
				"maxDelayMs":         60000,
				"honorRetryAfter":    true,
				"retryNetworkErrors": true,
			},
		},
	}
}

func GetZeroFrameModels(c *gin.Context) {
	c.JSON(http.StatusOK, zeroFrameModelResponse{Data: zeroFrameOfficialModels()})
}
