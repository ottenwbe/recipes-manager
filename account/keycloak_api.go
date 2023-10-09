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
	"encoding/json"
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
	keycloakEnabledCfg      = "keycloak.enabled"
)

var (
	keycloakEnabled      bool
	keycloakAddress      string
	keyCloakClientSecret string
	keyCloakClientID     string
	keyCloakHost         string
)

// AuthKeyCloakAPI for authorization
type AuthKeyCloakAPI struct {
	handler core.Handler
	db      *MongoAccountService
	states  *StateService
}

var (
	authKeyCloakApi *AuthKeyCloakAPI
)

func init() {

	utils.Config.SetDefault(keycloakEnabledCfg, false)

	keycloakEnabled = utils.Config.GetBool(keycloakEnabledCfg)
	if keycloakEnabled {
		keycloakAddress = utils.Config.GetString(keycloakAddressCfg)
		keyCloakClientSecret = utils.Config.GetString(keycloakClientSecretCfg)
		keyCloakClientID = utils.Config.GetString(keyCloakClientIDCfg)
		keyCloakHost = utils.Config.GetString(keyCloakHostCfg)
	}
}

// AddAuthAPIsToHandler constructs an API for recipes
func AddAuthAPIsToHandler(handler core.Handler, db core.DB) {

	if keycloakEnabled {

		if handler == nil {
			log.WithField("Component", "Auth Keycloak API").Fatal("No handler defined")
			return
		}

		authKeyCloakApi = &AuthKeyCloakAPI{
			handler: handler,
			db:      NewMongoAccountService(db),
			states:  NewStateService(db),
		}
		authKeyCloakApi.prepareAPI()
	} else {
		log.WithField("Component", "Auth Keycloak API").Info("Keycloak API disabled")
	}
}

func (a *AuthKeyCloakAPI) prepareAPI() {

	log.WithField("addr", keycloakAddress).Info("Prepare Keycloak API")

	provider, keyCloakConfig, err := a.prepareConfig()
	if err != nil {
		log.WithError(err).Error("Account Management Configuration Error")
		panic(err)
	}

	v1 := a.handler.API(1)

	v1.GET("/auth/keycloak/token", a.getKeyCloakToken(keyCloakConfig, provider))
	v1.GET("/auth/keycloak/login", a.getKeyCloakLogin(keyCloakConfig))
	v1.GET("/auth/keycloak/logout", a.handleLogout())
	v1.GET("/oauth", a.handleOAUTHResponse(keyCloakConfig, provider))
}

func (*AuthKeyCloakAPI) prepareConfig() (*oidc.Provider, *oauth2.Config, error) {
	provider, err := oidc.NewProvider(context.Background(), keycloakAddress)

	keyCloakConfig := &oauth2.Config{
		ClientID:     keyCloakClientID,
		ClientSecret: keyCloakClientSecret,
		RedirectURL:  keyCloakHost + "/api/v1/oauth",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}
	return provider, keyCloakConfig, err
}

// handleOAUTHResponse documentation
// @Summary OAuth endpoint
// @Description OAuth endpoint
// @Tags Auth
// @Produce json
// @Success 200 {integer} number
// @Router /oauth [get]
func (a *AuthKeyCloakAPI) handleOAUTHResponse(keyCloakConfig *oauth2.Config, provider *oidc.Provider) func(c *gin.Context) {
	return func(c *gin.Context) {
		log.Info("Return Auth response")

		state := c.Query("state")
		sessionState := c.Query("session_state")
		code := c.Query("code")

		log.WithField("state", state).
			WithField("session", sessionState).
			Info("code")

		if currentState := a.states.FindAndDelete(state); currentState != nil {

			token := getTokenFromKeyCloak(keyCloakConfig, code)

			idTokenClaim, err := GetClaims(provider, keyCloakConfig, token)
			if err != nil {
				log.Error("Signup error", err)
				c.Redirect(http.StatusUnauthorized, "/401")
				return
			}

			a.tryStoreAccountIfSignup(idTokenClaim, currentState)

			if _, err := a.db.FindAccount(idTokenClaim.Email, KEYCLOAK); err == nil {
				WriteTokenToCookie(c, keyCloakConfig, provider, token)
				c.Redirect(http.StatusFound, "/")
			} else {
				c.Redirect(http.StatusNotFound, "/404")
			}
		} else {
			log.Debug("State not found")
			c.Redirect(http.StatusNotFound, "/404")
		}
	}
}

func (a *AuthKeyCloakAPI) tryStoreAccountIfSignup(idTokenClaim *IDTokenClaim, currentState *State) {

	log.WithField("method", "tryStoreAccountIfSignup").Infof("Signup request, %v", currentState.Signup)

	if currentState.Signup {
		_, err := a.db.NewAccount(idTokenClaim.Email, KEYCLOAK)
		if err != nil {
			log.Error("Account was already saved", err)
		}
	}
}

func getTokenFromKeyCloak(keyCloakConfig *oauth2.Config, code string) *Token {
	token, err := keyCloakConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Error(err)
	}

	s, err := json.Marshal(token)
	if err != nil {
		log.Error("Could Not Unmarshal Token", err)
	}
	log.Debugf("Token received %s", s)

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Error("id_token is missing")
	}

	storedToken := NewToken(rawIDToken)
	return storedToken
}

// getKeyCloakToken documentation
// @Summary Get the token
// @Description Get the token for a user
// @Tags Auth
// @Produce json
// @Success 200 {integer} number
// @Router /auth/keycloak/token [get]
func (a *AuthKeyCloakAPI) getKeyCloakToken(keyCloakConfig *oauth2.Config, provider *oidc.Provider) func(c *core.APICallContext) {
	return func(c *core.APICallContext) {

		if token, err := GetTokenFromCookie(c, provider, keyCloakConfig); err != nil {
			c.String(http.StatusNotFound, "You have to login first")
		} else {
			idTokenClaim, err := GetClaims(provider, keyCloakConfig, token)
			if err != nil {
				log.Error("Token Claims Not Found", err)
				c.String(http.StatusNotFound, "try and login or signup again")
			} else {
				if _, err := a.db.FindAccount(idTokenClaim.Email, KEYCLOAK); err == nil {
					log.Debug("Token Reused")
					c.JSON(http.StatusOK, token)
				} else {
					log.Error("Token Could Not Be Reused", err)
					c.String(http.StatusNotFound, "you have to signup first")
				}
			}
		}
	}
}

// getKeyCloakToken documentation
// @Summary Login by creating a token
// @Description Login by creating a token
// @Tags Auth
// @Produce json
// @Success 200 {integer} number
// @Router /auth/keycloak/login [get]
func (a *AuthKeyCloakAPI) getKeyCloakLogin(keyCloakConfig *oauth2.Config) func(c *core.APICallContext) {
	return func(c *core.APICallContext) {

		signup := c.Query("signup")
		url := c.Query("returnTo")

		state := a.states.CreateState(url, signup == "true")

		authCodeURL := keyCloakConfig.AuthCodeURL(state.State)
		log.Debugf("Open %s\n", authCodeURL)

		c.Redirect(http.StatusFound, authCodeURL)
	}
}

// handleLogout documentation
// @Summary Logout by deleting the token
// @Description Logout by deleting the token.
// @Tags Auth
// @Produce json
// @Success 200 {integer} number
// @Router /auth/keycloak/logout [get]
func (*AuthKeyCloakAPI) handleLogout() func(c *core.APICallContext) {
	return func(c *core.APICallContext) {
		DeleteTokenCookie(c)
		c.JSON(http.StatusOK, nil)
	}
}
