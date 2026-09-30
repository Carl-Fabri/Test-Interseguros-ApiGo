package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elfab/retotecnico/api-go/internal/auth"
	"github.com/elfab/retotecnico/api-go/internal/domain"
	"github.com/elfab/retotecnico/api-go/internal/service"
)

var sample = []service.NamedMatrix{{Name: "Q", Values: domain.Matrix{{1, 0}, {0, 1}}}}

func TestComputeSuccess(t *testing.T) {
	var gotAuth string
	var gotBody statisticsRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != statisticsPath {
			t.Errorf("petición inesperada %s %s", r.Method, r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"max":1,"min":0,"average":0.5,"sum":2,"count":4,"anyDiagonal":true,
			"matrices":[{"name":"Q","isDiagonal":true}]}`))
	}))
	defer srv.Close()

	ctx := auth.WithToken(context.Background(), "abc")
	got, err := NewStatisticsHTTPClient(srv.URL, time.Second).Compute(ctx, sample)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if gotAuth != "Bearer abc" {
		t.Errorf("Authorization = %q, se debe reenviar el token del usuario", gotAuth)
	}
	if len(gotBody.Matrices) != 1 || gotBody.Matrices[0].Name != "Q" {
		t.Errorf("body enviado = %+v", gotBody)
	}
	want := service.Statistics{Max: 1, Min: 0, Average: 0.5, Sum: 2, Count: 4, AnyDiagonal: true,
		Matrices: []service.MatrixDiagonal{{Name: "Q", IsDiagonal: true}}}
	if got.Max != want.Max || got.Count != want.Count || !got.AnyDiagonal || len(got.Matrices) != 1 {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestComputeFailures(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		timeout time.Duration
		wantErr error
	}{
		{
			name:    "api-node responde error",
			handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusUnauthorized) },
			timeout: time.Second, wantErr: service.ErrStatisticsUnavailable,
		},
		{
			name:    "respuesta que no es JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) },
			timeout: time.Second, wantErr: service.ErrStatisticsUnavailable,
		},
		{
			name:    "api-node tarda más que el timeout",
			handler: func(_ http.ResponseWriter, _ *http.Request) { time.Sleep(200 * time.Millisecond) },
			timeout: 20 * time.Millisecond, wantErr: service.ErrStatisticsTimeout,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			_, err := NewStatisticsHTTPClient(srv.URL, tc.timeout).Compute(context.Background(), sample)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}

	t.Run("api-node caído", func(t *testing.T) {
		srv := httptest.NewServer(http.NotFoundHandler())
		url := srv.URL
		srv.Close() // puerto cerrado: conexión rechazada

		_, err := NewStatisticsHTTPClient(url, time.Second).Compute(context.Background(), sample)
		if !errors.Is(err, service.ErrStatisticsUnavailable) {
			t.Fatalf("err = %v, want ErrStatisticsUnavailable", err)
		}
	})
}
