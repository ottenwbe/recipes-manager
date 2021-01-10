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

//RecipeConfig allows the recipe-manager to retrieve configuration data.
//It acts as facade to the actual configuration.
type RecipeConfig interface {
	GetInt64(key string) int64
	GetString(key string) string
	SetDefault(key string, val interface{})
	Debug()
}

//Config allows this application to access its configuration
var Config RecipeConfig

//defaultConfigName is the name of the default configuration
const defaultConfigName = "recipes-manager-config"

//envPrefix is the prefix for all environment variables
const envPrefix = "go_cook"

var defaultPaths = []string{"/etc/recipes-manager/", "$HOME/.recipes-manager", "."}

func init() {
	recipeConfig := &viperConfig{}
	recipeConfig.initConfigEnv(envPrefix)
	recipeConfig.initConfigFile(defaultConfigName, defaultPaths)
	recipeConfig.readConfig()
	Config = recipeConfig
}

type viperConfig struct{}

//GetString returns a string for the given key
func (*viperConfig) GetString(key string) string {
	return viper.GetString(key)
}

//GetInt64 returns an integer for the given key
func (*viperConfig) GetInt64(key string) int64 {
	return viper.GetInt64(key)
}

//SetDefault sets the default value for a key
func (*viperConfig) SetDefault(key string, val interface{}) {
	viper.SetDefault(key, val)
}

func (*viperConfig) initConfigFile(name string, paths []string) {
	viper.SetConfigName(name) // name of config file (without extension)
	for i := range paths {
		viper.AddConfigPath(paths[i])
	}
}

func (*viperConfig) initConfigEnv(prefix string) {
	viper.SetEnvPrefix(prefix)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
}

func (*viperConfig) readConfig() {
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.WithError(err).Error("Error while reading config file.")
	}
}

//Debug prints all configurations
func (*viperConfig) Debug() {
	viper.Debug()
}
