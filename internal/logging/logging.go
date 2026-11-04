// SPDX-License-Identifier: MIT

package logging

import (
	"github.com/ottenwbe/recipes-manager/internal/config"
	log "github.com/sirupsen/logrus"
)

const (
	logLevelCFG = "log.level"
)

func init() {
	config.Config.SetDefault(logLevelCFG, "info")

	logLevelStr := config.Config.GetString(logLevelCFG)

	level, err := log.ParseLevel(logLevelStr)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)

}
