package service

import (
	"errors"
	"math"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
)

func ChannelBillingMultiplier(c *gin.Context) float64 {
	return channelBillingMultiplier(c, nil)
}

func channelBillingMultiplier(c *gin.Context, info *relaycommon.RelayInfo) float64 {
	if c != nil {
		multiplier, ok := common.GetContextKeyType[float64](c, constant.ContextKeyChannelCostMultiplier)
		if ok && multiplier > 0 && !math.IsNaN(multiplier) && !math.IsInf(multiplier, 0) {
			return multiplier
		}
	}
	if info != nil && info.ChannelMeta != nil {
		setting := info.ChannelMeta.ChannelSetting
		return setting.CostMultiplierForModel(info.OriginModelName)
	}
	return 1
}

func ChannelBillingMultiplierForChannel(channelID int, modelName string) float64 {
	if channelID <= 0 || modelName == "" {
		return 1
	}
	channel, err := model.GetChannelById(channelID, false)
	if err != nil || channel == nil {
		return 1
	}
	setting := channel.GetSetting()
	return setting.CostMultiplierForModel(model.CostMultiplierModel(channel.GetModels(), modelName))
}

func ScaleChannelBillingQuota(c *gin.Context, info *relaycommon.RelayInfo, quota int) int {
	if quota <= 0 {
		return quota
	}
	scaled, clamp := common.QuotaFromFloatChecked(float64(quota) * channelBillingMultiplier(c, info))
	noteQuotaClamp(info, clamp)
	return scaled
}

func PrepareChannelBilling(c *gin.Context, info *relaycommon.RelayInfo, baseQuota int) *types.NewAPIError {
	targetQuota := ScaleChannelBillingQuota(c, info, baseQuota)
	if info != nil && info.Billing != nil {
		if err := info.Billing.ReserveBeforeSubmit(targetQuota); err != nil {
			var apiErr *types.NewAPIError
			if errors.As(err, &apiErr) {
				return apiErr
			}
			return types.NewErrorWithStatusCode(err, types.ErrorCodeInsufficientUserQuota, http.StatusForbidden, types.ErrOptionWithSkipRetry())
		}
		return nil
	}
	return PreConsumeBilling(c, targetQuota, info)
}
