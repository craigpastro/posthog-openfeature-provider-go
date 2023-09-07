.PHONY: generate-mocks
generate-mocks:
	mockgen --destination mocks/client.go --package=mocks --build_flags=--mod=mod github.com/posthog/posthog-go Client
