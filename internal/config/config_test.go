package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// validEnv devuelve las variables mínimas obligatorias, más los overrides indicados.
func validEnv(overrides map[string]string) map[string]string {
	env := map[string]string{
		"JWT_SECRET":    strings.Repeat("s", 32),
		"AUTH_USERNAME": "admin",
		"AUTH_PASSWORD": "admin123",
	}
	for k, v := range overrides {
		env[k] = v
	}
	return env
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		want      Config
		wantErrIn string // fragmento esperado en el error; vacío si no se espera error
	}{
		{
			name: "aplica valores por defecto con solo las obligatorias",
			env:  validEnv(nil),
			want: Config{
				Port: "8080", NodeAPIURL: "http://localhost:3000", NodeAPITimeout: 5 * time.Second,
				JWTSecret: strings.Repeat("s", 32), JWTIssuer: "api-go", JWTAudience: "matrix-services", JWTTTL: time.Hour,
				AuthUsername: "admin", AuthPassword: "admin123",
				CORSAllowedOrigins: []string{"http://localhost:4200"}, MatrixMaxDimension: 100,
			},
		},
		{
			name: "lee overrides y normaliza URL y lista de orígenes",
			env: validEnv(map[string]string{
				"PORT": "9000", "NODE_API_URL": "http://api-node:3000/", "NODE_API_TIMEOUT_MS": "1500",
				"JWT_TTL_MINUTES": "15", "CORS_ALLOWED_ORIGINS": "http://a.com, ,http://b.com", "MATRIX_MAX_DIMENSION": "50",
			}),
			want: Config{
				Port: "9000", NodeAPIURL: "http://api-node:3000", NodeAPITimeout: 1500 * time.Millisecond,
				JWTSecret: strings.Repeat("s", 32), JWTIssuer: "api-go", JWTAudience: "matrix-services", JWTTTL: 15 * time.Minute,
				AuthUsername: "admin", AuthPassword: "admin123",
				CORSAllowedOrigins: []string{"http://a.com", "http://b.com"}, MatrixMaxDimension: 50,
			},
		},
		{name: "rechaza secreto corto", env: validEnv(map[string]string{"JWT_SECRET": "corto"}), wantErrIn: "JWT_SECRET"},
		{name: "rechaza credenciales faltantes", env: validEnv(map[string]string{"AUTH_PASSWORD": ""}), wantErrIn: "AUTH_USERNAME"},
		{name: "rechaza timeout no numérico", env: validEnv(map[string]string{"NODE_API_TIMEOUT_MS": "abc"}), wantErrIn: "NODE_API_TIMEOUT_MS"},
		{name: "rechaza dimensión no positiva", env: validEnv(map[string]string{"MATRIX_MAX_DIMENSION": "0"}), wantErrIn: "MATRIX_MAX_DIMENSION"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Load(func(k string) string { return tc.env[k] })
			if tc.wantErrIn != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrIn) {
					t.Fatalf("error = %v, se esperaba que mencione %q", err, tc.wantErrIn)
				}
				return
			}
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Load() =\n%+v\nwant\n%+v", got, tc.want)
			}
		})
	}
}
