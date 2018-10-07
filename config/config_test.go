package config

import (
	"os"
	"testing"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test get version return server version",
			want: "v1.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetVersion(); got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetActualValue(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name       string
		args       args
		beforeFunc func() error
		afterFunc  func() error
		want       string
	}{
		{
			name: "GetActualValue without env var",
			args: args{
				val: "test_env",
			},
			want: "test_env",
		},
		{
			name: "GetActualValue with env var",
			args: args{
				val: "_dummy_key_",
			},
			beforeFunc: func() error {
				return os.Setenv("dummy_key", "dummy_value")
			},
			afterFunc: func() error {
				return os.Remove("dummy_key")
			},
			want: "dummy_value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			if got := GetActualValue(tt.args.val); got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPrefixAndSuffix(t *testing.T) {
	type args struct {
		str  string
		pref string
		suf  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Check true prefix and suffix",
			args: args{
				str:  "_dummy_",
				pref: "_",
				suf:  "_",
			},
			want: true,
		},
		{
			name: "Check false prefix and suffix",
			args: args{
				str:  "dummy",
				pref: "_",
				suf:  "_",
			},
			want: false,
		},
		{
			name: "Check true prefix but false suffix",
			args: args{
				str:  "_dummy",
				pref: "_",
				suf:  "_",
			},
			want: false,
		},
		{
			name: "Check false prefix but true suffix",
			args: args{
				str:  "dummy_",
				pref: "_",
				suf:  "_",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPrefixAndSuffix(tt.args.str, tt.args.pref, tt.args.suf); got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
