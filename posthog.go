package posthog

import (
	"context"
	"fmt"

	"github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/posthog/posthog-go"
)

const DistinctIdKey = openfeature.TargetingKey

var _ openfeature.FeatureProvider = (*Provider)(nil)

type Provider struct {
	client posthog.Client
}

// New creates an OpenFeature provider backed by PostHog.
func New(client posthog.Client) *Provider {
	return &Provider{
		client: client,
	}
}

func (p *Provider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{
		Name: "PostHog",
	}
}

// Hooks are not currently implemented so nil is returned.
func (p *Provider) Hooks() []openfeature.Hook {
	return nil
}

func (p *Provider) BooleanEvaluation(
	ctx context.Context,
	flag string,
	defaultValue bool,
	evalCtx openfeature.FlattenedContext,
) openfeature.BoolResolutionDetail {
	distinctID, resDetails := extractDistinctID(evalCtx)
	if resDetails != nil {
		return openfeature.BoolResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: *resDetails,
		}
	}

	// Notes:
	// 1. No error if api key is incorrect. It does log "posthog 2023/09/07 15:43:34 ERROR: Error calling /decide/" which is not that useful.
	// 2. It would be nice if this method returned a boolean
	// 3. If the flag doesn't exist we have no way of knowing this and we just get back false
	// 4. Every request logs "posthog 2023/09/07 15:04:36 ERROR: Unable to fetch feature flags%!(EXTRA <nil>)"
	// 5. Not exactly related here: but if I simply specify a non-empty personal api key when constructing the client, this request succeeds
	resp, err := p.client.IsFeatureEnabled(posthog.FeatureFlagPayload{
		Key:        flag,
		DistinctId: distinctID,
	})
	if err != nil {
		return openfeature.BoolResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: fromPostHogError(err),
		}
	}

	respBool, ok := resp.(bool)
	if !ok {
		return openfeature.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError(
					fmt.Sprintf("unable to convert response to boolean: %v", resp),
				),
				Reason: openfeature.ErrorReason,
			},
		}
	}

	return openfeature.BoolResolutionDetail{
		Value: respBool,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			// Note: we don't actually know if this was a TargetingMatchReason,
			// there is no way to tell if the flag exists or not.
			Reason: openfeature.TargetingMatchReason,
		},
	}
}

// StringEvaluation will always return the default value and PostHog does not
// have string evaluation.
func (p *Provider) StringEvaluation(
	ctx context.Context,
	flag string,
	defaultValue string,
	evalCtx openfeature.FlattenedContext,
) openfeature.StringResolutionDetail {
	return openfeature.StringResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason:          openfeature.DefaultReason,
			ResolutionError: openfeature.NewGeneralResolutionError("string evaluation not implemented"),
		},
	}
}

// FloatEvaluation will always return the default value and PostHog does not
// have float evaluation.
func (p *Provider) FloatEvaluation(
	ctx context.Context,
	flag string,
	defaultValue float64,
	evalCtx openfeature.FlattenedContext,
) openfeature.FloatResolutionDetail {
	return openfeature.FloatResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason:          openfeature.DefaultReason,
			ResolutionError: openfeature.NewGeneralResolutionError("float evaluation not implemented"),
		},
	}
}

// IntEvaluation will always return the default value and PostHog does not
// have int evaluation.
func (p *Provider) IntEvaluation(
	ctx context.Context,
	flag string,
	defaultValue int64,
	evalCtx openfeature.FlattenedContext,
) openfeature.IntResolutionDetail {
	return openfeature.IntResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason:          openfeature.DefaultReason,
			ResolutionError: openfeature.NewGeneralResolutionError("int evaluation not implemented"),
		},
	}
}

// ObjectEvaluation will always return the default value and PostHog does not
// have object evaluation.
func (p *Provider) ObjectEvaluation(
	ctx context.Context,
	flag string,
	defaultValue any,
	evalCtx openfeature.FlattenedContext,
) openfeature.InterfaceResolutionDetail {
	return openfeature.InterfaceResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason:          openfeature.DefaultReason,
			ResolutionError: openfeature.NewGeneralResolutionError("object evaluation not implemented"),
		},
	}
}

func extractDistinctID(evalCtx openfeature.FlattenedContext) (string, *openfeature.ProviderResolutionDetail) {
	for key, val := range evalCtx {
		if key == DistinctIdKey {
			v, ok := val.(string)
			if !ok {
				return "", &openfeature.ProviderResolutionDetail{
					ResolutionError: openfeature.NewTargetingKeyMissingResolutionError(
						"value of targetingKey/distinctId cannnot be converted to string",
					),
					Reason: openfeature.ErrorReason,
				}
			}
			return v, nil
		}
	}

	return "", &openfeature.ProviderResolutionDetail{
		ResolutionError: openfeature.NewTargetingKeyMissingResolutionError(
			"no targetingKey/distinctId",
		),
		Reason: openfeature.ErrorReason,
	}
}

func fromPostHogError(err error) openfeature.ProviderResolutionDetail {
	return openfeature.ProviderResolutionDetail{
		ResolutionError: openfeature.NewGeneralResolutionError(
			fmt.Sprintf("posthog client error: %s", err.Error()),
		),
		Reason: openfeature.ErrorReason,
	}
}
