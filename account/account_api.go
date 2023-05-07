package account

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/ottenwbe/recipes-manager/core"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
)

// API for accounts
type API struct {
	handler core.Handler
	db      core.DB
}

var (
	api *API
)

// AddAccountAPIToHandler constructs an API for recipes
func AddAccountAPIToHandler(handler core.Handler, db core.DB) {
	api = &API{
		handler,
		db,
	}

	api.prepareAPI()
}

func (a *API) prepareAPI() {
	provider, err := oidc.NewProvider(oauth2.NoContext, "http://lars-nas:8889/auth/realms/Test")
	if err != nil {
		panic(err)
	}

	var config oauth2.Config

	if a.handler == nil {
		log.WithField("Component", "Account API").Fatal("No handler defined")
		return
	}

	if a.db == nil {
		log.WithField("Component", "Account API").Fatal("No persistence defined")
		return
	}

	v1 := a.handler.API(1)

	//GET the list of accounts
	v1.POST("/accounts", a.postAccounts(config, provider))
}

// postAccounts example
// @Summary Get the number of recipes
// @Description The number of recipes is returned that is managed by the service.
// @Tags Accounts
// @Produce json
// @Success 200 {integer} number
// @Router /accounts [post]
func (a *API) postAccounts(config oauth2.Config, provider *oidc.Provider) func(c *core.APICallContext) {
	return func(c *core.APICallContext) {
		config = oauth2.Config{
			ClientID:     "GoTest",
			ClientSecret: "HnLyz3tDab3DyUxB9QK3UJoKTr8qvAOE",
			RedirectURL:  "http://localhost:8080/oauth",
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "email"},
		}

		var stateSeed uint64
		binary.Read(rand.Reader, binary.LittleEndian, &stateSeed)
		state := fmt.Sprintf("%x", stateSeed)

		authCodeURL := config.AuthCodeURL(state)
		fmt.Printf("Open %s\n", authCodeURL)
		fmt.Println()

		//overallState = state

		c.Redirect(http.StatusFound, authCodeURL)
	}
}
