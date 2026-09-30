package domain

import (
	"math"
	"testing"
)

const tolerance = 1e-9

func TestQRGivensProperties(t *testing.T) {
	tests := []struct {
		name string
		a    Matrix
	}{
		{"cuadrada 3x3", Matrix{{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}}},
		{"alta 4x2 (m > n)", Matrix{{1, 2}, {3, 4}, {5, 6}, {7, 8}}},
		{"ancha 2x4 (m < n)", Matrix{{1, 2, 3, 4}, {5, 6, 7, 8}}},
		{"un elemento negativo", Matrix{{-5}}},
		{"fila única", Matrix{{3, 4, 5}}},
		{"columna única", Matrix{{3}, {4}}},
		{"singular (filas dependientes)", Matrix{{1, 2}, {2, 4}}},
		{"con ceros que evitan rotaciones", Matrix{{2, 0}, {0, 3}}},
		{"valores grandes", Matrix{{1e150, 1e150}, {1e150, -1e150}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, n := tc.a.Dimensions()
			original := tc.a.Clone()

			got := QRGivens(tc.a)

			assertDimensions(t, "Q", got.Q, m, m)
			assertDimensions(t, "R", got.R, m, n)
			assertMatrixClose(t, "la entrada no debe modificarse", tc.a, original)
			assertMatrixRelClose(t, "Q·R debe reconstruir A", multiply(got.Q, got.R), tc.a)
			assertMatrixClose(t, "QᵀQ debe ser la identidad", multiply(got.Q.Transpose(), got.Q), Identity(m))

			for i := range m {
				for j := 0; j < min(i, n); j++ {
					if got.R[i][j] != 0 {
						t.Errorf("R[%d][%d] = %g; R debe ser triangular superior (cero exacto)", i, j, got.R[i][j])
					}
				}
			}
			for k := range min(m, n) {
				if got.R[k][k] < 0 {
					t.Errorf("R[%d][%d] = %g; la diagonal de R debe ser no negativa", k, k, got.R[k][k])
				}
			}
		})
	}
}

func TestQRGivensKnownResult(t *testing.T) {
	// Ejemplo clásico: A = [[12,-51,4],[6,167,-68],[-4,24,-41]]
	got := QRGivens(Matrix{{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}})

	wantR := Matrix{{14, 21, -14}, {0, 175, -70}, {0, 0, 35}}
	wantQ := Matrix{
		{6.0 / 7, -69.0 / 175, -58.0 / 175},
		{3.0 / 7, 158.0 / 175, 6.0 / 175},
		{-2.0 / 7, 6.0 / 35, -33.0 / 35},
	}
	assertMatrixClose(t, "R", got.R, wantR)
	assertMatrixClose(t, "Q", got.Q, wantQ)
}

func TestQRGivensDiagonalInput(t *testing.T) {
	got := QRGivens(Matrix{{2, 0}, {0, 3}})
	assertMatrixClose(t, "Q de una diagonal positiva es la identidad", got.Q, Identity(2))
	assertMatrixClose(t, "R de una diagonal positiva es la misma matriz", got.R, Matrix{{2, 0}, {0, 3}})
}

// --- helpers de prueba ---

func multiply(a, b Matrix) Matrix {
	rows, inner := a.Dimensions()
	_, cols := b.Dimensions()
	out := Zeros(rows, cols)
	for i := range rows {
		for k := range inner {
			for j := range cols {
				out[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return out
}

func assertDimensions(t *testing.T, name string, m Matrix, rows, cols int) {
	t.Helper()
	if r, c := m.Dimensions(); r != rows || c != cols {
		t.Fatalf("%s: dimensiones %dx%d, want %dx%d", name, r, c, rows, cols)
	}
}

func assertMatrixClose(t *testing.T, msg string, got, want Matrix) {
	t.Helper()
	compareMatrices(t, msg, got, want, func(g, w float64) bool { return math.Abs(g-w) <= tolerance })
}

// assertMatrixRelClose compara con tolerancia relativa, útil para valores muy grandes.
func assertMatrixRelClose(t *testing.T, msg string, got, want Matrix) {
	t.Helper()
	compareMatrices(t, msg, got, want, func(g, w float64) bool {
		return math.Abs(g-w) <= tolerance*math.Max(1, math.Max(math.Abs(g), math.Abs(w)))
	})
}

func compareMatrices(t *testing.T, msg string, got, want Matrix, eq func(g, w float64) bool) {
	t.Helper()
	gr, gc := got.Dimensions()
	wr, wc := want.Dimensions()
	if gr != wr || gc != wc {
		t.Fatalf("%s: dimensiones %dx%d, want %dx%d", msg, gr, gc, wr, wc)
	}
	for i := range gr {
		for j := range gc {
			if !eq(got[i][j], want[i][j]) {
				t.Errorf("%s: [%d][%d] = %.12g, want %.12g", msg, i, j, got[i][j], want[i][j])
			}
		}
	}
}
