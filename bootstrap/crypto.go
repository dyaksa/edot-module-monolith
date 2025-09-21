package bootstrap

import (
	"github.com/dyaksa/warehouse/infrastructure/crypto"
	"github.com/dyaksa/warehouse/pkg/log"
)

func NewDerivaleCrypto(l log.Logger) crypto.Crypto {
	c, err := crypto.New()
	if err != nil {
		l.Error("failed to create crypto", log.Error("error", err))
	}

	return c
}
