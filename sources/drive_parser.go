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
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"

	"github.com/ottenwbe/go-cook/recipes"
	"github.com/ottenwbe/go-cook/utils"
)

//driveRecipeParser enumerates all states of the driveRecipeParser
type driveRecipeParser struct {
	descriptionBuffer bytes.Buffer
	recipe            *recipes.Recipe
	pictures          map[string]*recipes.RecipePicture
	lastErr           error
	parseState        parseState
}

//parseState enumerates all states of the driveRecipeParser
type parseState int

const (
	titleState parseState = iota
	imgState
	ingredientState
	descriptionState
)

//ParseRecipe transforms a html file to the Recipe format
func ParseRecipe(htmlFile io.Reader, fileID recipes.RecipeID) (*recipes.Recipe, map[string]*recipes.RecipePicture, error) {
	parser := newDriveRecipeParser(fileID)
	return parser.parseHTML(htmlFile)
}

func newDriveRecipeParser(fileID recipes.RecipeID) *driveRecipeParser {
	return &driveRecipeParser{
		recipe:     recipes.NewRecipe(fileID),
		pictures:   make(map[string]*recipes.RecipePicture),
		lastErr:    nil,
		parseState: titleState,
	}
}

func (p *driveRecipeParser) nextState(state parseState) {
	p.parseState = state
}

func (p *driveRecipeParser) parseHTML(htmlFile io.Reader) (*recipes.Recipe, map[string]*recipes.RecipePicture, error) {
	tokenizer := html.NewTokenizer(htmlFile)

	var (
		parsing = true
		err     error
	)

	for parsing {

		tokenType := tokenizer.Next()

		//Debug
		n, hasMore := tokenizer.TagName()
		tagName := string(n)
		text := string(tokenizer.Text())

		log.Debugf("Parsing: %v (%v): %v", tagName, tokenType, text)

		switch tokenType {
		case html.ErrorToken:

			parsing = false

			err = tokenizer.Err()
			p.finalizeRecipe()

			log.Debugf("Parsing Error: %v (%v): %v", tagName, tokenType, err)

			p.recipe, err = p.handleParsingError(p.recipe, err)

		case html.TextToken:

			p.handleText(text)

		case html.StartTagToken:

			p.handleStartToken(tagName, hasMore, tokenizer)

		default:

		}
	}

	return p.recipe, p.pictures, err
}

func (p *driveRecipeParser) handleStartToken(tagName string, hasMore bool, tokenizer *html.Tokenizer) {
	switch tagName {

	case "p":
		if p.parseState == descriptionState {
			log.Debug("p added to description")
			p.descriptionBuffer.WriteString("\n")
		}
	case "img":

		p.nextState(imgState)
		log.Debug("Entering image state... ")

		name, img64 := extractImg(hasMore, tokenizer)

		p.recipe.PictureLink = append(p.recipe.PictureLink, name)
		log.Debugf("img appending %v", name)
		p.pictures[name] = &recipes.RecipePicture{ID: p.recipe.ID, Name: name, Picture: img64}
	}
}

func extractImg(hasMore bool, tokenizer *html.Tokenizer) (string, string) {
	var (
		name  = ""
		img64 = ""
		err   error
	)
	for hasMore {
		var key, val []byte
		key, val, hasMore = tokenizer.TagAttr()

		if string(key) == "alt" {
			name = string(val)
		} else if string(key) == "src" {
			img64, err = utils.DownloadIMGAsBase64(string(val))
			if err != nil {
				log.WithError(err).Error("Could not download image")
			}
			log.Debugf("new image: %v", img64)
		}
	}
	return name, img64
}

func (p *driveRecipeParser) handleText(text string) {
	if strings.HasPrefix(text, ingredients) {
		log.Debug("Now parsing ingredients...")
		p.nextState(ingredientState)
	} else if strings.HasPrefix(text, instruction) {
		log.Debug("Now parsing the description...")
		p.nextState(descriptionState)
	} else if p.parseState == titleState {
		log.Debugf("Title - add: %v", text)
		p.recipe.Name = text
	} else if p.parseState == ingredientState {
		log.Debugf("Ingredient - add: %v", text)
		handleIngredient(p, text)
	} else if p.parseState == descriptionState {
		log.Debugf("Description - add: %v", text)
		p.descriptionBuffer.WriteString(text)
	}
}

var (
	validNumber  = regexp.MustCompile(`^[0-9]+`)
	validStrings = regexp.MustCompile(`[a-zA-zßäüöÄÜÖ]+`)
)

func handleIngredient(p *driveRecipeParser, text string) {

	num := validNumber.FindAllString(text, -1)
	strs := validStrings.FindAllString(text, -1)

	var amount = -1.0
	var err error
	if len(num) > 0 {
		amount, err = strconv.ParseFloat(num[0], 32)
		if err != nil {
			log.WithError(err).Error("Unexpected error while parsing a number")
		}
	}

	var unit = ""
	if len(strs) > 1 {
		unit = strs[0]
	}

	var name = ""
	if len(strs) == 1 {
		name = strs[0]
	} else if len(strs) > 1 {
		name = strings.Join(strs[1:], " ")
	}

	p.recipe.Ingredients = append(p.recipe.Ingredients, recipes.Ingredients{Name: name, Amount: amount, Unit: unit})
}

func (p *driveRecipeParser) finalizeRecipe() {
	p.recipe.Description = p.descriptionBuffer.String()
}

func (p *driveRecipeParser) handleParsingError(recipe *recipes.Recipe, err error) (*recipes.Recipe, error) {
	if err != io.EOF {
		return recipes.NewInvalidRecipe(), err
	} else if p.parseState != descriptionState {
		return recipes.NewInvalidRecipe(), errors.New("malformed recipe")
	}
	return recipe, nil
}

const (
	driveParserIngredientsTitle  = "drive.connection.secret.file"
	driveRecipeInstructionsTitle = "drive.recipes.folder"
)

var (
	ingredients = ""
	instruction = ""
)

func init() {
	utils.Config.SetDefault(driveParserIngredientsTitle, "Zutaten")
	utils.Config.SetDefault(driveRecipeInstructionsTitle, "Zubereitung")

	ingredients = utils.Config.GetString(driveParserIngredientsTitle)
	instruction = utils.Config.GetString(driveRecipeInstructionsTitle)
}
