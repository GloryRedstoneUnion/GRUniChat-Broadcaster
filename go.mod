module GRUniChat-Broadcaster

go 1.21

require (
	github.com/go-sql-driver/mysql v1.7.1
	github.com/gorilla/websocket v1.5.1
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.7.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace github.com/gorilla/websocket => github.com/gorilla/websocket v1.5.0
