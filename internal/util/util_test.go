package util

import "testing"

func TestGetShortURL(t *testing.T) {
	type args struct {
		baseURL  string
		shortURI string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				baseURL:  "http://localhost:8080",
				shortURI: "abc",
			},
			want: "http://localhost:8080/abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetShortURL(tt.args.baseURL, tt.args.shortURI); got != tt.want {
				t.Errorf("GetShortURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandStringRunes(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				n: 8,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandStringRunes(tt.args.n); len(got) != tt.args.n {
				t.Errorf("RandStringRunes() = %v, want %v", len(got), tt.args.n)
			}
		})
	}
}
