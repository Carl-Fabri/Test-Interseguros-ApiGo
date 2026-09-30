package domain

import "math"

// QRResult es la factorización QR completa de una matriz A (m×n): A = Q·R, donde
// Q (m×m) es ortogonal y R (m×n) es triangular superior.
type QRResult struct {
	Q Matrix
	R Matrix
}

// QRGivens calcula la factorización QR completa de a mediante rotaciones de Givens.
//
// Algoritmo: se recorre cada columna j y, de abajo hacia arriba, se anula el elemento
// R[i][j] rotando el par de filas adyacentes (i-1, i) con la rotación
//
//	G = [ c  s ]   con  c = x/r, s = y/r, r = hypot(x, y)
//	    [-s  c ]
//
// La misma rotación se acumula en Qᵀ, de modo que Qᵀ·A = R y, al final, Q = (Qᵀ)ᵀ.
//
// Por qué Givens: es numéricamente estable (rotaciones ortogonales), funciona con
// cualquier matriz rectangular (m ≥ n o m < n) y es la "rotación de la matriz" que
// describe el enunciado. Costo: O(m·n·min(m,n)) para R y O(m²·min(m,n)) para Q.
//
// Para que el resultado sea determinista se normalizan los signos de modo que la
// diagonal de R sea no negativa (la QR es única con esa condición si A tiene rango completo).
// a no se modifica; se asume que ya fue validada con Validate.
func QRGivens(a Matrix) QRResult {
	// * Método central del servicio: todo el resto (handler, servicio, cliente) orquesta esta función.
	m, n := a.Dimensions()
	r := a.Clone()
	qt := Identity(m)

	for j := 0; j < n && j < m-1; j++ {
		for i := m - 1; i > j; i-- {
			x, y := r[i-1][j], r[i][j]
			// ? Saltar ceros ya existentes hace a Givens eficiente en matrices dispersas (ventaja frente a Householder).
			if y == 0 {
				continue // ya es cero: no hace falta rotar
			}
			// ! math.Hypot y no math.Sqrt(x*x+y*y): con valores ~1e200 el cuadrado desborda a +Inf.
			h := math.Hypot(x, y)
			c, s := x/h, y/h

			rotateRows(r[i-1], r[i], c, s, j)
			// * Cero exacto (no ~1e-17): R queda estrictamente triangular y la detección de diagonal es confiable.
			r[i][j] = 0
			rotateRows(qt[i-1], qt[i], c, s, 0)
		}
	}

	// ? Signos normalizados (diag(R) ≥ 0): la QR es única para rango completo → salida determinista y comparable.
	normalizeSigns(qt, r, min(m, n))
	return QRResult{Q: cleanNegativeZeros(qt.Transpose()), R: cleanNegativeZeros(r)}
}

// rotateRows aplica la rotación de Givens a las filas top y bottom desde la columna from.
func rotateRows(top, bottom []float64, c, s float64, from int) {
	for k := from; k < len(top); k++ {
		t, b := top[k], bottom[k]
		top[k] = c*t + s*b
		bottom[k] = -s*t + c*b
	}
}

// normalizeSigns deja la diagonal de R no negativa. Cambiar el signo de la fila k de R
// y de la fila k de Qᵀ mantiene intacta la igualdad Qᵀ·A = R.
func normalizeSigns(qt, r Matrix, diag int) {
	for k := range diag {
		if r[k][k] < 0 {
			negate(r[k])
			negate(qt[k])
		}
	}
}

func negate(row []float64) {
	for i := range row {
		row[i] = -row[i]
	}
}

// cleanNegativeZeros reemplaza -0 por 0 para que la salida JSON sea limpia.
func cleanNegativeZeros(m Matrix) Matrix {
	for _, row := range m {
		for j, v := range row {
			if v == 0 {
				row[j] = 0
			}
		}
	}
	return m
}
