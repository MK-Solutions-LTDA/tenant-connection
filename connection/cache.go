package connection

import (
	"sync"

	"github.com/dgraph-io/ristretto"
)

var (
	Mutex       sync.Mutex
	Connections *ristretto.Cache
)

func init() {
	var err error
	Connections, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // número de contadores de bits
		MaxCost:     1 << 30, // tamanho máximo do cache em bytes
		BufferItems: 64,      // tamanho do buffer interno
	})
	if err != nil {
		panic(err)
	}
}
