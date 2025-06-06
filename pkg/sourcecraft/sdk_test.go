package sourcecraft_test

import (
	"testing"

	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
)

func TestParseRepoNameFromURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Default repository",
			url:  "https://git.o.cloud.yandex.net/test/sourcecraft-actions.git",
			want: "sourcecraft-actions",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourcecraft.ParseRepoNameFromURL(tt.url); got != tt.want {
				t.Errorf("GetSourcecraftRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseRepoOwnerFromURL(t *testing.T) {
	type args struct {
		repoURL string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Default repository",
			args: args{repoURL: "https://git.o.cloud.yandex.net/test/sourcecraft-actions.git"},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourcecraft.ParseRepoOwnerFromURL(tt.args.repoURL); got != tt.want {
				t.Errorf("parseRepoOwnerFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
