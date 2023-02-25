module github.com/jakob-moeller-cloud/octi-sync-server

go 1.19

require (
	// generated openapi
	github.com/deepmap/oapi-codegen v1.11.1-0.20220912230023-4a1477f6a8ba

	// openapi
	github.com/getkin/kin-openapi v0.103.0

	// storage
	github.com/go-redis/redis/v9 v9.0.0-beta.2

	// generated mocks
	github.com/golang/mock v1.6.0

	// id generation
	github.com/google/uuid v1.3.0

	// json parsing
	github.com/json-iterator/go v1.1.12

	// routing
	github.com/labstack/echo/v4 v4.9.0
	github.com/labstack/gommon v0.3.1

	// logging
	github.com/rs/zerolog v1.28.0

	// high-entropy password generation
	github.com/sethvargo/go-password v0.2.0

	// testing assertions
	github.com/stretchr/testify v1.8.0

	// yaml parsing
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/invopop/yaml v0.2.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/time v0.0.0-20220922220347-f3bd1da661af // indirect
)
