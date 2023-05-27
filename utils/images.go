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

package utils

import (
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

const metaData = "data:image/jpeg;base64,"

// DownloadIMGAsBase64 will download an image from an url. It returns a base64 encoded image.
func DownloadIMGAsBase64(url string) (base64img string, err error) {

	response, err := http.Get(url)
	if err != nil {
		//log.WithError(err).WithField("url", url).Error("Could not download image...")
		return "", err
	}
	defer func() { _ = response.Body.Close() }()

	buf, err := io.ReadAll(response.Body)
	if err != nil {
		log.WithError(err).WithField("url", url).Error("Could not read response while downloading an image...")
	}

	// convert the buffered bytes to a base64 string
	imgBase64Str := base64.StdEncoding.EncodeToString(buf)

	// add meta data
	base64img = metaData + imgBase64Str

	return base64img, err
}

// IMGFileToBase64 reads an image from a file at given path, i.e., /home/user/test.jpeg. This image is returned as base64 encoded string.
func IMGFileToBase64(path string) string {
	buf, err := os.ReadFile(path)
	if err != nil {
		log.WithError(err).Error("Could not open image")
		return ""
	}

	// convert the buffered bytes to a base64 string
	imgBase64Str := base64.StdEncoding.EncodeToString(buf)

	// add meta data
	base64img := metaData + imgBase64Str

	return base64img
}
