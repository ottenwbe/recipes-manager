/*
 * MIT License - see LICENSE file for details
 */

package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/ottenwbe/recipes-manager/config"
	"github.com/ottenwbe/recipes-manager/core"
	"github.com/ottenwbe/recipes-manager/recipes"
	"github.com/ottenwbe/recipes-manager/sources"
)

func init() {
	config.Config.SetDefault("html.address", ":8080")
	config.Config.SetDefault("html.cors.origin", "*")
	config.Config.SetDefault("recipeDB.host", "mongodb://127.0.0.1:27017")

	log.Infof("Initializing cooking application version=%v API=%v", core.AppVersion().App, core.AppVersion().API)
}

// @title Swagger API documentation for recipes-manager
// @version 1.0
// @description This is the API documentation for recipes-manager.

// @license.name MIT
// @license.url https://github.com/ottenwbe/recipes-manager/blob/master/LICENSE

// @BasePath /api/v1
func main() {

	// configure the cooking app
	coreDB := newCoreDatabase()
	defer closeDatabase(coreDB)

	recipesDB := newRecipeDatabase(coreDB)

	srcRepository := newSources()

	server := newServer(recipesDB, srcRepository)
	defer closeServer(server)

	// start the application
	waitForStop := server.Run()
	waitForStop.Wait()
	log.Info("Stopping Application")
}

func newCoreDatabase() core.DB {
	addr := config.Config.GetString("recipeDB.host")
	db, err := core.NewDatabaseClient(addr)
	failOnError(err)
	return db
}

func newRecipeDatabase(db core.DB) recipes.RecipeDB {
	recipesDB, err := recipes.NewRecipeDB(db)
	failOnError(err)
	return recipesDB
}

func closeDatabase(db core.DB) {
	if db != nil {
		err := db.Close()
		logOnError(err, "Could not close database")
	}
}

func closeServer(server *core.Server) {
	if server != nil {
		err := server.Close()
		logOnError(err, "Error closing server")
	}
}

func newServer(recipesDB recipes.RecipeDB, srcRepository sources.Sources) *core.Server {
	corsOrigin := config.Config.GetString("html.cors.origin")
	handler := core.NewHandler(corsOrigin)

	address := config.Config.GetString("html.address")
	server := core.NewServerWithAddress(address, handler)

	addAPIsToServer(handler, recipesDB, srcRepository)

	return server
}

func addAPIsToServer(handler core.Handler, recipesDB recipes.RecipeDB, srcRepository sources.Sources) {
	err := recipes.AddRecipesAPIToHandler(handler, recipesDB)
	failOnError(err)
	err = sources.AddSourcesAPIToHandler(handler, srcRepository, recipesDB)
	failOnError(err)
	core.AddCoreAPIToHandler(handler)
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
