package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type dailyResetPolicyUserSubRepoStub struct {
	userSubRepoNoop

	resetDailyCalled bool
}

func (r *dailyResetPolicyUserSubRepoStub) ResetDailyUsage(_ context.Context, _ int64, _ time.Time) error {
	r.resetDailyCalled = true
	return nil
}

func singleDayCardTestGroup(limit float64) *Group {
	return &Group{
		SubscriptionType:  SubscriptionTypeSubscription,
		SingleDayCardMode: true,
		DailyLimitUSD:     &limit,
	}
}

func TestValidateAndCheckLimits_SingleDayCardSkipsDailyReset(t *testing.T) {
	svc := &SubscriptionService{}
	limit := 10.0
	now := time.Now()
	startsAt := now.Add(-20 * time.Hour)
	expiresAt := startsAt.AddDate(0, 0, 1)
	dailyWindowStart := startOfDay(startsAt)

	sub := &UserSubscription{
		Status:           SubscriptionStatusActive,
		StartsAt:         startsAt,
		ExpiresAt:        expiresAt,
		DailyWindowStart: &dailyWindowStart,
		DailyUsageUSD:    9,
	}
	group := singleDayCardTestGroup(limit)

	needsMaintenance, err := svc.ValidateAndCheckLimits(sub, group)

	require.NoError(t, err)
	require.False(t, needsMaintenance, "单日卡在有效期内不应触发日重置维护")
	require.Equal(t, 9.0, sub.DailyUsageUSD, "单日卡不应在0点后被内存重置")
}

func TestValidateAndCheckLimits_RegularOneDaySubscriptionStillMarksDailyMaintenance(t *testing.T) {
	svc := &SubscriptionService{}
	limit := 10.0
	now := time.Now()
	startsAt := now.Add(-20 * time.Hour)
	expiresAt := startsAt.AddDate(0, 0, 1)
	dailyWindowStart := startOfDay(startsAt)

	sub := &UserSubscription{
		Status:           SubscriptionStatusActive,
		StartsAt:         startsAt,
		ExpiresAt:        expiresAt,
		DailyWindowStart: &dailyWindowStart,
		DailyUsageUSD:    9,
	}
	group := &Group{DailyLimitUSD: &limit}

	needsMaintenance, err := svc.ValidateAndCheckLimits(sub, group)

	require.NoError(t, err)
	require.True(t, needsMaintenance, "普通1天订阅在日窗口过期后仍应触发维护")
	require.Equal(t, 0.0, sub.DailyUsageUSD, "普通1天订阅应维持原有的内存重置行为")
}

func TestValidateAndCheckLimits_MultiDaySubscriptionStillMarksDailyMaintenance(t *testing.T) {
	svc := &SubscriptionService{}
	limit := 10.0
	now := time.Now()
	startsAt := now.Add(-60 * time.Hour)
	expiresAt := startsAt.AddDate(0, 0, 7)
	dailyWindowStart := startOfDay(now.Add(-48 * time.Hour))

	sub := &UserSubscription{
		Status:           SubscriptionStatusActive,
		StartsAt:         startsAt,
		ExpiresAt:        expiresAt,
		DailyWindowStart: &dailyWindowStart,
		DailyUsageUSD:    9,
	}
	group := &Group{DailyLimitUSD: &limit}

	needsMaintenance, err := svc.ValidateAndCheckLimits(sub, group)

	require.NoError(t, err)
	require.True(t, needsMaintenance, "多天订阅窗口过期后应继续触发维护")
	require.Equal(t, 0.0, sub.DailyUsageUSD, "多天订阅应维持原有的内存重置行为")
}

func TestCheckAndResetWindows_SingleDayCardSkipsDailyReset(t *testing.T) {
	repo := &dailyResetPolicyUserSubRepoStub{}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	now := time.Now()
	startsAt := now.Add(-20 * time.Hour)
	expiresAt := startsAt.AddDate(0, 0, 1)
	dailyWindowStart := startOfDay(startsAt)

	sub := &UserSubscription{
		ID:               1,
		UserID:           100,
		GroupID:          200,
		StartsAt:         startsAt,
		ExpiresAt:        expiresAt,
		DailyWindowStart: &dailyWindowStart,
		DailyUsageUSD:    4.2,
		Group:            singleDayCardTestGroup(10),
	}

	err := svc.CheckAndResetWindows(context.Background(), sub)

	require.NoError(t, err)
	require.False(t, repo.resetDailyCalled, "单日卡在有效期内不应触发数据库日重置")
	require.Equal(t, 4.2, sub.DailyUsageUSD)
	require.NotNil(t, sub.DailyWindowStart)
}

