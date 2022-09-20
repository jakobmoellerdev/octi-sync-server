module github.com/jakob-moeller-cloud/octi-sync-server

go 1.19

require (
	// storage
	github.com/go-redis/redis/v9 v9.0.0-beta.2

	// id generation
	github.com/google/uuid v1.3.0

	// routing
	github.com/labstack/echo/v4 v4.9.0

	// logging
	github.com/rs/zerolog v1.28.0

	// yaml+json parsing
	gopkg.in/yaml.v3 v3.0.1
)

// testing assertions
require github.com/stretchr/testify v1.8.0

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/labstack/gommon v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	golang.org/x/crypto v0.0.0-20220919173607-35f4265a4bc0 // indirect
	golang.org/x/net v0.0.0-20220919232410-f2f64ebce3c1 // indirect
	golang.org/x/sys v0.0.0-20220919091848-fb04ddd9f9c8 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220920022843-2ce7c2934d45 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
