package overleash

import (
	"reflect"
	"testing"
	"time"
)

func TestFeatureFlags_Get(t *testing.T) {
	type args struct {
		key string
	}

	feature := Feature{
		Name:         "flag",
		Type:         "",
		Description:  "",
		Enabled:      false,
		Strategies:   nil,
		CreatedAt:    &time.Time{},
		Strategy:     "",
		Variants:     nil,
		Dependencies: nil,
		SearchTerm:   "",
	}

	tests := []struct {
		name    string
		f       FeatureFlags
		args    args
		want    Feature
		wantErr bool
	}{
		{
			name: "Find feature flag",
			f: FeatureFlags{
				feature,
			},
			args: args{
				key: "flag",
			},
			want:    feature,
			wantErr: false,
		},
		{
			name: "Find feature flag",
			f: FeatureFlags{
				feature,
			},
			args: args{
				key: "flag123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}
