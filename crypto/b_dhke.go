package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// backwards compatibility with old hash_to_curve < 0.3.3
// LegacyHashToCurve will generate a public key on the curve.
func LegacyHashToCurve(secretMessage []byte) *secp256k1.PublicKey {
	msg := secretMessage
	for {
		hasher := sha256.New()
		hasher.Write(msg)
		h := hex.EncodeToString(hasher.Sum(nil))
		hash := []byte(h)[:33]
		hash[0] = 0x02
		pub, err := secp256k1.ParsePubKey(hash)
		if err != nil {
			msg = hash
			continue
		}
		if pub.IsOnCurve() {
			return pub
		}
		continue
	}

}

// HashToCurve will generate a public key on the curve.
func HashToCurve(secretMessage []byte) *secp256k1.PublicKey {
	msg := secretMessage
	for {
		hasher := sha256.New()
		hasher.Write(msg)
		hash := hasher.Sum(nil)
		point, err := secp256k1.ParsePubKey(append([]byte{0x02}, hash...))
		if err != nil {
			msg = hash
			continue
		}
		if point.IsOnCurve() {
			return point
		}
		continue
	}
}

// FirstStepAlice creates blinded secrets and produces outputs
func FirstStepAlice(secretMessage string, r *secp256k1.PrivateKey) (*secp256k1.PublicKey, *secp256k1.PrivateKey) {
	Y := HashToCurve([]byte(secretMessage))
	var pointr, pointy, result secp256k1.JacobianPoint

	r.PubKey().AsJacobian(&pointr)
	Y.AsJacobian(&pointy)
	secp256k1.AddNonConst(&pointr, &pointy, &result)
	result.ToAffine()
	B_ := secp256k1.NewPublicKey(&result.X, &result.Y)
	return B_, r
}

// SecondStepBob signs blinded secrets and produces promises
func SecondStepBob(B_ secp256k1.PublicKey, a secp256k1.PrivateKey) *secp256k1.PublicKey {
	var pointB_, Cp_ secp256k1.JacobianPoint
	B_.AsJacobian(&pointB_)
	secp256k1.ScalarMultNonConst(&a.Key, &pointB_, &Cp_)
	Cp_.ToAffine()
	C_ := secp256k1.NewPublicKey(&Cp_.X, &Cp_.Y)
	return C_
}

// ThirdStepAlice Alice unbinds blinded signatures and produces proofs
func ThirdStepAlice(c_ secp256k1.PublicKey, r secp256k1.PrivateKey, A secp256k1.PublicKey) *secp256k1.PublicKey {
	var pointA, AMult, C_, Cp secp256k1.JacobianPoint
	A.AsJacobian(&pointA)

	secp256k1.ScalarMultNonConst(r.Key.Negate(), &pointA, &AMult)
	c_.AsJacobian(&C_)
	secp256k1.AddNonConst(&AMult, &C_, &Cp)
	Cp.ToAffine()
	return secp256k1.NewPublicKey(&Cp.X, &Cp.Y)
}

// Verify that secret was signed by bob.
// curveFunc should be legacyHashToCurve for client version < 0.3.3
// this could be removed and replaced with static invocation of hashToCurve
func Verify(a secp256k1.PrivateKey, c secp256k1.PublicKey, secretMessage string, curveFunc func(msg []byte) *secp256k1.PublicKey) bool {
	var Y, Result secp256k1.JacobianPoint
	k := []byte(secretMessage)
	y := curveFunc(k)
	y.AsJacobian(&Y)
	secp256k1.ScalarMultNonConst(&a.Key, &Y, &Result)
	Result.ToAffine()
	YMult := secp256k1.NewPublicKey(&Result.X, &Result.Y)
	return c.IsEqual(YMult)
}
