/*
 * MIT License
 *
 * Copyright (c) 2023 Beate Ottenwälder
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

package account

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/ottenwbe/recipes-manager/core"
	"github.com/ottenwbe/recipes-manager/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
)

const (
	keycloakAddressCfg      = "keycloak.address"
	keycloakClientSecretCfg = "keycloak.clientSecret"
	keyCloakClientIDCfg     = "keycloak.clientID"
	keyCloakHostCfg         = "keycloak.host"

	cookieTokenName = "token"
)

var (
	keycloakAddress      string
	keyCloakClientSecret string
	keyCloakClientID     string
	keyCloakHost         string

	states map[string]string
)

// AuthKeyCloakAPI for authorization
type AuthKeyCloakAPI struct {
	handler core.Handler
}

var (
	authKeyCloakApi *AuthKeyCloakAPI
)

func init() {
	keycloakAddress = utils.Config.GetString(keycloakAddressCfg)
	keyCloakClientSecret = utils.Config.GetString(keycloakClientSecretCfg)
	keyCloakClientID = utils.Config.GetString(keyCloakClientIDCfg)
	keyCloakHost = utils.Config.GetString(keyCloakHostCfg)

	states = make(map[string]string, 0)
}

// AddAuthAPIsToHandler constructs an API for recipes
func AddAuthAPIsToHandler(handler core.Handler) {
	authKeyCloakApi = &AuthKeyCloakAPI{
		handler,
	}

	authKeyCloakApi.prepareAPI()
}

func (a *AuthKeyCloakAPI) prepareAPI() {

	log.WithField("addr", keycloakAddress).Info("addr")

	if a.handler == nil {
		log.WithField("Component", "Auth Keycloak API").Fatal("No handler defined")
		return
	}

	provider, err := oidc.NewProvider(context.Background(), keycloakAddress)
	if err != nil {
		panic(err)
	}

	keyCloakConfig := &oauth2.Config{
		ClientID:     keyCloakClientID,
		ClientSecret: keyCloakClientSecret,
		RedirectURL:  "http://" + keyCloakHost + "/api/v1/oauth",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}

	v1 := a.handler.API(1)

	//GET the list of accounts
	v1.GET("/auth/keycloak/token", a.getKeyCloak(keyCloakConfig, provider))
	v1.GET("/oauth", a.handleOAUTHResponse(keyCloakConfig, provider))
}

func (a *AuthKeyCloakAPI) handleOAUTHResponse(keyCloakConfig *oauth2.Config, provider *oidc.Provider) func(c *gin.Context) {
	return func(c *gin.Context) {
		log.Info("Token is Back")

		state := c.Query("state")
		sessionState := c.Query("session_state")
		code := c.Query("code")

		log.WithField("state", state).
			WithField("session", sessionState).Info("code")

		if _, ok := states[state]; ok {
			delete(states, state)

			token, err := keyCloakConfig.Exchange(context.Background(), code)
			if err != nil {
				panic(err)
			}

			rawIDToken, ok := token.Extra("id_token").(string)
			if !ok {
				panic("id_token is missing")
			}

			writeToken(c, keyCloakConfig, provider, NewToken(rawIDToken))
			c.JSON(http.StatusOK, &Token{Token: rawIDToken})
		} else {
			log.Debug("State not found")
			c.String(http.StatusNotFound, "")
		}
	}
}

func (a *AuthKeyCloakAPI) getKeyCloak(keyCloakConfig *oauth2.Config, provider *oidc.Provider) func(c *core.APICallContext) {
	return func(c *core.APICallContext) {

		if token, err := getToken(c, provider, keyCloakConfig); err != nil {
			state := createState()
			states[state] = state

			authCodeURL := keyCloakConfig.AuthCodeURL(state)
			log.Infof("Open %s\n", authCodeURL)

			c.Redirect(http.StatusFound, authCodeURL)
		} else {
			log.Info("Token Reused")
			c.JSON(http.StatusOK, token)
		}
	}
}

func createState() string {
	var stateSeed uint64
	err := binary.Read(rand.Reader, binary.LittleEndian, &stateSeed)
	if err != nil {
		log.Error(err.Error())
	}
	state := fmt.Sprintf("%x", stateSeed)
	return state
}
