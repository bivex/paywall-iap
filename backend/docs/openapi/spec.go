package openapi

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var specYAML []byte

// Bytes returns a copy of the embedded OpenAPI YAML document.
func Bytes() []byte {
	return append([]byte(nil), specYAML...)
}

// ServeYAML serves the embedded OpenAPI schema.
func ServeYAML(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", specYAML)
}