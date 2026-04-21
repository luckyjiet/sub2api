package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNormalizeResponsesRequestServiceTier(t *testing.T) {
	t.Parallel()

	req := &apicompat.ResponsesRequest{ServiceTier: " fast "}
	normalizeResponsesRequestServiceTier(req)
	require.Equal(t, "priority", req.ServiceTier)

	req.ServiceTier = "flex"
	normalizeResponsesRequestServiceTier(req)
	require.Equal(t, "flex", req.ServiceTier)

	req.ServiceTier = "default"
	normalizeResponsesRequestServiceTier(req)
	require.Empty(t, req.ServiceTier)
}

func TestNormalizeResponsesBodyServiceTier(t *testing.T) {
	t.Parallel()

	body, tier, err := normalizeResponsesBodyServiceTier([]byte(`{"model":"gpt-5.1","service_tier":"fast"}`))
	require.NoError(t, err)
	require.Equal(t, "priority", tier)
	require.Equal(t, "priority", gjson.GetBytes(body, "service_tier").String())

	body, tier, err = normalizeResponsesBodyServiceTier([]byte(`{"model":"gpt-5.1","service_tier":"flex"}`))
	require.NoError(t, err)
	require.Equal(t, "flex", tier)
	require.Equal(t, "flex", gjson.GetBytes(body, "service_tier").String())

	body, tier, err = normalizeResponsesBodyServiceTier([]byte(`{"model":"gpt-5.1","service_tier":"default"}`))
	require.NoError(t, err)
	require.Empty(t, tier)
	require.False(t, gjson.GetBytes(body, "service_tier").Exists())
}

func TestNormalizeResponsesPromptCacheKey_CamelCaseToSnakeCase(t *testing.T) {
	body := []byte(`{"model":"gpt-5.4","input":[],"promptCacheKey":" ses_camel_1 "}`)
	normalized, err := normalizeResponsesPromptCacheKey(body, "ses_camel_1")
	require.NoError(t, err)
	require.Equal(t, "ses_camel_1", gjson.GetBytes(normalized, "prompt_cache_key").String())
	require.False(t, gjson.GetBytes(normalized, "promptCacheKey").Exists())
}

func TestNormalizeResponsesPromptCacheKey_KeepSnakeCase(t *testing.T) {
	body := []byte(`{"model":"gpt-5.4","input":[],"prompt_cache_key":"ses_snake_1"}`)
	normalized, err := normalizeResponsesPromptCacheKey(body, "ses_snake_1")
	require.NoError(t, err)
	require.Equal(t, "ses_snake_1", gjson.GetBytes(normalized, "prompt_cache_key").String())
	require.False(t, gjson.GetBytes(normalized, "promptCacheKey").Exists())
}

func TestNormalizeResponsesPromptCacheKey_EmptyBody(t *testing.T) {
	normalized, err := normalizeResponsesPromptCacheKey(nil, "ses_any")
	require.NoError(t, err)
	require.Nil(t, normalized)
}
