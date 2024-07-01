package app

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestURLShortenerApp_handleRequest(t *testing.T) {
	type fields struct {
		urls map[string]string
	}

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}

	type want struct {
		status         int
		contentType    string
		locationHeader string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "add success",
			fields: fields{
				urls: map[string]string{},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`https://google.com`)),
			},
			want: want{
				status:      http.StatusCreated,
				contentType: "plain/text",
			},
		},
		{
			name: "get success",
			fields: fields{
				urls: map[string]string{
					"abc": "https://google.com",
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/abc", nil),
			},
			want: want{
				status:         http.StatusTemporaryRedirect,
				contentType:    "plain/text",
				locationHeader: "https://google.com",
			},
		},
		{
			name: "not supported http method",
			fields: fields{
				urls: map[string]string{},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPut, "/abc", nil),
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &URLShortenerApp{
				urls: tt.fields.urls,
			}

			a.handleRequest(tt.args.w, tt.args.r)

			assert.Equal(t, tt.want.status, tt.args.w.Code)
			if tt.want.status == http.StatusCreated {
				assert.Equal(t, tt.want.contentType, tt.args.w.Header().Get("Content-Type"))
				assert.NotEmpty(t, tt.args.w.Body.String())
			} else if tt.want.status == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.want.contentType, tt.args.w.Header().Get("Content-Type"))
				assert.Equal(t, tt.want.locationHeader, tt.args.w.Header().Get("Location"))
			}
		})
	}
}
