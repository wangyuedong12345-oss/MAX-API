package relay

import (
	"testing"

	"github.com/MAX-API-Next/MAX-API/constant"
	"github.com/MAX-API-Next/MAX-API/dto"
	relaycommon "github.com/MAX-API-Next/MAX-API/relay/common"
	"github.com/MAX-API-Next/MAX-API/setting/model_setting"

	"github.com/stretchr/testify/require"
)

func TestShouldPassThroughImageRequestSkipsVolcengine(t *testing.T) {
	original := model_setting.GetGlobalSettings().PassThroughRequestEnabled
	model_setting.GetGlobalSettings().PassThroughRequestEnabled = true
	t.Cleanup(func() {
		model_setting.GetGlobalSettings().PassThroughRequestEnabled = original
	})

	require.False(t, shouldPassThroughImageRequest(&relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:    constant.ChannelTypeVolcEngine,
			ChannelSetting: dto.ChannelSettings{PassThroughBodyEnabled: true},
		},
	}))
}

func TestShouldPassThroughImageRequestKeepsGenericBehavior(t *testing.T) {
	require.True(t, shouldPassThroughImageRequest(&relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:    constant.ChannelTypeOpenAI,
			ChannelSetting: dto.ChannelSettings{PassThroughBodyEnabled: true},
		},
	}))
}
