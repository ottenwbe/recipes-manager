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

import "github.com/google/uuid"

type Type int64

const (
	// KEYCLOAK type
	KEYCLOAK Type = 0
)

// Account document that can be stored in the database
type Account struct {
	Name string `json:"name"`
	ID   AccID  `json:"id"`
	Type Type   `json:"type"`
}

// AccID identifies the account uniquely
type AccID uuid.UUID

// NewAccount is created with a specific name (eMail ID) and type (e.g., KEYCLOAK)
func NewAccount(name string, accountType Type) *Account {
	return &Account{
		Name: name,
		ID:   AccID(uuid.New()),
		Type: accountType,
	}
}
