// SPDX-License-Identifier: MIT

package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/ottenwbe/recipes-manager/account"
	"github.com/ottenwbe/recipes-manager/config"
	"github.com/ottenwbe/recipes-manager/db"
	"github.com/ottenwbe/recipes-manager/recipes"
	"github.com/ottenwbe/recipes-manager/server"
	"github.com/ottenwbe/recipes-manager/sources"
	"github.com/ottenwbe/recipes-manager/version"
)

// @title Swagger API documentation for recipes-manager
// @version 1.0
// @description This is the API documentation for recipes-manager.

// @license.name MIT
// @license.url https://github.com/ottenwbe/recipes-manager/blob/master/LICENSE

// @BasePath /api/v1
func main() {

	log.Infof("Initializing cooking application version=%v API=%v", version.AppVersion().App, version.AppVersion().API)

	// configure the cooking app
	database := newDatabase()
	defer closeDatabase(database)

	recipesDB := newRecipeDatabase(database)

	srcRepository := newSources()

	server := newServer(recipesDB, srcRepository)
	defer closeServer(server)

	// start the application
	waitForStop := server.Run()

	waitForStop.Wait()
	log.Info("Stopping cooking application")
}

func newDatabase() db.DB {
	addr := config.Config.GetString(db.RecipeDBHostCfg)
	database, err := db.NewDatabaseClient(addr)
	failOnError(err)
	return database
}

func newRecipeDatabase(database db.DB) recipes.RecipeDB {
	recipesDB, err := recipes.NewRecipeDB(database)
	failOnError(err)
	return recipesDB
}

func closeDatabase(database db.DB) {
	if database != nil {
		err := database.Close()
		logOnError(err, "Could not close database")
	}
}

func closeServer(server *server.Server) {
	if server != nil {
		err := server.Close()
		logOnError(err, "Error closing server")
	}
}

func newServer(recipesDB recipes.RecipeDB, srcRepository sources.Sources) *server.Server {
	corsOrigin := config.Config.GetString(server.CorsOriginCfg)
	handler := server.NewHandler(corsOrigin)

	address := config.Config.GetString(server.AddressCfg)
	server := server.NewServerWithAddress(address, handler)

	addAPIsToServer(handler, recipesDB, srcRepository)

	return server
}

func addAPIsToServer(handler server.Handler, recipesDB recipes.RecipeDB, srcRepository sources.Sources) {
	err := recipes.AddRecipesAPIToHandler(handler, recipesDB)
	failOnError(err)
	err = sources.AddSourcesAPIToHandler(handler, srcRepository, recipesDB)
	failOnError(err)
	server.AddServerAPIToHandler(handler)
	account.AddAuthAPIsToHandler(handler, recipesDB)
}

func newSources() sources.Sources {
	srcRepository := sources.NewSources()

	if sources.IsDriveEnabled() {
		source := sources.OpenNewGoogleDriveConnection()
		cfg, err := source.OAuthLoginConfig()
		if err != nil {
			log.WithError(err).Warn("Could not create OAuth Config")
		} else {
			err = srcRepository.Add(
				sources.NewSourceDescription(source.ID(), source.Name(), source.Version(), cfg),
				source,
			)
			warnOnError(err, "Could not add source")
		}
	}

	return srcRepository
}

func logOnError(err error, message string) {
	if err != nil {
		log.WithError(err).Error(message)
	}
}

func warnOnError(err error, message string) {
	if err != nil {
		log.WithError(err).Warn(message)
	}
}

func failOnError(err error) {
	if err != nil {
		log.WithError(err).Fatal("Fatal Error")
	}
}
