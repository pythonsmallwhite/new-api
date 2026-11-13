package service

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaleChannelBillingQuotaUsesSelectedMultiplier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	common.SetContextKey(c, constant.ContextKeyChannelCostMultiplier, 0.8)

	assert.Equal(t, 80, ScaleChannelBillingQuota(c, &relaycommon.RelayInfo{}, 100))

	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	require.Equal(t, float64(1), ChannelBillingMultiplier(c2))
	assert.Equal(t, 100, ScaleChannelBillingQuota(c2, &relaycommon.RelayInfo{}, 100))
}
