# Unofficial PostHog OpenFeature Provider for Go

This is an experimental unofficial [PostHog](https://posthog.com/)
[OpenFeature](https://openfeature.dev/) provider for Go. It uses the
[official PostHog SDK](https://github.com/PostHog/posthog-go) under the hood.

## Usage

```go
import (
	"context"
	"fmt"

	pgprovider "github.com/craigpastro/posthog-openfeature-provider-go"
	"github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/posthog/posthog-go"
)

func main() {
    phClient, err := posthog.NewWithConfig(
		"<POSTHOG_API_KEY>",
		posthog.Config{
			PersonalApiKey: "<PERSONAL_API_KEY>",
			Endpoint:       "https://app.posthog.com",
		},
	)
	if err != nil {
		panic(err)
	}
	defer phClient.Close()

	err = openfeature.SetProvider(pgprovider.New(phClient))
	if err != nil {
		panic(err)
	}

	client := openfeature.NewClient("app")

	v2_enabled, err := client.BooleanValue(
		context.Background(),
		"v2_enabled",
		false,
		openfeature.NewEvaluationContext("my_distinct_id", nil),
	)
	if err != nil {
		panic(err)
	}

	if v2_enabled {
		fmt.Println("v2 is enabled")
	} else {
		fmt.Println("v2 is NOT enabled")
	}
}
```
