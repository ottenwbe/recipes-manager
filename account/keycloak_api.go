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
}

var (
	authKeyCloakApi *AuthKeyCloakAPI
)

func init() {

	utils.Config.SetDefault(keycloakEnabledCfg, true)

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
		panic(err)
	}

	v1 := a.handler.API(1)

	//GET the list of accounts
	v1.GET("/auth/keycloak/token", a.getKeyCloakToken(keyCloakConfig, provider))
	v1.GET("/oauth", a.handleOAUTHResponse(keyCloakConfig, provider))
}

func (a *AuthKeyCloakAPI) prepareConfig() (*oidc.Provider, *oauth2.Config, error) {
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

func (a *AuthKeyCloakAPI) handleOAUTHResponse(keyCloakConfig *oauth2.Config, provider *oidc.Provider) func(c *gin.Context) {
	return func(c *gin.Context) {
		log.Info("Return Auth response")

		state := c.Query("state")
		sessionState := c.Query("session_state")
		code := c.Query("code")

		log.WithField("state", state).
			WithField("session", sessionState).
			Info("code")

		if currentState := States.FindAndDelete(state); currentState != nil {

			token, err := keyCloakConfig.Exchange(context.Background(), code)
			if err != nil {
				log.Fatal(err)
			}

			s, _ := json.Marshal(token)
			log.Info(string(s))

			rawIDToken, ok := token.Extra("id_token").(string)
			if !ok {
				log.Fatal("id_token is missing")
			}

			storedToken := NewToken(rawIDToken)

			// TODO let frontend store the token
			WriteTokenToCookie(c, keyCloakConfig, provider, storedToken)

			if currentState.Signup {
				idToken, err := ValidateToken(provider, keyCloakConfig, storedToken)
				if err != nil {
					log.Error(err)
				}

				idTokenClaim := IDTokenClaim{}
				if err := idToken.Claims(&idTokenClaim); err != nil {
					panic(err)
				}

				_, err = a.db.NewAccount(idTokenClaim.Email)
				if err != nil {
					log.Error(err)
				}
			}

			// TODO: redirect to UI
			c.Redirect(http.StatusFound, "/login")
		} else {
			log.Debug("State not found")
			c.String(http.StatusUnauthorized, "Could Not Authenticate")
		}
	}
}

func (a *AuthKeyCloakAPI) getKeyCloakToken(keyCloakConfig *oauth2.Config, provider *oidc.Provider) func(c *core.APICallContext) {
	return func(c *core.APICallContext) {

		signup := c.Query("signup")
		url := c.Query("returnTo")

		if token, err := GetTokenFromCookie(c, provider, keyCloakConfig); err != nil {
			state := States.CreateState(url, signup == "true")

			authCodeURL := keyCloakConfig.AuthCodeURL(state.StateString)
			log.Infof("Open %s\n", authCodeURL)

			c.Redirect(http.StatusFound, authCodeURL)
		} else {
			log.Debug("Token reused")
			c.JSON(http.StatusOK, token)
		}
	}
}
