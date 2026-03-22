/*
 * MIT License - see LICENSE file for details
 */

package config

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// RecipeConfig allows the recipe application to retrieve configuration data
type RecipeConfig interface {
	GetInt64(key string) int64
	GetString(key string) string
	GetBool(key string) bool
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

type viperConfig struct {
	v *viper.Viper
}

// NewViperConfig creates a configuration based on the viper framework
func NewViperConfig(name string, paths []string) RecipeConfig {
	c := &viperConfig{
		v: viper.New(),
	}
	c.config(name, paths)
	return c
}

// GetBool returns a boolean for the given key
func (c *viperConfig) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetString returns a string for the given key
func (c *viperConfig) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt64 returns an integer for the given key
func (c *viperConfig) GetInt64(key string) int64 {
	return c.v.GetInt64(key)
}

// SetDefault sets the default value for a key
func (c *viperConfig) SetDefault(key string, val interface{}) {
	c.v.SetDefault(key, val)
}

// BindEnv binds a key to an environment variable
func (c *viperConfig) BindEnv(key string) {
	c.v.BindEnv(key)
}

func (c *viperConfig) configEnv() {
	c.v.SetEnvPrefix("go_cook")
	c.v.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	c.v.SetEnvKeyReplacer(replacer)
}

func (c *viperConfig) config(name string, paths []string) {
	c.configEnv()
	c.v.SetConfigName(name) // name of config file (without extension)
	for i := range paths {
		c.v.AddConfigPath(paths[i])
	}

	err := c.v.ReadInConfig() // Find and read the config file
	if err != nil {           // Handle errors reading the config file
		log.WithError(err).Error("Error while reading config file ")
	}
}

// Debug prints all configurations
func (c *viperConfig) Debug() {
	c.v.Debug()
}
