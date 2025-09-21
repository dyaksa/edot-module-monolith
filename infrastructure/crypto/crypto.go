package crypto

import (
	"github.com/dyaksa/encryption-pii/crypto"
	"github.com/dyaksa/encryption-pii/crypto/aesx"
	"github.com/dyaksa/encryption-pii/crypto/core"
)

type Crypto interface {
	AESFunc() func() (core.PrimitiveAES, error)
	Encrypt(data string) aesx.AES[string, core.PrimitiveAES]
	Decrypt(def string) aesx.AES[string, core.PrimitiveAES]
	BindHeap(entity any) error
	HashString(s string) string
}

type derivaleCrypto struct {
	c *crypto.Crypto
}

func (d *derivaleCrypto) Encrypt(data string) aesx.AES[string, core.PrimitiveAES] {
	return d.c.Encrypt(data, aesx.AesCBC)
}

func (d *derivaleCrypto) BindHeap(entity any) error {
	return d.c.BindHeap(entity)
}

func (d *derivaleCrypto) Decrypt(def string) aesx.AES[string, core.PrimitiveAES] {
	return aesx.AESChiper(d.c.AESFunc(), def, aesx.AesCBC)
}

func (d *derivaleCrypto) AESFunc() func() (core.PrimitiveAES, error) {
	return d.c.AESFunc()
}

func (d *derivaleCrypto) HashString(s string) string {
	return d.c.HashString(s)
}

func New() (Crypto, error) {
	c, err := crypto.New(
		crypto.Aes256KeySize,
		crypto.WithInitHeapConnection(),
	)

	if err != nil {
		return nil, err
	}

	return &derivaleCrypto{c: c}, nil
}
