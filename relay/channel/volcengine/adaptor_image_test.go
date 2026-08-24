package volcengine

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MAX-API-Next/MAX-API/common"
	"github.com/MAX-API-Next/MAX-API/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBuildVolcengineImageRequestPreservesReferenceImage(t *testing.T) {
	var request dto.ImageRequest
	require.NoError(t, common.Unmarshal([]byte(`{
		"model": "doubao-seedream-5-0-260128",
		"prompt": "将参考图中的小猪替换为小狗，保持风格与参考图一致",
		"reference_images": ["https://example.com/pig.png"],
		"size": "1920x1920"
	}`), &request))

	payload, err := buildVolcengineImageRequest(nil, request)

	require.NoError(t, err)
	require.JSONEq(t, `["https://example.com/pig.png"]`, string(payload["image"]))
	require.NotContains(t, payload, "reference_images")
}

func TestBuildVolcengineImageRequestNormalizesImageURLAlias(t *testing.T) {
	var request dto.ImageRequest
	require.NoError(t, common.Unmarshal([]byte(`{
		"model": "doubao-seedream-5-0-260128",
		"prompt": "将参考图中的小猪替换为小狗，保持风格与参考图一致",
		"image_url": "https://example.com/pig.png",
		"size": "1920x1920"
	}`), &request))

	payload, err := buildVolcengineImageRequest(nil, request)

	require.NoError(t, err)
	require.JSONEq(t, `["https://example.com/pig.png"]`, string(payload["image"]))
	require.NotContains(t, payload, "image_url")
}

func TestBuildVolcengineImageRequestKeepsImageArray(t *testing.T) {
	var request dto.ImageRequest
	require.NoError(t, common.Unmarshal([]byte(`{
		"model": "doubao-seedream-5-0-260128",
		"prompt": "use references",
		"image": ["https://example.com/a.png", "https://example.com/b.png"],
		"size": "1920x1920"
	}`), &request))

	payload, err := buildVolcengineImageRequest(nil, request)

	require.NoError(t, err)
	require.JSONEq(t, `["https://example.com/a.png", "https://example.com/b.png"]`, string(payload["image"]))
}

func TestBuildVolcengineImageRequestExpandsExtraFields(t *testing.T) {
	var request dto.ImageRequest
	require.NoError(t, common.Unmarshal([]byte(`{
		"model": "doubao-seedream-5-0-260128",
		"prompt": "a cute dog",
		"size": "1920x1920",
		"extra_fields": {
			"guidance_scale": 3.5,
			"watermark": false
		}
	}`), &request))

	payload, err := buildVolcengineImageRequest(nil, request)

	require.NoError(t, err)
	require.NotContains(t, payload, "extra_fields")
	require.JSONEq(t, `3.5`, string(payload["guidance_scale"]))
	require.JSONEq(t, `false`, string(payload["watermark"]))
}

func TestBuildVolcengineImageRequestConvertsMultipartImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("model", "doubao-seedream-5-0-260128"))
	require.NoError(t, writer.WriteField("prompt", "将参考图中的小猪替换为小狗，保持风格与参考图一致"))
	part, err := writer.CreateFormFile("reference_images[]", "pig.png")
	require.NoError(t, err)
	_, err = part.Write([]byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	})
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/images/edits", &body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())

	payload, err := buildVolcengineImageRequest(ctx, dto.ImageRequest{
		Model:  "doubao-seedream-5-0-260128",
		Prompt: "将参考图中的小猪替换为小狗，保持风格与参考图一致",
		Size:   "1920x1920",
	})

	require.NoError(t, err)
	require.Contains(t, payload, "image")
	require.True(t, strings.HasPrefix(string(payload["image"]), `["data:image/png;base64,`))
}
