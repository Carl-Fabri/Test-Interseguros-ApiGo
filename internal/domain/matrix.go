// Package domain contiene el núcleo matemático de api-go: el tipo Matrix, su validación
// y la factorización QR. No depende de HTTP, Fiber ni de ningún otro paquete del servicio.
package domain

import (
	"fmt"
	"math"
)

// Matrix es una matriz densa en orden de filas: m[i][j] es la fila i, columna j.
type Matrix [][]float64

// ValidationError indica que una matriz de entrada no cumple las reglas del dominio.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func invalid(format string, args ...any) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

// Dimensions devuelve la cantidad de filas y columnas (asume una matriz ya validada).
func (m Matrix) Dimensions() (rows, cols int) {
	if len(m) == 0 {
		return 0, 0
	}
	return len(m), len(m[0])
}

// Validate comprueba que la matriz sea no vacía, rectangular, con valores finitos
// y con a lo sumo maxDimension filas y columnas.
func (m Matrix) Validate(maxDimension int) error {
	// * Única validación de entrada del dominio: el handler solo parsea JSON, las reglas viven aquí.
	// ! El límite de tamaño protege al servicio de matrices gigantes (costo O(m·n·min(m,n))).
	if len(m) == 0 {
		return invalid("la matriz no puede estar vacía")
	}
	cols := len(m[0])
	if cols == 0 {
		return invalid("las filas de la matriz no pueden estar vacías")
	}
	if len(m) > maxDimension || cols > maxDimension {
		return invalid("la matriz excede el tamaño máximo de %dx%d", maxDimension, maxDimension)
	}
	for i, row := range m {
		if len(row) != cols {
			return invalid("la matriz debe ser rectangular: la fila %d tiene %d columnas y se esperaban %d", i, len(row), cols)
		}
		for j, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return invalid("el valor en [%d][%d] no es un número finito", i, j)
			}
		}
	}
	return nil
}

// Clone devuelve una copia profunda de la matriz.
func (m Matrix) Clone() Matrix {
	out := make(Matrix, len(m))
	for i, row := range m {
		out[i] = append([]float64(nil), row...)
	}
	return out
}

// Transpose devuelve la transpuesta de la matriz.
func (m Matrix) Transpose() Matrix {
	rows, cols := m.Dimensions()
	out := Zeros(cols, rows)
	for i := range rows {
		for j := range cols {
			out[j][i] = m[i][j]
		}
	}
	return out
}

// Zeros crea una matriz rows×cols llena de ceros.
func Zeros(rows, cols int) Matrix {
	out := make(Matrix, rows)
	for i := range out {
		out[i] = make([]float64, cols)
	}
	return out
}

// Identity crea la matriz identidad n×n.
func Identity(n int) Matrix {
	out := Zeros(n, n)
	for i := range n {
		out[i][i] = 1
	}
	return out
}
