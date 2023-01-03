module github.com/etesync/etebase-go

go 1.14

require (
	github.com/google/uuid v1.1.2
	github.com/stretchr/testify v1.6.1
	github.com/vmihailenco/msgpack/v5 v5.0.0
	golang.org/x/crypto v0.0.0-20201116153603-4be66e5b6582
	golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f // indirect
)

retract (
	v0.1.0 // security: insecure random number generation, leading to predictable account keys
	v0.0.3 // security: insecure random number generation, leading to predictable account keys
	v0.0.2 // security: insecure random number generation, leading to predictable account keys
	v0.0.1 // security: insecure random number generation, leading to predictable account keys
)
