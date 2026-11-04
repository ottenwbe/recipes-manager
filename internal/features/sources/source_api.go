// SPDX-License-Identifier: MIT

package sources

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/ottenwbe/recipes-manager/internal/config"
	"github.com/ottenwbe/recipes-manager/internal/features/recipes"
	"github.com/ottenwbe/recipes-manager/internal/server"
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

// AddSourcesAPIToHandler creates the API for sources
func AddSourcesAPIToHandler(handler server.Handler, sources Sources, recipesDB recipes.RecipeDB) error {
	sourcesAPI := API{sources, recipesDB}
	sourcesAPI.PrepareAPI(handler)
	return nil
}

// PrepareAPI registers all api endpoints
func (s API) PrepareAPI(router server.Handler) {
	s.prepareV1API(router, s.sources, s.recipes)
}

func (s API) prepareV1API(router server.Handler, sources Sources, recipes recipes.RecipeDB) {

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
func oAuthHandler(sources Sources) func(c *server.APICallContext) {
	return func(c *server.APICallContext) {
		sourceID := c.Param("source")

		state := c.Query("state")
		code := c.Query("code")

		if state == "" || code == "" {
			c.String(http.StatusBadRequest, "Missing state or code parameter")
			return
		}

		decodedState, err := base64.URLEncoding.DecodeString(state)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid state encoding")
			return
		}
		stateParts := strings.Split(string(decodedState), "|")
		if len(stateParts) != 2 {
			c.String(http.StatusBadRequest, "Invalid state format")
			return
		}

		redirectURL := stateParts[1]
		if stateParts[0] != sourceID {
			redirectWithError(c, redirectURL, "source_mismatch")
			return
		}

		src, err := sourceClient(sourceID, sources)
		if err != nil {
			redirectWithError(c, redirectURL, "invalid_source")
			return
		}

		err = src.ConnectOAuth(code)
		if err != nil {
			log.WithError(err).Error("Cannot connect to source")
			redirectWithError(c, redirectURL, "connection_failed")
			return
		}

		c.Redirect(http.StatusFound, stateParts[1])
	}
}

func redirectWithError(c *server.APICallContext, target string, errorMsg string) {
	u, err := url.Parse(target)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid redirect url in state")
		return
	}
	q := u.Query()
	q.Set("error", errorMsg)
	u.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, u.String())
}

// oAuthHandler example
// @Summary Trigger the oauth process
// @Description Trigger the oauth process
// @Tags Sources
// @Produce json
// @Param source path string true "Source ID"
// @Success 200 {object} SourceOAuthConnectResponse
// @Router /sources/{source}/connect [get]
func oAuthConnect(sources Sources) func(c *server.APICallContext) {
	return func(c *server.APICallContext) {
		sourceID := c.Param("source")
		log.Infof("Exchange Token with Source %v", sourceID)

		query := c.Request.URL.Query()
		redirectURL := extractSourceRedirectOrDefault(query)

		src, err := sourceClient(sourceID, sources)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		oauthConfig, err := src.OAuthLoginConfig()
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		state := base64.URLEncoding.EncodeToString([]byte(sourceID + "|" + redirectURL))

		oAuthResponse := SourceOAuthConnectResponse{
			ID:       sourceID,
			OAuthURL: oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")),
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
func listSources(sources Sources) func(c *server.APICallContext) {
	return func(c *server.APICallContext) {
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
// @Param source path string true "Source ID" format string
// @Success 200
// @Router /sources/{source}/recipes [patch]
func synchronizeSourceRecipes(sources Sources, recipes recipes.RecipeDB) func(c *server.APICallContext) {
	return func(c *server.APICallContext) {
		sourceID := c.Param("source")
		log.WithField("sourceID", sourceID).Debugf("Patch source %v", sourceID)

		src, err := sourceClient(sourceID, sources)
		if err != nil {
			c.String(http.StatusBadRequest, "Source could not be found")
			return
		}

		count := 0
		for _, recipe := range src.Recipes().List() {
			log.WithField("sourceID", sourceID).Infof("Inserted New Recipe: %v", recipe.String())
			// Try to insert; if it fails (duplicate ID), update it instead.
			if err := recipes.Insert(recipe); err != nil {
				if updateErr := recipes.Update(recipe.ID, recipe); updateErr != nil {
					log.WithError(updateErr).Error("Could not synchronize (insert/update) a recipe to the db")
				}
			}
			for _, pic := range src.Recipes().Pictures(recipe.ID) {
				log.Infof("Inserted New Recipe Picture: %v", pic.Name)
				err = recipes.AddPicture(pic)
				if err != nil {
					log.WithError(err).Error("Could not synchronize a picture to the db")
				}
			}
			count++
		}

		c.JSON(http.StatusOK, server.H{"synced": count})
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

func sourceDescription(sourceID string, c *server.APICallContext, sources Sources) (*SourceDescription, error) {
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
	return config.Config.GetString(SOURCEREDIRECT)
}

const (
	//SOURCEREDIRECT represents the host address configuration name
	SOURCEREDIRECT = "source.redirect"
	//REDIRECT represents a query parameter that can be set to change source.redirect
	REDIRECT = "redirect"
)

func init() {
	config.Config.SetDefault(SOURCEREDIRECT, "http://localhost:8080/#!/src")
}
