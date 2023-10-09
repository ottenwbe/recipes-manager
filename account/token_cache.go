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
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/ottenwbe/recipes-manager/core"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const cookieTokenName = "token"

// Token document that can be stored, i.e., in a cookie
type Token struct {
	Token string `json:"token"`
}

// NewToken document is created
func NewToken(token string) *Token {
	return &Token{
		Token: token,
	}
}

// IDTokenClaim represents the relevant claims
type IDTokenClaim struct {
	Email string `json:"email"`
}

// GetTokenFromCookie returns the token or an error
func GetTokenFromCookie(c *core.APICallContext, provider *oidc.Provider, config *oauth2.Config) (*Token, error) {
	cookie, err := c.Cookie(cookieTokenName)

	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID})
	_, err = verifier.Verify(context.Background(), cookie)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return NewToken(cookie), err
}

// DeleteTokenCookie by setting the cookie maxAge to -1
func DeleteTokenCookie(c *core.APICallContext) {
	c.SetCookie(cookieTokenName, "", -1, "/", keyCloakHost, false, false)
}

// WriteTokenToCookie stores the token in a cookie
func WriteTokenToCookie(c *core.APICallContext, config *oauth2.Config, provider *oidc.Provider, token *Token) {

	idToken, err := ValidateToken(provider, config, token)
	if err != nil {
		log.WithField("Token", "WriteTokenToCookie").Error(err)
	} else {
		c.SetCookie(cookieTokenName, token.Token, 3600, "/", keyCloakHost, false, false)
	}

	log.Debugf("Cookie value: %s \n", idToken)
}

// ValidateToken ensures the validity of the token
func ValidateToken(provider *oidc.Provider, config *oauth2.Config, token *Token) (*oidc.IDToken, error) {
	verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID})
	idToken, err := verifier.Verify(context.Background(), token.Token)
	return idToken, err
}

func GetClaims(provider *oidc.Provider, config *oauth2.Config, token *Token) (*IDTokenClaim, error) {
	idToken, err := ValidateToken(provider, config, token)
	if err != nil {
		log.Error("Could Not Validate ID Token", err)
		return nil, err
	}

	idTokenClaim := &IDTokenClaim{}
	if err := idToken.Claims(idTokenClaim); err != nil {
		log.Error("Could not determine ID Token", err)
		return nil, err
	}

	return idTokenClaim, nil
}
