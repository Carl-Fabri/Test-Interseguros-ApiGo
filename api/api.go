// Package api embebe el contrato OpenAPI de api-go y la página de documentación (Scalar),
// para que el binario los sirva sin depender de archivos externos.
package api

import _ "embed"

// OpenAPISpec es el contrato OpenAPI 3.1 del servicio (fuente de verdad de la API HTTP).
//
//go:embed openapi.yaml
var OpenAPISpec []byte

// DocsHTML es la página de Scalar que renderiza OpenAPISpec.
//
//go:embed docs.html
var DocsHTML []byte
