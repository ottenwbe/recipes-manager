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
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	// based on swagger documentation
	_ "github.com/ottenwbe/go-cook/docs"

	"github.com/ottenwbe/go-cook/utils"
)

const (
	addressCfg         = "html.address"
	corsAllowOriginCfg = "html.cors.origin"

	baseAPIPath = "api"
)

var (
	defaultAddress string
	corsOrigin     string
)

// init configures the handler for api calls when the core package is initialized
func init() {
	utils.Config.SetDefault(addressCfg, ":8080")
	utils.Config.SetDefault(corsAllowOriginCfg, "*")
	defaultAddress = utils.Config.GetString(addressCfg)
	corsOrigin = utils.Config.GetString(corsAllowOriginCfg)
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
	//PUT endpoint is added to the routes set and registers a corresponding handler
	PUT(string, func(c *APICallContext))
	//DELETE endpoint is added to the routes set and registers a corresponding handler
	DELETE(string, func(c *APICallContext))
}

//Handler is a facade for a HTTP handler and can be implemented by a concrete handler like gin.
type Handler interface {
	API(version int16) Routes
	http.Handler
}

//APICallContext is a facade for any concrete Context, e.g. gins
type APICallContext = gin.Context

//NewHandler creates a handler for API calls with a pre-configured ADDRESS
func NewHandler() Handler {
	handler := &ginHandler{
		gin.New(),
		make(map[string]Routes),
	}
	handler.configure()
	return handler
}

// @title GO-Cook API
// @version 1.0
// @description This is the go-cook api

// @license.name MIT
// @BasePath /api/v1

type ginHandler struct {
	handler      *gin.Engine
	routerGroups map[string]Routes
}

func (g *ginHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	g.handler.ServeHTTP(writer, request)
}

func (g *ginHandler) addSubGroup(groupName string, subGroupName string) Routes {
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
func (g *ginHandler) API(version int16) Routes {
	rg, ok := g.routerGroups[v(version)]
	if !ok {
		rg = g.addSubGroup(baseAPIPath, v(version))
		g.routerGroups[v(version)] = rg
	}
	return rg
}

func (g *ginHandler) route(route string) Routes {
	return &ginRoutes{g.handler.Group(route)}
}

// configure the default middleware with a logger and recovery (crash-free) middleware
func (g *ginHandler) configure() {

	url := ginSwagger.URL("doc.json") // The url pointing to API definition
	g.handler.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	g.handler.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	g.handler.Use(g.corsMiddleware())
	// Return 500 if there was a panic.
	g.handler.Use(gin.Recovery())
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

//PUT endpoint for a specific path and a corresponding handler
func (g *ginRoutes) PUT(path string, handler func(c *APICallContext)) {
	g.rg.PUT(path, handler)
}

//DELETE endpoint for a specific path and a corresponding handler
func (g *ginRoutes) DELETE(path string, handler func(c *APICallContext)) {
	g.rg.DELETE(path, handler)
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

func (g *ginHandler) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, POST, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Server interface which extends the http.Server
type Server struct {
	Address       string
	server        *http.Server
	stopWaitGroup *sync.WaitGroup
}

//NewServerA creates a new server using a given address to listen to
func NewServerA(addr string, handler http.Handler) Server {
	return Server{
		Address: addr,
		server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		stopWaitGroup: &sync.WaitGroup{}}
}

//NewServerH creates a new server using the default address with a custom handler
func NewServerH(handler http.Handler) Server {
	return NewServerA(defaultAddress, handler)
}

//NewServer creates a new server to listen on the defaultAddress
func NewServer() Server {
	return NewServerA(defaultAddress, NewHandler())
}

//Run the server for the API
func (s Server) Run() *sync.WaitGroup {
	s.stopWaitGroup.Add(1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Errorf("Server's not running: %s\n", err)
		}
		s.stopWaitGroup.Done()
	}()
	return s.stopWaitGroup
}

//Close the server
func (s Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	return err
}
