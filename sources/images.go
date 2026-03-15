/*
 * MIT License - see LICENSE file for details
 */

package sources

import (
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const metaData = "data:image/jpeg;base64,"

// DownloadIMGAsBase64 will download an image from an url. It returns a base64 encoded image.
func DownloadIMGAsBase64(url string) (base64img string, err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	response, err := client.Get(url)
	if err != nil {
		// log.WithError(err).WithField("url", url).Error("Could not download image...")
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
func IMGFileToBase64(path string) (string, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		log.WithError(err).Error("Could not open image")
		return "", err
	}

	// convert the buffered bytes to a base64 string
	imgBase64Str := base64.StdEncoding.EncodeToString(buf)

	// add meta data
	base64img := metaData + imgBase64Str

	return base64img, nil
}
