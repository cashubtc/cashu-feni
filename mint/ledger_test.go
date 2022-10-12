package mint

import (
	"bytes"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gohumble/cashu-feni/core"
	"reflect"
	"testing"
)

func Test_amountSplit(t *testing.T) {
	type args struct {
		amount int64
	}
	tests := []struct {
		name string
		args args
		want []int64
	}{
		{name: "13", args: args{amount: 13}, want: []int64{1, 4, 8}},
		{name: "12", args: args{amount: 12}, want: []int64{4, 8}},
		{name: "512", args: args{amount: 512}, want: []int64{512}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := amountSplit(tt.args.amount); !reflect.DeepEqual(got, tt.want) {
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
	B_, r := core.FirstStepAlice(secretMessage)
	C_ := core.SecondStepBob(*B_, *a)
	C := core.ThirdStepAlice(*C_, *r, *A)
	fmt.Printf("secretMessage: %s\n", secretMessage)
	if !core.Verify(*a, *C, secretMessage, core.HashToCurve) {
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
	if core.Verify(*a, *CjKey, secretMessage, core.HashToCurve) {
		t.Errorf("verify(a, C + C, secret_msg) should be false == %v\n", true)
		return
	}
	if core.Verify(*a, *A, secretMessage, core.HashToCurve) {
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
