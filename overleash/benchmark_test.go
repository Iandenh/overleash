package overleash

import (
	"fmt"
	"testing"

	"github.com/Iandenh/overleash/config"
)

func BenchmarkOverrideUpdate(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("flags=%d", size), func(b *testing.B) {
			o := NewOverleash(&config.Config{Token: "*:dev.benchmark", Storage: "null", Reload: "0"})
			file := FeatureFile{Version: 2, Features: make(FeatureFlags, size)}
			for i := range file.Features {
				file.Features[i] = Feature{
					Name:    fmt.Sprintf("feature-%04d", i),
					Enabled: true,
					Project: "default",
					Strategies: []Strategy{{
						Name:       "flexibleRollout",
						Parameters: ParameterMap{"rollout": "50", "stickiness": "userId"},
						Constraints: []Constraint{{
							ContextName: "region",
							Operator:    OperatorIn,
							Values:      []string{"eu", "us"},
						}},
					}},
				}
			}
			o.LoadFeatureFile(file)
			b.ReportAllocs()
			b.ResetTimer()
			enabled := true
			for b.Loop() {
				o.AddOverride("feature-0000", enabled)
				enabled = !enabled
			}
		})
	}
}
