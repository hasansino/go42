package aerospike

import (
	aelib "github.com/aerospike/aerospike-client-go/v8"
)

type Option func(*aelib.ClientPolicy)
