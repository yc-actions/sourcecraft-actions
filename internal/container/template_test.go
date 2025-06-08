package container

import (
	"reflect"
	"testing"
)

func TestReplaceVariablesInSpec(t *testing.T) {
	type args struct {
		specContent []byte
		variables   map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "No variables",
			args: args{
				specContent: []byte("openapi: 3.0.0\ninfo:\n  title: Sample API\n  version: 1.0.0"),
			},
			want:    []byte("openapi: 3.0.0\ninfo:\n  title: Sample API\n  version: 1.0.0"),
			wantErr: false,
		},
		{
			name: "With variables",
			args: args{
				specContent: []byte("openapi: 3.0.0\ninfo:\n  title: {{.Title}}\n  version: {{.Version}}"),
				variables: map[string]string{
					"Title":   "Sample API",
					"Version": "1.0.0",
				},
			},
			want: []byte("openapi: 3.0.0\ninfo:\n  title: Sample API\n  version: 1.0.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceVariablesInSpec(tt.args.specContent, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceVariablesInSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReplaceVariablesInSpec() got = %v, want %v", got, tt.want)
			}
		})
	}
}
