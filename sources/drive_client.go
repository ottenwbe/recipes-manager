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

package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"

	"github.com/ottenwbe/go-life/recipes"
	"github.com/ottenwbe/go-life/utils"
)

//driveRecipes is a cache for recipes from Drive
type driveRecipes struct {
	driveService *drive.Service
	pictures     map[recipes.RecipeID]map[string]*recipes.RecipePicture
	recipes      []*recipes.Recipe
	initializer  sync.Once
}

func newDriveRecipes(driveService *drive.Service) *driveRecipes {
	return &driveRecipes{
		driveService: driveService,
	}
}

func cacheDriveRecipesFromFiles(driveService *drive.Service, fileList *drive.FileList) ([]*recipes.Recipe, map[recipes.RecipeID]map[string]*recipes.RecipePicture) {
	resultRecipes := make([]*recipes.Recipe, 0)
	resultPictures := make(map[recipes.RecipeID]map[string]*recipes.RecipePicture)
	for _, file := range fileList.Files {
		htmlFile, err := downloadHTML(driveService, file.Id)
		if err != nil {
			log.WithError(err).Error("Could not download html...")
		} else {
			recipe, pictures, err := ParseRecipe(htmlFile, recipes.NewRecipeID())
			if err != nil {
				log.WithError(err).Errorf("could not parse recipe")
			}
			resultRecipes = append(resultRecipes, recipe)
			resultPictures = appendPictures(pictures, resultPictures, recipe.ID)
		}
	}
	return resultRecipes, resultPictures
}

func appendPictures(pictures map[string]*recipes.RecipePicture, resultPictures map[recipes.RecipeID]map[string]*recipes.RecipePicture, id recipes.RecipeID) map[recipes.RecipeID]map[string]*recipes.RecipePicture {
	for name, pic := range pictures {
		if _, ok := resultPictures[id]; !ok {
			resultPictures[id] = make(map[string]*recipes.RecipePicture)
		}
		resultPictures[id][name] = pic
	}
	return resultPictures
}

//Insert is not supported for Drive
func (r *driveRecipes) Insert(recipe *recipes.Recipe) error {
	return errors.New(recipes.NotSupportedError)
}

//Remove is not supported for Drive
func (r *driveRecipes) Remove(name string) error {
	return errors.New(recipes.NotSupportedError)
}

//Picture by ID and name
func (r *driveRecipes) Picture(id recipes.RecipeID, name string) *recipes.RecipePicture {
	r.ensureCache()
	if pic, ok := r.pictures[id][name]; ok {
		return pic
	}
	return nil
}

//Pictures of a given recipe
func (r *driveRecipes) Pictures(id recipes.RecipeID) map[string]*recipes.RecipePicture {
	r.ensureCache()
	if pics, ok := r.pictures[id]; ok {
		return pics
	}
	return nil
}

func (r *driveRecipes) ensureCache() {
	r.initializer.Do(func() {
		r.recipes, r.pictures = cacheDriveRecipesFromFiles(r.driveService, getRecipesList(r.driveService))
	})
}

//AddPicture is not supported for Drive
func (r *driveRecipes) AddPicture(pic *recipes.RecipePicture) error {
	return errors.New(recipes.NotSupportedError)
}

//GetByName is not supported for Drive
func (r *driveRecipes) GetByName(name string) (*recipes.Recipe, error) {
	return nil, errors.New(recipes.NotSupportedError)
}

//Get a recipe by ID
func (r *driveRecipes) Get(id recipes.RecipeID) *recipes.Recipe {
	r.ensureCache()
	for _, recipe := range r.recipes {
		if recipe.ID.String() == id.String() {
			return recipe
		}
	}
	return recipes.NewInvalidRecipe()
}

//Num recipes that are found in Drive
func (r *driveRecipes) Num() int64 {
	r.ensureCache()
	return int64(len(r.List()))
}

//Random recipe from Drive
func (r *driveRecipes) Random() *recipes.Recipe {
	r.ensureCache()
	list := r.List()
	lenRecipes := len(list)
	if lenRecipes > 0 {
		n := rand.Int() % lenRecipes
		return list[n]
	}
	return recipes.NewInvalidRecipe()
}

//IDs returns a list of all recipe IDs
func (r *driveRecipes) IDs() []string {
	r.ensureCache()
	recipeNames := make([]string, 0)
	for _, recipe := range r.List() {
		recipeNames = append(recipeNames, recipe.Name)
	}
	return recipeNames
}

//List all recipes
func (r *driveRecipes) List() []*recipes.Recipe {
	r.ensureCache()
	return r.recipes
}

const (
	driveConnectionSecretCfg  = "drive.connection.secret.file"
	driveRecipesFolderNameCfg = "drive.recipes.folder"
)

