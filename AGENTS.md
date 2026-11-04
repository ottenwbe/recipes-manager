# AGENTS.md — recipes-manager (Go backend)

Go backend for Recipes Manager: Gin REST API, MongoDB, Viper config, Keycloak auth, Google Drive sources, Swagger.

Module: `github.com/ottenwbe/recipes-manager` (Go 1.26). Application-internal packages live under `internal/` (currently `internal/server`, `internal/db`, `internal/version`, `internal/logging`). The shared recipe domain types live in `models/`. Feature packages live in `account/`, `config/`, `docs/`, `recipes/`, `sources/`; entrypoint is `main.go` at the root. The former monolithic `core/` package has been split into focused packages.

## Go standards

- Follow standard Go conventions: idiomatic naming, small focused functions, explicit error handling (`if err != nil { ... }`), no panics in request paths.
- Every `.go` file starts with an SPDX license identifier as its first line (same convention as the Python recommender service):
  ```go
  // SPDX-License-Identifier: MIT
  ```
  (The project is MIT-licensed — see `LICENSE`. Older files still carry the `/* MIT License - see LICENSE file */` block comment; new and touched files use the SPDX line.)
- Logging via logrus, imported as `log "github.com/sirupsen/logrus"`.
- Format and vet before committing: `make fmt` (gofmt -s) and `make vet` (`go vet ./...`). `make verify` runs all QA steps (mod verify, vet, test).
- Keep dependencies tidy with `go mod tidy`; add new deps via `go get`.
- API endpoints carry Swagger annotations; regenerate docs with `make api-docu` (`swag init`).

## Testing — use Ginkgo

All tests use **Ginkgo v2 + Gomega** (already in `go.mod`). Do not write plain `testing` table tests for new specs.

- Each package has one suite bootstrap file, e.g. `recipes/recipes_suite_test.go`:
  ```go
  package recipes

  import (
      "testing"

      . "github.com/onsi/ginkgo/v2"
      . "github.com/onsi/gomega"
  )

  func TestRecipes(t *testing.T) {
      RegisterFailHandler(Fail)
      RunSpecs(t, "Recipes Suite")
  }
  ```
- Specs live in `*_test.go` files in the same package, using dot imports and BDD style:
  ```go
  var _ = Describe("Utils", func() {
      Context("UniqueSlice", func() {
          It("should remove duplicates", func() {
              Expect(UniqueSlice([]string{"a", "b", "a"})).To(Equal([]string{"a", "b"}))
          })
      })
  })
  ```
- Run tests with `make test` (runs `test.sh`: starts a MongoDB container, then `ginkgo -v -cover ./...`). DB-backed tests require that MongoDB, so prefer `make test` over raw `go test`. For a single package: `ginkgo ./recipes/...`.
- New package? Add a `<pkg>_suite_test.go` bootstrap with `RunSpecs` before adding specs.
