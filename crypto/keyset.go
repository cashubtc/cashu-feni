package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/samber/lo"
	"math"
	"sort"
	"strconv"
	"time"
)

const MaxOrder = 64

type KeySet struct {
	Id             string
	DerivationPath string
	PublicKeys     PublicKeyList  `gorm:"-"`
	PrivateKeys    PrivateKeyList `gorm:"-"`
	MintUrl        string
	ValidFrom      time.Time
	ValidTo        time.Time
	FirstSeen      time.Time
	Active         bool
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
func deriveKeys(masterKey string, derivationPath string) PrivateKeyList {
	pk := make(PrivateKeyList, 0)
	for i := 0; i < MaxOrder; i++ {
		hasher := sha256.New()
		hasher.Write([]byte(masterKey + derivationPath + strconv.Itoa(i)))
		pk = append(pk, PrivateKey{Amount: uint64(math.Pow(2, float64(i))), Key: secp256k1.PrivKeyFromBytes(hasher.Sum(nil)[:32])})
	}
	sort.Sort(pk)
	return pk
}

// derivePublicKeys will generate public keys for the mint server
func derivePublicKeys(pk PrivateKeyList) PublicKeyList {
	PublicKeys := make(PublicKeyList, 0)
	for _, key := range pk {
		PublicKeys = append(PublicKeys, PublicKey{Amount: key.Amount, Key: key.Key.PubKey()})
	}
	sort.Sort(pk)
	return PublicKeys
}

// deriveKeySetId will derive the keySetId from all public key of a keySet
func deriveKeySetId(publicKeys PublicKeyList) string {
	var publicKeysConcatenated []byte
	// all public keys into concatenated and compressed hex string
	lo.ForEach[PublicKey](publicKeys,
		func(key PublicKey, _ int) {
			publicKeysConcatenated = append(publicKeysConcatenated, []byte(hex.EncodeToString(key.Key.SerializeCompressed()))...)
		})
	// hash and encode
	hasher := sha256.New()
	hasher.Write(publicKeysConcatenated)
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))[:12]
}
