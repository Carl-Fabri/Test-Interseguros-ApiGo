package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elfab/retotecnico/api-go/internal/domain"
)

// fakeStats es un StatisticsClient falso que registra lo recibido.
type fakeStats struct {
	received []NamedMatrix
	result   Statistics
	err      error
}

func (f *fakeStats) Compute(_ context.Context, matrices []NamedMatrix) (Statistics, error) {
	f.received = matrices
	return f.result, f.err
}

func TestMatrixServiceAnalyze(t *testing.T) {
	stats := &fakeStats{result: Statistics{Max: 5, AnyDiagonal: true}}
	svc := NewMatrixService(stats, 10)

	got, err := svc.Analyze(context.Background(), domain.Matrix{{2, 0}, {0, 3}})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(stats.received) != 2 || stats.received[0].Name != "Q" || stats.received[1].Name != "R" {
		t.Fatalf("se esperaba enviar Q y R a estadísticas, se envió %+v", stats.received)
	}
	if got.Statistics.Max != 5 || !got.Statistics.AnyDiagonal {
		t.Errorf("las estadísticas deben propagarse sin cambios, got %+v", got.Statistics)
	}
	if got.QR.R[1][1] != 3 {
		t.Errorf("R[1][1] = %g, want 3", got.QR.R[1][1])
	}
}

func TestMatrixServiceAnalyzeErrors(t *testing.T) {
	t.Run("matriz inválida no llama a estadísticas", func(t *testing.T) {
		stats := &fakeStats{}
		_, err := NewMatrixService(stats, 10).Analyze(context.Background(), domain.Matrix{{1, 2}, {3}})

		var vErr *domain.ValidationError
		if !errors.As(err, &vErr) {
			t.Fatalf("err = %v, want ValidationError", err)
		}
		if stats.received != nil {
			t.Error("no se debe llamar a estadísticas con una matriz inválida")
		}
	})

	t.Run("propaga el error del cliente de estadísticas", func(t *testing.T) {
		stats := &fakeStats{err: ErrStatisticsTimeout}
		_, err := NewMatrixService(stats, 10).Analyze(context.Background(), domain.Matrix{{1}})
		if !errors.Is(err, ErrStatisticsTimeout) {
			t.Fatalf("err = %v, want ErrStatisticsTimeout", err)
		}
	})
}

type fakeIssuer struct{ subject string }

func (f *fakeIssuer) Issue(subject string) (string, time.Time, error) {
	f.subject = subject
	return "token-" + subject, time.Unix(100, 0), nil
}

func TestAuthServiceLogin(t *testing.T) {
	tests := []struct {
		name, user, pass string
		wantErr          error
	}{
		{"credenciales correctas", "admin", "secret", nil},
		{"contraseña incorrecta", "admin", "otra", ErrInvalidCredentials},
		{"usuario incorrecto", "otro", "secret", ErrInvalidCredentials},
		{"credenciales vacías", "", "", ErrInvalidCredentials},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issuer := &fakeIssuer{}
			got, err := NewAuthService("admin", "secret", issuer).Login(tc.user, tc.pass)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && (got.Token != "token-admin" || issuer.subject != "admin") {
				t.Errorf("token = %q, subject = %q", got.Token, issuer.subject)
			}
		})
	}
}
