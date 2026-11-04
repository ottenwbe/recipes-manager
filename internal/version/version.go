// SPDX-License-Identifier: MIT

package version

import (
	"sync"
)

var (
	versionGuard     sync.Once
	appVersionString string
	apiVersion       = "v1"
	appVersion       *Version
)

// Version of the application and the exposed api
type Version struct {
	// APP is the version of the current app
	App string `json:"app"`
	// API is the MAJOR API Version supported by the app
	API string `json:"api"`
}

// AppVersion returns the current major version of the API as well as the applications version
func AppVersion() *Version {
	versionGuard.Do(func() {
		appVersion = &Version{
			App: appVersionString,
			API: apiVersion,
		}
	})
	return appVersion
}
