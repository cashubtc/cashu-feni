package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/samber/lo"
	"math"
	"strconv"
	"time"
)

const MaxOrder = 64

type KeySet struct {
	Id             string
	DerivationPath string
	PublicKeys     map[int64]*secp256k1.PublicKey  `gorm:"-"`
	PrivateKeys    map[int64]*secp256k1.PrivateKey `gorm:"-"`
	MintUrl        string
	ValidFrom      time.Time
	ValidTo        time.Time
	FirstSeen      time.Time
	Active         time.Time
}

func NewKeySet(masterKey, derivationPath string) *KeySet {
	ks := &KeySet{DerivationPath: derivationPath}
	ks.DeriveKeys(masterKey)
	ks.DerivePublicKeys()
	ks.DeriveKeySetId()
	return ks
}

func (k *KeySet) DeriveKeys(masterKey string) {
	k.PrivateKeys = deriveKeys(masterKey, k.DerivationPath)
}

func (k *KeySet) DerivePublicKeys() {
	k.PublicKeys = derivePublicKeys(k.PrivateKeys)
}

func (k *KeySet) DeriveKeySetId() {
	k.Id = deriveKeySetId(k.PublicKeys)
}

// deriveKeys will generate private keys for the mint server
func deriveKeys(masterKey string, derivationPath string) map[int64]*secp256k1.PrivateKey {
	pk := make(map[int64]*secp256k1.PrivateKey, 0)
	for i := 0; i < MaxOrder; i++ {
		hasher := sha256.New()
		hasher.Write([]byte(masterKey + derivationPath + strconv.Itoa(i)))
		pk[int64(math.Pow(2, float64(i)))] = secp256k1.PrivKeyFromBytes(hasher.Sum(nil)[:32])
	}
	return pk
}

// derivePublicKeys will generate public keys for the mint server
func derivePublicKeys(pk map[int64]*secp256k1.PrivateKey) map[int64]*secp256k1.PublicKey {
	PublicKeys := make(map[int64]*secp256k1.PublicKey, 0)
	for amt, key := range pk {
		PublicKeys[amt] = key.PubKey()
	}
	return PublicKeys
}

// deriveKeySetId will derive the keySetId from all public key of a keySet
func deriveKeySetId(publicKeys map[int64]*secp256k1.PublicKey) string {
	var publicKeysConcatenated []byte
	// all public keys into concatenated and compressed hex string
	lo.ForEach[*secp256k1.PublicKey](lo.Values[int64, *secp256k1.PublicKey](publicKeys),
		func(key *secp256k1.PublicKey, _ int) {
			publicKeysConcatenated = append(publicKeysConcatenated, []byte(hex.EncodeToString(key.SerializeCompressed()))...)
		})
	// hash and encode
	hasher := sha256.New()
	hasher.Write(publicKeysConcatenated)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))[:12]
}
