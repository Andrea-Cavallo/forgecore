module github.com/Andrea-Cavallo/golang-modules/services/forgecore-admin

go 1.26

require (
	github.com/Andrea-Cavallo/golang-modules/shared v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
)

require gopkg.in/yaml.v3 v3.0.1 // indirect

replace github.com/Andrea-Cavallo/golang-modules/shared => ../../shared