var (
	clientSecretFile  string
	recipesFolderName string
)

//DriveClient is handling the interaction with Drive
type DriveClient struct {
	driveRecipes recipes.Recipes
	oAuthConfig  *oauth2.Config
}

//ID of this SourceClient
func (c *DriveClient) ID() SourceID {
	id, err := uuid.FromString("9647df42-737e-412d-bfb6-0c95c71f8218")

	if err != nil {
		return SourceID(uuid.Nil)
	}

	return SourceID(id)
}

//Name of this SourceClient
func (c *DriveClient) Name() string {
	return "Google Drive SourceClient"
}

//Version of this SourceClient
func (c *DriveClient) Version() string {
	return "0.1.0"
}

//Recipes of this SourceClient
func (c *DriveClient) Recipes() recipes.Recipes {
	return c.driveRecipes
}

//Refresh cleans the internal cache of recipes and refreshes the token from file
func (c *DriveClient) Refresh() (err error) {
	c.driveRecipes = nil
	c.oAuthConfig, err = c.OAuthLoginConfig()
	if err != nil {
		return
	}
	tok, tokenErr := c.tokenFromFile()
	if tokenErr == nil {
		err = c.configureDriveConnection(tok)
	}
	return
}

//OpenNewGoogleDriveConnection with an empty cache of recipes
func OpenNewGoogleDriveConnection() *DriveClient {
	c := &DriveClient{}
	c.Refresh()
	return c
}

//Connected returns true if there is a connection established to Drive
func (c *DriveClient) Connected() bool {
	return c.driveRecipes != nil
}

//ConnectOAuth gets a new initial Token
func (c *DriveClient) ConnectOAuth(code string) (err error) {
	tok, err := c.getToken(code)
	if err != nil {
		return err
	}
	return c.configureDriveConnection(tok)
}

func (c *DriveClient) getToken(code string) (*oauth2.Token, error) {
	log.WithField("code", code).Debugf("Fetching token from google ...")
	tok, err := c.oAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	err = c.saveToken(tok)
	return tok, err
}

func (c *DriveClient) configureDriveConnection(token *oauth2.Token) (err error) {
	client := c.oAuthConfig.Client(context.Background(), token)
	service, err := drive.New(client)
	if err != nil {
		return err
	}
	c.driveRecipes = newDriveRecipes(service)
	return nil
}

func (c *DriveClient) oAuthLoginConfig() (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, err
	}
	return config, nil
}

//OAuthLoginConfig returns the configuration for the Authentication endpoint
func (c *DriveClient) OAuthLoginConfig() (*oauth2.Config, error) {
	return c.oAuthLoginConfig()
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func (c *DriveClient) tokenFromFile() (*oauth2.Token, error) {
	file := c.tokenFile()
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	defer f.Close()
	return tok, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func (c *DriveClient) saveToken(token *oauth2.Token) error {
	file := c.tokenFile()
	log.Infof("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return json.NewEncoder(f).Encode(token)
}

func (c *DriveClient) tokenFile() string {
	return "token-" + c.ID().String() + ".json"
}

func getRecipesList(srv *drive.Service) *drive.FileList {
	r, err := srv.Files.List().PageSize(10).OrderBy("folder").Do()
	if err != nil {
		log.WithError(err).Fatalf("Unable to retrieve files from drive")
	}
	recipesID := ""
	for len(recipesID) == 0 {

		for _, k := range r.Files {
			if k.Name == recipesFolderName {
				recipesID = k.Id
			}
		}

		if (len(recipesID) == 0) && (len(r.NextPageToken) > 0) {
			r, err = srv.Files.List().PageToken(r.NextPageToken).PageSize(100).OrderBy("folder").Do()
			if err != nil {
				log.Fatalf("Unable to retrieve next files: %v", err)
			}
		}
	}
	r1, err := srv.Files.List().Q(fmt.Sprintf("'%v' in parents", recipesID)).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve recipes folder: %v", err)
	}

	return r1
}

func downloadHTML(srv *drive.Service, id string) (io.Reader, error) {
	log.WithField("file", id).Info("Download from drive\n")

	//document: "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	child, err := srv.Files.Export(id, "text/html").Download()
	if err != nil {
		log.WithError(err).WithField("file", id).Error("Unable to download file")
	}

	return child.Body, err
}

func init() {
	utils.Config.SetDefault(driveConnectionSecretCfg, "client_secret.json")
	utils.Config.SetDefault(driveRecipesFolderNameCfg, "Rezepte Test")

	clientSecretFile = utils.Config.GetString(driveConnectionSecretCfg)
	recipesFolderName = utils.Config.GetString(driveRecipesFolderNameCfg)
}
