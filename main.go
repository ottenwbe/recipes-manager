/*
 * MIT License - see LICENSE file for details
 */

package main

import (
	"github.com/ottenwbe/recipes-manager/account"
	log "github.com/sirupsen/logrus"

	"github.com/ottenwbe/recipes-manager/core"
	"github.com/ottenwbe/recipes-manager/recipes"
	"github.com/ottenwbe/recipes-manager/sources"
)

func init() {
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
	recipesDB := newCloseableDatabase()
	defer closeDatabase(recipesDB)

	srcRepository := newSources()

	server := newServer(recipesDB, srcRepository)

	// start the application
	waitForStop := server.Run()
	waitForStop.Wait()
	log.Info("Stopping Application")
}

func newCloseableDatabase() recipes.RecipeDB {
	recipesDB, err := recipes.NewDatabaseClient()
	failOnError(err)
	return recipesDB
}

func closeDatabase(recipesDB recipes.RecipeDB) {
	err := recipesDB.Close()
	logOnError(err, "Could not close database ...")
}

func newServer(recipesDB recipes.RecipeDB, srcRepository sources.Sources) core.Server {
	handler := core.NewHandler()
	server := core.NewServerWithHandler(handler)

	addAPIsToServer(handler, recipesDB, srcRepository)

	return server
}

func addAPIsToServer(handler core.Handler, recipesDB recipes.RecipeDB, srcRepository sources.Sources) {
	recipes.AddRecipesAPIToHandler(handler, recipesDB)
	sourcesAPI := sources.NewSourceAPI(srcRepository, recipesDB)
	sourcesAPI.PrepareAPI(handler, srcRepository, recipesDB)
	core.AddCoreAPIToHandler(handler)

	if mongoDB, ok := recipesDB.(*recipes.MongoRecipeDB); ok {
		account.AddAuthAPIsToHandler(handler, mongoDB.MongoClient())
	} else {
		log.Warn("Account API not enabled: underlying database is not MongoDB")
	}
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
		log.WithError(err).Fatal("Cannot open database connection ...")
	}
}