func TestCheckAndResetWindows_RegularOneDaySubscriptionStillResetsDaily(t *testing.T) {
	repo := &dailyResetPolicyUserSubRepoStub{}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	now := time.Now()
	startsAt := now.Add(-20 * time.Hour)
	expiresAt := startsAt.AddDate(0, 0, 1)
	dailyWindowStart := startOfDay(startsAt)

	sub := &UserSubscription{
		ID:               2,
		UserID:           101,
		GroupID:          201,
		StartsAt:         startsAt,
		ExpiresAt:        expiresAt,
		DailyWindowStart: &dailyWindowStart,
		DailyUsageUSD:    4.2,
		Group:            &Group{},
	}

	err := svc.CheckAndResetWindows(context.Background(), sub)

	require.NoError(t, err)
	require.True(t, repo.resetDailyCalled, "普通1天订阅应维持原有数据库日重置行为")
	require.Equal(t, 0.0, sub.DailyUsageUSD)
	require.NotNil(t, sub.DailyWindowStart)
}

func TestCheckAndResetWindows_MultiDaySubscriptionStillResetsDaily(t *testing.T) {
	repo := &dailyResetPolicyUserSubRepoStub{}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	now := time.Now()
	startsAt := now.Add(-60 * time.Hour)
	expiresAt := startsAt.AddDate(0, 0, 7)
	dailyWindowStart := startOfDay(now.Add(-48 * time.Hour))

	sub := &UserSubscription{
		ID:               2,
		UserID:           101,
		GroupID:          201,
		StartsAt:         startsAt,
		ExpiresAt:        expiresAt,
		DailyWindowStart: &dailyWindowStart,
		DailyUsageUSD:    4.2,
	}

	err := svc.CheckAndResetWindows(context.Background(), sub)

	require.NoError(t, err)
	require.True(t, repo.resetDailyCalled, "多天订阅应维持原有数据库日重置行为")
	require.Equal(t, 0.0, sub.DailyUsageUSD)
	require.NotNil(t, sub.DailyWindowStart)
}

func TestNormalizeExpiredWindows_SingleDayCardSkipsDailyReset(t *testing.T) {
	now := time.Now()
	startsAt := now.Add(-20 * time.Hour)
	expiresAt := startsAt.AddDate(0, 0, 1)
	dailyWindowStart := startOfDay(startsAt)

	subs := []UserSubscription{
		{
			StartsAt:         startsAt,
			ExpiresAt:        expiresAt,
			DailyWindowStart: &dailyWindowStart,
			DailyUsageUSD:    3.6,
			Group:            singleDayCardTestGroup(10),
		},
	}

	normalizeExpiredWindows(subs)

	require.NotNil(t, subs[0].DailyWindowStart)
	require.Equal(t, 3.6, subs[0].DailyUsageUSD, "单日卡在返回数据归一化时也不应被重置")
}

func TestNormalizeExpiredWindows_RegularOneDaySubscriptionStillResetsDaily(t *testing.T) {
	now := time.Now()
	startsAt := now.Add(-20 * time.Hour)
	expiresAt := startsAt.AddDate(0, 0, 1)
	dailyWindowStart := startOfDay(startsAt)

	subs := []UserSubscription{
		{
			StartsAt:         startsAt,
			ExpiresAt:        expiresAt,
			DailyWindowStart: &dailyWindowStart,
			DailyUsageUSD:    3.6,
			Group:            &Group{},
		},
	}

	normalizeExpiredWindows(subs)

	require.Nil(t, subs[0].DailyWindowStart)
	require.Equal(t, 0.0, subs[0].DailyUsageUSD, "普通1天订阅在返回数据归一化时仍应被重置")
}
