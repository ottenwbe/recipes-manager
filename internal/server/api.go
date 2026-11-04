// SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/ottenwbe/recipes-manager/internal/version"
)

// AddServerAPIToHandler constructs the base API for the server, i.e., version and health
func AddServerAPIToHandler(handler Handler) {
	v1 := handler.API(1)
	v1.GET("/version", prepareVersionRoutes)
	v1.GET("/health", prepareHealthRoute)
}

// Version example
// @Summary Get the curent version
// @Description get the current version
// @Produce  json
// @Success 200 {object} version.Version
// @Router /version [get]
func prepareVersionRoutes(c *APICallContext) {
	c.JSON(http.StatusOK, version.AppVersion())
}

// Health example
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Produce  json
// @Success 200 {object} map[string]string
// @Router /health [get]
func prepareHealthRoute(c *APICallContext) {
	c.JSON(http.StatusOK, H{"status": "ok"})
}
