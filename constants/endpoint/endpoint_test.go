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
			name: "RootLeadingAndTrailingSlash",
			path: "/",
			args: args{format: LeadingAndTrailingSlash},
			want: "/",
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

func TestEndpoint_FormattedPathWithPathVariable(t *testing.T) {
	type args struct {
		format Format
	}
	tests := []struct {
		name         string
		path         string
		pathVariable string
		args         args
		want         string
	}{
		{
			name:         "NoSlash",
			path:         "/v1/test/",
			pathVariable: "allocation",
			args:         args{format: NoSlash},
			want:         "v1/test/{allocation}",
		},
		{
			name:         "TrailingSlash",
			path:         "/v1/test/",
			pathVariable: "allocation",
			args:         args{format: TrailingSlash},
			want:         "v1/test/{allocation}/",
		},
		{
			name:         "LeadingSlash",
			path:         "/v1/test/",
			pathVariable: "allocation",
			args:         args{format: LeadingSlash},
			want:         "/v1/test/{allocation}",
		},
		{
			name:         "LeadingAndTrailingSlash",
			path:         "/v1/test/",
			pathVariable: "allocation",
			args:         args{format: LeadingAndTrailingSlash},
			want:         "/v1/test/{allocation}/",
		},
		{
			name:         "RootLeadingAndTrailingSlashNoVariable",
			path:         "/",
			pathVariable: "",
			args:         args{format: LeadingAndTrailingSlash},
			want:         "/",
		},
		{
			name:         "RootLeadingAndTrailingSlashWithVariable",
			path:         "/",
			pathVariable: "allocation",
			args:         args{format: LeadingAndTrailingSlash},
			want:         "/{allocation}/",
		},
		{
			name:         "Trimmed",
			path:         "   /v1/test/   ",
			pathVariable: " allocation ",
			args:         args{format: LeadingAndTrailingSlash},
			want:         "/v1/test/{allocation}/",
		},
		{
			name:         "MultiSlash",
			path:         "   //v1/test//   ",
			pathVariable: "  //allocation//  ",
			args:         args{format: NoSlash},
			want:         "v1/test/{allocation}",
		},
		{
			name:         "SlashAndTrim",
			path:         "  / v1/test /",
			pathVariable: " / allocation / ",
			args:         args{format: NoSlash},
			want:         "v1/test/{allocation}",
		},
		{
			name:         "NoPathVariable",
			path:         "  / v1/test /",
			pathVariable: "",
			args:         args{format: NoSlash},
			want:         "v1/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewWithPathVariable(tt.path, tt.pathVariable)
			if got := m.FormattedPathWithPathVariable(tt.args.format); got != tt.want {
				t.Errorf("FormattedPathWithPathVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}
