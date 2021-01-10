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
var Config = newRecipeConfig()

//defaultConfigName is the name of the default configuration
const defaultConfigName = "recipes-manager-config"

//envPrefix is the prefix for all environment variables
const defaultEnvPrefix = "go_cook"

var defaultPaths = []string{"/etc/recipes-manager/", "$HOME/.recipes-manager", "."}

func newRecipeConfig() RecipeConfig {
	recipeConfig := &viperConfig{
		envPrefix:      defaultEnvPrefix,
		configFile:     defaultConfigName,
		configFilePath: defaultPaths,
	}
	recipeConfig.initConfigEnv()
	recipeConfig.initConfigFile()
	recipeConfig.readConfig()
	return recipeConfig
}

type viperConfig struct {
	envPrefix      string
	configFile     string
	configFilePath []string
}

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

func (v *viperConfig) initConfigFile() {
	viper.SetConfigName(v.configFile) // name of config file (without extension)
	for i := range v.configFilePath {
		viper.AddConfigPath(v.configFilePath[i])
	}
}

func (v *viperConfig) initConfigEnv() {
	viper.SetEnvPrefix(v.envPrefix)
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
