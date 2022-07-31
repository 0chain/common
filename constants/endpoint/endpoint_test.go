package endpoint

import (
	"testing"
)

func TestEndpoint_FormattedPath(t *testing.T) {
	type args struct {
		format Format
	}
	tests := []struct {
		name string
		path string
		args args
		want string
	}{
		{
			name: "NoSlash",
			path: "/v1/test/",
			args: args{format: NoSlash},
			want: "v1/test",
		},
		{
			name: "TrailingSlash",
			path: "/v1/test/",
			args: args{format: TrailingSlash},
			want: "v1/test/",
		},
		{
			name: "LeadingSlash",
			path: "/v1/test/",
			args: args{format: LeadingSlash},
			want: "/v1/test",
		},
		{
			name: "LeadingAndTrailingSlash",
			path: "/v1/test/",
			args: args{format: LeadingAndTrailingSlash},
			want: "/v1/test/",
		},
		{
			name: "Trimmed",
			path: "   /v1/test/   ",
			args: args{format: LeadingAndTrailingSlash},
			want: "/v1/test/",
		},
		{
			name: "MultiSlash",
			path: "   //v1/test//   ",
			args: args{format: NoSlash},
			want: "v1/test",
		},
		{
			name: "SlashAndTrim",
			path: "  / v1/test /",
			args: args{format: NoSlash},
			want: "v1/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.path)
			if got := m.FormattedPath(tt.args.format); got != tt.want {
				t.Errorf("FormattedPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
