package mint

import (
	"bytes"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gohumble/cashu-feni/crypto"
	"github.com/gohumble/cashu-feni/db"
	"github.com/gohumble/cashu-feni/lightning"
	"github.com/gohumble/cashu-feni/lightning/lnbits"
	"math"
	"os"
	"reflect"
	"testing"
)

func Test_amountSplit(t *testing.T) {
	type args struct {
		amount uint64
	}
	tests := []struct {
		name string
		args args
		want []uint64
	}{
		{name: "13", args: args{amount: 13}, want: []uint64{1, 4, 8}},
		{name: "12", args: args{amount: 12}, want: []uint64{4, 8}},
		{name: "512", args: args{amount: 512}, want: []uint64{512}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AmountSplit(tt.args.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("amountSplit() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_Steps will tests all 3 steps
func Test_Steps(t *testing.T) {
	a, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	A := a.PubKey()
	secretMessage := "HI"
	r, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}

	B_, r := crypto.FirstStepAlice(secretMessage, r)
	C_ := crypto.SecondStepBob(*B_, *a)
	C := crypto.ThirdStepAlice(*C_, *r, *A)
	fmt.Printf("secretMessage: %s\n", secretMessage)
	if !crypto.Verify(*a, *C, secretMessage, crypto.HashToCurve) {
		t.Errorf("verify(a, C, secret_msg) == %v \n", false)
		return
	}
	var CJ, result secp256k1.JacobianPoint
	C.AsJacobian(&CJ)
	secp256k1.AddNonConst(&CJ, &CJ, &result)
	if err != nil {
		panic(err)
	}
	CjKey := secp256k1.NewPublicKey(&result.X, &result.Y)
	if crypto.Verify(*a, *CjKey, secretMessage, crypto.HashToCurve) {
		t.Errorf("verify(a, C + C, secret_msg) should be false == %v\n", true)
		return
	}
	if crypto.Verify(*a, *A, secretMessage, crypto.HashToCurve) {
		t.Errorf("verify(a, A, secret_msg) should be false ==  %v\n", true)
		return
	}
	var A1, A2, A3 secp256k1.JacobianPoint
	A.AsJacobian(&A2)
	A2.ToAffine()
	A.AsJacobian(&A1)
	A1.Y.Negate(1)
	A1.ToAffine()
	A1.Y.Set(A1.Y.Add(&A1.Y))
	A1.ToAffine()
	secp256k1.AddNonConst(&A1, &A2, &A3)
	A2Pub := secp256k1.NewPublicKey(&A3.X, &A3.Y)
	if !bytes.Equal(A2Pub.Y().Add(A2Pub.Y(), A2Pub.Y().Add(A2Pub.Y().Neg(A2Pub.Y()), A2Pub.Y().Neg(A2Pub.Y()))).Bytes(), A2Pub.Y().Neg(A2Pub.Y()).Bytes()) {
		t.Errorf("assert -A -A + A == -A  should be true ==  %v\n", false)
	}
}

func TestMint_LoadKeySet(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want *crypto.KeySet
	}{
		{name: "loadKeySet"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New("master", WithInitialKeySet("0/0/0/0"))
			if m.LoadKeySet("JHV8eUnoAln/") != nil {
				return
			}
			t.Errorf("LoadKeySet()")
		})
	}
}

func TestMint_RequestMint(t *testing.T) {
	type args struct {
		amount uint64
	}
	tests := []struct {
		name    string
		args    args
		want    lightning.Invoice
		wantErr bool
	}{
		{name: "request_mint", args: args{amount: 10}, wantErr: false, want: &lnbits.Invoice{Amount: 10, Hash: "invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LIGHTNING", "false")
			m := New("master", WithStorage(db.NewSqlDatabase()))
			got, err := m.RequestMint(tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestMint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.GetHash() == "" {
				t.Errorf("RequestMint() got = %v, want %v", got, tt.want)
			}
			if got.GetAmount() != tt.args.amount {
				t.Errorf("RequestMint() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_verifyAmount(t *testing.T) {
	type args struct {
		amount uint64
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{name: "verifyAmount", want: 123, args: args{amount: 123}},
		{name: "verifyAmountMax", want: uint64(math.Pow(2, MaxOrder)), args: args{amount: uint64(math.Pow(2, MaxOrder))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := verifyAmount(tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("verifyAmount() got = %v, want %v", got, tt.want)
			}
		})
	}
}
