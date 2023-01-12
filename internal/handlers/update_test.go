package handlers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateHandler_Update(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Верные аргументы",
			request: "/update/counter/testCount/100",
			want: want{
				code: 200,
			},
		},
		{
			name:    "Неверное количество переданных аргументов",
			request: "/update/counter/testCount/",
			want: want{
				code: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := &UpdateHandler{}
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equalf(t, tt.want.code, res.StatusCode,
				"Полученный статус код (%d) не соответствует ожидаемому (%d)", res.StatusCode, tt.want.code)
		})
	}
}
