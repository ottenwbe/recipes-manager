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

package sources

import (
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/ottenwbe/recipes-manager/core"
	"github.com/ottenwbe/recipes-manager/recipes"
	"github.com/ottenwbe/recipes-manager/utils"
)

// SourceResponse describes a sourceClient in detail
type SourceResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
	Version   string `json:"version"`
}

// SourceOAuthConnectResponse informs about the oAuth url
type SourceOAuthConnectResponse struct {
	ID       string `json:"id"`
	OAuthURL string `json:"oAuthURL"`
}

func newSourceResponse(sourceDescription *SourceDescription) *SourceResponse {
	return &SourceResponse{
		ID:        sourceDescription.ID.String(),
		Name:      sourceDescription.Name,
		Connected: sourceDescription.Connected,
		Version:   sourceDescription.Version,
	}
}

// API related sourceDescription data
type API struct {
	sources Sources
	recipes recipes.RecipeDB
}

// NewSourceAPI creates the API for sources
func NewSourceAPI(sources Sources, recipes recipes.RecipeDB) API {
	return API{sources, recipes}
}

// PrepareAPI registers all api endpoints
func (s API) PrepareAPI(router core.Handler, sources Sources, recipes recipes.RecipeDB) {
	s.prepareV1API(router, sources, recipes)
}

func (s API) prepareV1API(router core.Handler, sources Sources, recipes recipes.RecipeDB) {

	v1 := router.API(1)

	// handle oAuth responses
	v1.GET("/sources/:source/oauth", oAuthHandler(sources))

	// start oAuth login process
	v1.GET("/sources/:source/connect", oAuthConnect(sources))

	// lists all sources
	v1.GET("/sources", listSources(sources))

	// sync recipes from sourceClient with local Recipe DB
	v1.PATCH("/sources/:source/recipes", synchronizeSourceRecipes(sources, recipes))
}

// oAuthHandler example
// @Summary Handles Tokens
// @Description Handles Tokens. Typically not directly called.
// @Tags Sources
// @Produce json
// @Param source path string true "Source ID"
// @Success 301 {string} redirect
// @Router /sources/{source}/oauth [get]
func oAuthHandler(sources Sources) func(c *core.APICallContext) {
	return func(c *core.APICallContext) {
		sourceID := c.Param("source")

		query := c.Request.URL.Query()
		state := query["state"][0]
		if state != sourceID {
			c.String(http.StatusNotFound, "Invalid source tried to connect")
			return
		}

		src, err := sourceClient(sourceID, sources)
		if err != nil {
			c.String(http.StatusNotFound, "Invalid Source tried to connect")
			return
		}

		code := query["code"][0]

		err = src.ConnectOAuth(code)
		if err != nil {
			c.String(http.StatusBadRequest, "Cannot connect to Source")
			log.Error(err)
			return
		}

		c.Redirect(http.StatusMovedPermanently, host)
	}
}

// oAuthHandler example
// @Summary Trigger the oauth process
// @Description Trigger the oauth process
// @Tags Sources
// @Produce json
// @Param source path string true "Source ID"
// @Success 200 {object} SourceOAuthConnectResponse
// @Router /sources/{source}/connect [get]
func oAuthConnect(sources Sources) func(c *core.APICallContext) {
	return func(c *core.APICallContext) {
		sourceID := c.Param("source")
		log.Infof("Exchange Token with Source %v", sourceID)

		query := c.Request.URL.Query()
		host = extractSourceRedirectOrDefault(query)

		src, err := sourceClient(sourceID, sources)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		config, err := src.OAuthLoginConfig()
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		oAuthResponse := SourceOAuthConnectResponse{
			ID:       sourceID,
			OAuthURL: config.AuthCodeURL(sourceID, oauth2.AccessTypeOffline),
		}

		c.JSON(http.StatusOK, oAuthResponse)
	}
}

// listSources example
// @Summary List sources
// @Description List sources
// @Tags Sources
// @Produce json
// @Success 200 {object} map[string]SourceResponse
// @Router /sources [get]
func listSources(sources Sources) func(c *core.APICallContext) {
	return func(c *core.APICallContext) {
		sources, err := sources.List()
		if err != nil {
			c.String(http.StatusBadRequest, "Sources could not be listed")
			return
		}
		result := map[string]*SourceResponse{}
		for srcID, source := range sources {
			result[srcID.String()] = newSourceResponse(source)
		}

		c.JSON(http.StatusOK, result)
	}
}

// synchronizeSourceRecipes example
// @Summary Download Recipes from a Source
// @Description Download recipes from a source
// @Tags Sources
// @Produce json
// @Param source path string true "Source ID"
// @Success 200
// @Router /sources/{source}/recipes [get]
func synchronizeSourceRecipes(sources Sources, recipes recipes.RecipeDB) func(c *core.APICallContext) {
	return func(c *core.APICallContext) {
		sourceID := c.Param("source")
		log.WithField("sourceID", sourceID).Debugf("Patch source %v", sourceID)

		src, err := sourceClient(sourceID, sources)
		if err != nil {
			c.String(http.StatusBadRequest, "Source could not be found")
			return
		}

		for _, recipe := range src.Recipes().List() {
			log.WithField("sourceID", sourceID).Infof("Inserted New Recipe: %v", recipe.String())
			err = recipes.Insert(recipe)
			if err != nil {
				log.WithError(err).Error("Could not synchronize a recipe to the db")
			}
			for _, pic := range src.Recipes().Pictures(recipe.ID) {
				log.Infof("Inserted New Recipe Picture: %v", pic.Name)
				err = recipes.AddPicture(pic)
				if err != nil {
					log.WithError(err).Error("Could not synchronize a picture to the db")
				}
			}
		}

		c.String(http.StatusOK, "")
	}
}

func sourceClient(sourceID string, sources Sources) (SourceClient, error) {
	sid, err := SourceIDFromString(sourceID)
	if err != nil {
		return nil, err
	}
	src, err := sources.Client(sid)
	if err != nil {
		return nil, err
	}
	return src, err
}

func sourceDescription(sourceID string, c *core.APICallContext, sources Sources) (*SourceDescription, error) {
	sid, err := SourceIDFromString(sourceID)
	if err != nil {
		return nil, err
	}
	src, err := sources.Description(sid)
	if err != nil {
		return nil, err
	}
	return src, err
}

func extractSourceRedirectOrDefault(query url.Values) string {
	if len(query[REDIRECT]) > 0 {
		log.Debugf("Got Redirect to %v", query[REDIRECT][0])
		return query[REDIRECT][0]
	}
	return host
}

const (
	//SOURCEREDIRECT represents the host address configuration name
	SOURCEREDIRECT = "source.redirect"
	//REDIRECT represents a query parameter that can be set to change source.redirect
	REDIRECT = "redirect"
)

var (
	host string
)

func init() {
	utils.Config.SetDefault(SOURCEREDIRECT, "http://localhost:8080/#!/src")
	host = utils.Config.GetString(SOURCEREDIRECT)
}
