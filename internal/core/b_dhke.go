package core

import (
	"crypto/sha256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

/*
def hash_to_curve(message: bytes):
    """Generates a point from the message hash and checks if the point lies on the curve.
    If it does not, it tries computing a new point from the hash."""
    point = None
    msg_to_hash = message
    while point is None:
        try:
            _hash = hashlib.sha256(msg_to_hash).digest()
            point = PublicKey(b"\x02" + _hash, raw=True)
        except:
            msg_to_hash = _hash
    return point

*/
// hashToCurve will generate a public key on the curve.
func hashToCurve(secretMessage []byte) *secp256k1.PublicKey {
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
func FirstStepAlice(secretMessage string) (*secp256k1.PublicKey, *secp256k1.PrivateKey) {
	Y := hashToCurve([]byte(secretMessage))
	r, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	var pointr, pointy, result secp256k1.JacobianPoint

	r.PubKey().AsJacobian(&pointr)
	Y.AsJacobian(&pointy)
	secp256k1.AddNonConst(&pointr, &pointy, &result)
	result.ToAffine()
	B_ := secp256k1.NewPublicKey(&result.X, &result.Y)
	return B_, r
}

// SecondStepBob signes blinded secrets and produces promises
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
func Verify(a secp256k1.PrivateKey, c secp256k1.PublicKey, secretMessage string) bool {
	var Y, Result secp256k1.JacobianPoint
	k := []byte(secretMessage)
	y := hashToCurve(k)
	y.AsJacobian(&Y)
	secp256k1.ScalarMultNonConst(&a.Key, &Y, &Result)
	Result.ToAffine()
	YMult := secp256k1.NewPublicKey(&Result.X, &Result.Y)
	return c.IsEqual(YMult)
}
