package domain

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		m         Matrix
		wantErrIn string
	}{
		{name: "matriz rectangular válida", m: Matrix{{1, 2, 3}, {4, 5, 6}}},
		{name: "matriz de un elemento", m: Matrix{{7}}},
		{name: "vacía", m: Matrix{}, wantErrIn: "vacía"},
		{name: "nil", m: nil, wantErrIn: "vacía"},
		{name: "fila vacía", m: Matrix{{}}, wantErrIn: "filas"},
		{name: "filas de distinto largo", m: Matrix{{1, 2}, {3}}, wantErrIn: "rectangular"},
		{name: "valor NaN", m: Matrix{{1, math.NaN()}}, wantErrIn: "finito"},
		{name: "valor infinito", m: Matrix{{math.Inf(1)}}, wantErrIn: "finito"},
		{name: "excede filas", m: Zeros(4, 1), wantErrIn: "tamaño máximo"},
		{name: "excede columnas", m: Zeros(1, 4), wantErrIn: "tamaño máximo"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.m.Validate(3)
			if tc.wantErrIn == "" {
				if err != nil {
					t.Fatalf("error inesperado: %v", err)
				}
				return
			}
			var vErr *ValidationError
			if !errors.As(err, &vErr) || !strings.Contains(err.Error(), tc.wantErrIn) {
				t.Fatalf("err = %v, se esperaba ValidationError con %q", err, tc.wantErrIn)
			}
		})
	}
}

func TestTransposeAndClone(t *testing.T) {
	m := Matrix{{1, 2, 3}, {4, 5, 6}}

	got := m.Transpose()
	want := Matrix{{1, 4}, {2, 5}, {3, 6}}
	assertMatrixClose(t, "Transpose", got, want)

	clone := m.Clone()
	clone[0][0] = 99
	if m[0][0] != 1 {
		t.Error("Clone debe devolver una copia independiente")
	}
}
