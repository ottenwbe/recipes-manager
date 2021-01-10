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

// AddCoreAPIToHandler constructs an API for recipes
func AddCoreAPIToHandler(handler Handler) {
	v1 := handler.API(1)
	v1.GET("/version", prepareVersionRoutes)
	v1.GET("/health", prepareHealthRoutes)
}

// Version API
// @Summary Get the current version
// @Description get the current version
// @Produce  json
// @Success 200 {object} Version
// @Router /health [get]
func prepareVersionRoutes(c *APICallContext) {
	c.JSON(200, AppVersion())
}

// Health Endpoint API
// @Summary Get the current version
// @Description get the current version
// @Produce  json
// @Success 200 {object} Version
// @Router /version [get]
func prepareHealthRoutes(c *APICallContext) {
	c.String(200, "{}")
}
