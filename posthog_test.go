package posthog

import (
	"context"
	"errors"
	"testing"

	"github.com/craigpastro/posthog-openfeature-provider-go/mocks"
	"github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMeta(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPhClient := mocks.NewMockClient(ctrl)
	client := New(mockPhClient)

	t.Run("metadataIsPostHog", func(t *testing.T) {
		require.Equal(t, client.Metadata().Name, "PostHog")
	})

	t.Run("hooksAreNil", func(t *testing.T) {
		require.Nil(t, client.Hooks())
	})
}

func TestBooleanEvaluation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPhClient := mocks.NewMockClient(ctrl)
	client := New(mockPhClient)

	ctx := context.Background()
	evalCtx := map[string]any{DistinctIdKey: "my_distinct_id"}

	t.Run("noDistinctIdFails", func(t *testing.T) {
		detail := client.BooleanEvaluation(ctx, "flag", false, openfeature.FlattenedContext{})

		require.False(t, detail.Value)
		require.Equal(t, detail.Reason, openfeature.ErrorReason)
		require.ErrorContains(t, detail.ResolutionError, "no targetingKey/distinctId")
	})

	t.Run("isFeaturedEnabledFails", func(t *testing.T) {
		mockPhClient.EXPECT().IsFeatureEnabled(gomock.Any()).Return(nil, errors.New("no feature flag"))

		detail := client.BooleanEvaluation(ctx, "flag", false, evalCtx)
		require.False(t, detail.Value)
		require.Equal(t, detail.Reason, openfeature.ErrorReason)
		require.ErrorContains(t, detail.ResolutionError, "posthog client error: no feature flag")
	})

	t.Run("responseIsNotABool", func(t *testing.T) {
		mockPhClient.EXPECT().IsFeatureEnabled(gomock.Any()).Return("hello", nil)

		detail := client.BooleanEvaluation(ctx, "flag", false, evalCtx)
		require.False(t, detail.Value)
		require.Equal(t, detail.Reason, openfeature.ErrorReason)
		require.ErrorContains(t, detail.ResolutionError, "unable to convert response to boolean")
	})

	t.Run("falseResponseNoError", func(t *testing.T) {
		mockPhClient.EXPECT().IsFeatureEnabled(gomock.Any()).Return(false, nil)

		detail := client.BooleanEvaluation(ctx, "flag", false, evalCtx)
		require.False(t, detail.Value)
		require.Equal(t, detail.Reason, openfeature.UnknownReason) // for now
		require.Equal(t, detail.ResolutionError, openfeature.ResolutionError{})
	})

	t.Run("trueResponseNoError", func(t *testing.T) {
		mockPhClient.EXPECT().IsFeatureEnabled(gomock.Any()).Return(true, nil)

		detail := client.BooleanEvaluation(ctx, "flag", false, evalCtx)
		require.True(t, detail.Value)
		require.Equal(t, detail.Reason, openfeature.UnknownReason) // for now
		require.Equal(t, detail.ResolutionError, openfeature.ResolutionError{})
	})
}

func TestUnimplemented(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPhClient := mocks.NewMockClient(ctrl)
	client := New(mockPhClient)

	ctx := context.Background()

	t.Run("stringEvaluationIsUnimplemented", func(t *testing.T) {
		defaultValue := "hello"
		detail := client.StringEvaluation(ctx, "flag", defaultValue, openfeature.FlattenedContext{})
		require.Equal(t, defaultValue, detail.Value)
		require.Equal(t, openfeature.DefaultReason, detail.Reason)
		require.ErrorContains(t, detail.ResolutionError, "string evaluation not implemented")
	})

	t.Run("floatEvaluationIsUnimplemented", func(t *testing.T) {
		defaultValue := 3.14
		detail := client.FloatEvaluation(ctx, "flag", defaultValue, openfeature.FlattenedContext{})
		require.Equal(t, defaultValue, detail.Value)
		require.Equal(t, openfeature.DefaultReason, detail.Reason)
		require.ErrorContains(t, detail.ResolutionError, "float evaluation not implemented")
	})

	t.Run("intEvaluationIsUnimplemented", func(t *testing.T) {
		var defaultValue int64 = 3
		detail := client.IntEvaluation(ctx, "flag", defaultValue, openfeature.FlattenedContext{})
		require.Equal(t, defaultValue, detail.Value)
		require.Equal(t, openfeature.DefaultReason, detail.Reason)
		require.ErrorContains(t, detail.ResolutionError, "int evaluation not implemented")
	})

	t.Run("objectEvaluationIsUnimplemented", func(t *testing.T) {
		defaultValue := map[string]any{"foo": "bar"}
		detail := client.ObjectEvaluation(ctx, "flag", defaultValue, openfeature.FlattenedContext{})
		require.Equal(t, defaultValue, detail.Value)
		require.Equal(t, openfeature.DefaultReason, detail.Reason)
		require.ErrorContains(t, detail.ResolutionError, "object evaluation not implemented")
	})
}
