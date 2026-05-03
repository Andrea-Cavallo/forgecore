module github.com/Andrea-Cavallo/golang-modules/services/forgecore-jobs

go 1.26

require (
	github.com/Andrea-Cavallo/golang-modules/shared v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.18.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	go.uber.org/atomic v1.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Andrea-Cavallo/golang-modules/shared => ../../shared
