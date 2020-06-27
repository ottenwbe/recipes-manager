/*
 * MIT License
 *
 * Copyright (c) 2020 Beate Ottenwälder
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package core

import (
	"fmt"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/ottenwbe/go-life/utils"
)

const (
	addressCfg    = "html.address"
	corsOriginCfg = "html.cors.origin"

	baseAPIPath = "api"
)

var (
	defaultAddress string
	corsOrigin     string
)

// init configures the router for api calls when the core package is initialized
func init() {
	utils.Config.SetDefault(addressCfg, ":8080")
	utils.Config.SetDefault(corsOriginCfg, "*")
	defaultAddress = utils.Config.GetString(addressCfg)
	corsOrigin = utils.Config.GetString(corsOriginCfg)
}

//Routes is managing a set of API endpoints.
//Routes implementation(s) call handler function to perform typical CRUD operations (GET, POST, PATCH, ...).
type Routes interface {
	//Route is created to a specific set of endpoints
	Route(string) Routes
	//GET endpoint is added to the routes set and registers a corresponding handler
	GET(string, func(c *APICallContext))
	//Path returns the base path
	Path() string
	//PATCH endpoint is added to the routes set and registers a corresponding handler
	PATCH(string, func(c *APICallContext))
	//POST endpoint is added to the routes set and registers a corresponding handler
	POST(string, func(c *APICallContext))
}

//Router is a facade for a HTTP router and can be implemented by a concrete router like gin.
type Router interface {
	Run() error
	API(version int16) Routes
}

//APICallContext is a facade for any concrete Context, e.g. gins
type APICallContext = gin.Context

//NewRouterA creates a Router for API Calls with a given ADDRESS
func NewRouterA(addr string) Router {
	router := &ginRouter{
		gin.New(),
		addr,
		make(map[string]Routes),
	}
	router.configure()
	router.prepareDefaultRoutes()
	return router
}

//NewRouter creates a router for API calls with a pre-configured ADDRESS
func NewRouter() Router {
	return NewRouterA(defaultAddress)
}

type ginRouter struct {
	router       *gin.Engine
	address      string
	routerGroups map[string]Routes
}

func (g *ginRouter) addSubGroup(groupName string, subGroupName string) Routes {
	rg, ok := g.routerGroups[groupName]
	if !ok {
		// we create the missing group if it cannot be found
		rg = g.route(groupName)
		g.routerGroups[groupName] = rg
	}
	return rg.Route(subGroupName)
}

func v(version int16) string {
	return fmt.Sprintf("v%v", version)
}

//API registers the endpoint /api/v<version> and returns a group of endpoints under /api/v<version>
func (g *ginRouter) API(version int16) Routes {
	rg, ok := g.routerGroups[v(version)]
	if !ok {
		rg = g.addSubGroup(baseAPIPath, v(version))
		g.routerGroups[v(version)] = rg
	}
	return rg
}

//Run the server for the api
func (g *ginRouter) Run() error {
	return g.router.Run(g.address)
}

func (g *ginRouter) route(route string) Routes {
	return &ginRoutes{g.router.Group(route)}
}

// configure the default middleware with a logger and recovery (crash-free) middleware
func (g *ginRouter) configure() {
	g.router.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	g.router.Use(g.corsMiddleware())
	// Return 500 if there was a panic.
	g.router.Use(gin.Recovery())
}

func (g *ginRouter) prepareDefaultRoutes() {
	g.router.GET("/version", func(c *gin.Context) {
		c.JSON(200, AppVersion())
	})
}

type ginRoutes struct {
	rg *gin.RouterGroup
}

func (g *ginRoutes) Route(path string) Routes {
	return &ginRoutes{g.rg.Group(path)}
}

//GET endpoint for a specific path and a corresponding handler
func (g *ginRoutes) GET(path string, handler func(c *APICallContext)) {
	g.rg.GET(path, handler)
}

//PATCH endpoint for a specific path and a corresponding handler
func (g *ginRoutes) PATCH(path string, handler func(c *APICallContext)) {
	g.rg.PATCH(path, handler)
}

//POST endpoint for a specific path and a corresponding handler
func (g *ginRoutes) POST(path string, handler func(c *APICallContext)) {
	g.rg.POST(path, handler)
}

//PATH of the given route
func (g *ginRoutes) Path() string {
	return g.rg.BasePath()
}

func (g *ginRouter) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, POST")

		if c.Request.Method == "OPTIONS" || c.Request.Method == "PUT" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
