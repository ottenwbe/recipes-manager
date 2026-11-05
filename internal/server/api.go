// SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/ottenwbe/recipes-manager/internal/version"
)

// AddServerAPIToHandler constructs the base API for the server, i.e., version and health
func AddServerAPIToHandler(handler Handler, healthCheck func() error) {
	v1 := handler.API(1)
	v1.GET("/version", prepareVersionRoutes)
	v1.GET("/health", func(c *APICallContext) { prepareHealthRoute(c, healthCheck) })
}

// Version example
// @Summary Get the current version
// @Description Get the current version
// @Produce  json
// @Success 200 {object} version.Version
// @Router /version [get]
func prepareVersionRoutes(c *APICallContext) {
	c.JSON(http.StatusOK, version.AppVersion())
}

// Health example
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Produce json
// @Success 200 {object} map[string]string
// @Success 503 {object} map[string]string
// @Router /health [get]
func prepareHealthRoute(c *APICallContext, healthCheck func() error) {
	if err := healthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, H{"status": "unhealthy", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, H{"status": "ok"})
}
