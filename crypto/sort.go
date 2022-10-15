package crypto

import "github.com/decred/dcrd/dcrec/secp256k1/v4"

type PublicKey struct {
	Amount int64
	Key    *secp256k1.PublicKey
}

type PublicKeyList []PublicKey

func (s PublicKeyList) GetKeyByAmount(amount int64) *PublicKey {
	for _, key := range s {
		if key.Amount == amount {
			return &key
		}
	}
	return nil
}

func (s PrivateKeyList) GetKeyByAmount(amount int64) *PrivateKey {
	for _, key := range s {
		if key.Amount == amount {
			return &key
		}
	}
	return nil
}

func (s PublicKeyList) Len() int           { return len(s) }
func (s PublicKeyList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s PublicKeyList) Less(i, j int) bool { return s[i].Amount < s[j].Amount }

type PrivateKey struct {
	Amount int64
	Key    *secp256k1.PrivateKey
}

type PrivateKeyList []PrivateKey

func (p PrivateKeyList) Len() int           { return len(p) }
func (p PrivateKeyList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PrivateKeyList) Less(i, j int) bool { return p[i].Amount < p[j].Amount }
