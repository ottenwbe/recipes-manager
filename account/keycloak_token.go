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
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// TokenMeta information about the token provider and configuration
type TokenMeta struct {
	Provider *oidc.Provider
	Config   *oauth2.Config
}

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

// ValidateToken ensures the validity of the token
func (t *TokenMeta) ValidateToken(token *Token) (*oidc.IDToken, error) {
	verifier := t.Provider.Verifier(&oidc.Config{ClientID: t.Config.ClientID})
	idToken, err := verifier.Verify(context.Background(), token.Token)
	return idToken, err
}

// GetClaims from token
func (t *TokenMeta) GetClaims(token *Token) (*IDTokenClaim, error) {
	idToken, err := t.ValidateToken(token)
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
