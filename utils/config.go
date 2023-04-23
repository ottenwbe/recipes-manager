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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

// RecipeConfig allows the recipe application to retrieve configuration data
type RecipeConfig interface {
	GetInt64(key string) int64
	GetString(key string) string
	SetDefault(key string, val interface{})
	BindEnv(key string)
	Debug()
}

// Config allows this application to access its configuration
var Config RecipeConfig

// defaultConfigName is the name of the default configuration
const defaultConfigName = "recipes-manager-config"

var defaultPaths = []string{"/etc/recipes-manager/", "$HOME/.recipes-manager", "."}

func init() {
	Config = NewViperConfig(defaultConfigName, defaultPaths)
}

type viperConfig struct{}

// NewViperConfig creates a configuration based on the viper framework
func NewViperConfig(name string, paths []string) RecipeConfig {
	c := &viperConfig{}
	c.config(name, paths)
	return c
}

// GetString returns a string for the given key
func (*viperConfig) GetString(key string) string {
	return viper.GetString(key)
}

// GetInt64 returns an integer for the given key
func (*viperConfig) GetInt64(key string) int64 {
	return viper.GetInt64(key)
}

// SetDefault sets the default value for a key
func (*viperConfig) SetDefault(key string, val interface{}) {
	viper.SetDefault(key, val)
}

// BindEnv binds a key to an environment variable
func (*viperConfig) BindEnv(key string) {
	viper.BindEnv(key)
}

func (*viperConfig) configEnv() {
	viper.SetEnvPrefix("go_cook")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
}

func (v *viperConfig) config(name string, paths []string) {

	v.configEnv()

	viper.SetConfigName(name) // name of config file (without extension)
	for i := range paths {
		viper.AddConfigPath(paths[i])
	}

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.WithError(err).Error("Error while reading config file ")
	}
}

// Debug prints all configurations
func (*viperConfig) Debug() {
	viper.Debug()
}
