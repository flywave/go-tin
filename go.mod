module github.com/flywave/go-tin

go 1.23.0

toolchain go1.24.4

require (
	github.com/flywave/go-geo v0.0.0-00010101000000-000000000000
	github.com/flywave/go-geoid v0.0.0-20210705014121-cd8f70cb88bb
	github.com/flywave/go3d v0.0.0-20250619003741-cab1a6ea6de6
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/flywave/go-geom v0.0.0-20250607125323-f685bf20f12c // indirect
	github.com/flywave/go-geos v0.0.0-20210901070302-5537e39a4985 // indirect
	github.com/flywave/go-proj v0.0.0-20210901061921-dbd10648e538 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/flywave/go-geo => ../go-geo

replace github.com/flywave/go-geos => ../go-geos

replace github.com/flywave/go-geoid => ../go-geoid

replace github.com/flywave/go-proj => ../go-proj
