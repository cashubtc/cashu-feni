package crypto

import (
	"encoding/hex"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"reflect"
	"testing"
)

func TestHashToCurve(t *testing.T) {
	type args struct {
		secretMessage []byte
	}
	pk, err := hex.DecodeString("049595c9df90075148eb06860365df33584b75bff782a510c6cd4883a419833d50bbf2e883bdb76cdbb58e57fc0a2df3bcadf9413358a603d7485d572589df9676")
	if err != nil {
		panic(err)
	}
	key, err := secp256k1.ParsePubKey(pk)
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name string
		args args
		want *secp256k1.PublicKey
	}{
		{name: "h2c", args: args{secretMessage: []byte("hello")}, want: key},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashToCurve(tt.args.secretMessage); !reflect.DeepEqual(got, tt.want) {
				fmt.Printf("%x\n", got.SerializeUncompressed())
				t.Errorf("HashToCurve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirstStepAlice(t *testing.T) {
	type args struct {
		secretMessage string
	}
	PK, err := hex.DecodeString("0249eb5dbb4fac2750991cf18083388c6ef76cde9537a6ac6f3e6679d35cdf4b0c")
	if err != nil {
		panic(err)
	}
	publicKey, err := secp256k1.ParsePubKey(PK)
	if err != nil {
		panic(err)
	}
	pk, err := hex.DecodeString("6d7e0abffc83267de28ed8ecc8760f17697e51252e13333ba69b4ddad1f95d05")
	if err != nil {
		panic(err)
	}
	privateKey := secp256k1.PrivKeyFromBytes(pk)
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name  string
		args  args
		want  *secp256k1.PublicKey
		want1 *secp256k1.PrivateKey
	}{
		{name: "firstStepAlice", args: args{secretMessage: "hello"}, want: publicKey, want1: privateKey},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FirstStepAlice(tt.args.secretMessage, privateKey)
			fmt.Printf("%x\n", got.SerializeUncompressed())
			fmt.Printf("%x\n", got1.Serialize())
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FirstStepAlice() got = %x, want %x", got.SerializeUncompressed(), tt.want.SerializeUncompressed())
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("FirstStepAlice() got1 = %x, want %x", got1.Serialize(), tt.want1.Serialize())
			}
		})
	}
}

func TestSecondStepBob(t *testing.T) {
	type args struct {
		B_ secp256k1.PublicKey
		a  secp256k1.PrivateKey
	}
	r, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}

	publicKey, privateKey := FirstStepAlice("hello", r)

	tests := []struct {
		name string
		args args
		want *secp256k1.PublicKey
	}{
		{name: "SecondStepBob", args: args{B_: *publicKey, a: *privateKey}, want: publicKey},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SecondStepBob(tt.args.B_, tt.args.a)
			if !got.IsOnCurve() || !publicKey.IsOnCurve() {
				t.Errorf("SecondStepBob() not on curve")
			}
			alice := ThirdStepAlice(*got, *privateKey, *r.PubKey())
			if !Verify(*privateKey, *alice, "hello", HashToCurve) {
				t.Errorf("verify(a, C, secret_msg) == %v \n", false)
				return
			}
		})
	}
}
