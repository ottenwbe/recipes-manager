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

package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/ottenwbe/go-cook/core"
	"github.com/ottenwbe/go-cook/recipes"
	"github.com/ottenwbe/go-cook/sources"
)

func init() {
	log.Infof("Initializing cooking application version=%v API=%v", core.AppVersion().App, core.AppVersion().API)
}

func main() {

	// configure the cooking app
	recipesDB := newCloseableDatabase()
	defer closeDatabase(recipesDB)

	srcRepository := newSources()

	server := newServer(recipesDB, srcRepository)

	// start the application
	server.Run()
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
	handler := core.NewRouter()
	server := core.NewServerH(handler)
	recipesAPI := recipes.NewRecipesAPI(handler, recipesDB)
	recipesAPI.PrepareAPI()
	sourcesAPI := sources.NewSourceAPI(srcRepository, recipesDB)
	sourcesAPI.PrepareAPI(handler, srcRepository, recipesDB)
	return server
}

func newSources() sources.Sources {
	srcRepository := sources.NewSources()

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
