package common

import (
	"testing"

	"github.com/MAX-API-Next/MAX-API/constant"
	"github.com/stretchr/testify/require"
)

func TestCodexEndpointTypesIncludeAlphaSearch(t *testing.T) {
	endpointTypes := GetEndpointTypesByChannelType(constant.ChannelTypeCodex, "gpt-5.1")

	require.Contains(t, endpointTypes, constant.EndpointTypeOpenAIResponse)
	require.Contains(t, endpointTypes, constant.EndpointTypeOpenAIResponseCompact)
	require.Contains(t, endpointTypes, constant.EndpointTypeOpenAIAlphaSearch)
}

func TestSeedreamEndpointTypesPreferImageGeneration(t *testing.T) {
	endpointTypes := GetEndpointTypesByChannelType(constant.ChannelTypeVolcEngine, "doubao-seedream-5-0-260128")

	require.NotEmpty(t, endpointTypes)
	require.Equal(t, constant.EndpointTypeImageGeneration, endpointTypes[0])
}

func TestSeedreamDisplayNameEndpointTypesPreferImageGeneration(t *testing.T) {
	endpointTypes := GetEndpointTypesByChannelType(constant.ChannelTypeVolcEngine, "Doubao-Seedream-5.0-lite")

	require.NotEmpty(t, endpointTypes)
	require.Equal(t, constant.EndpointTypeImageGeneration, endpointTypes[0])
}
