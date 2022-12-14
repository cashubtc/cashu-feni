package crypto

import (
	"reflect"
	"testing"
)

func TestPublicKeyList_ByAmount(t *testing.T) {
	type args struct {
		amount uint64
	}
	tests := []struct {
		name string
		s    PublicKeyList
		args args
		want *PublicKey
	}{
		{name: "TestPublicKeyList_ByAmount", args: args{amount: 2}, s: PublicKeyList{PublicKey{Amount: 1}, PublicKey{Amount: 2}}, want: &PublicKey{Amount: 2}},
		{name: "TestPublicKeyList_ByAmount_nil", args: args{amount: 2}, s: nil, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetKeyByAmount(tt.args.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ByAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrivateKeyList_ByAmount(t *testing.T) {
	type args struct {
		amount uint64
	}
	tests := []struct {
		name string
		s    PrivateKeyList
		args args
		want *PrivateKey
	}{
		{name: "TestPrivateKeyList_ByAmount", args: args{amount: 2}, s: PrivateKeyList{PrivateKey{Amount: 1}, PrivateKey{Amount: 2}}, want: &PrivateKey{Amount: 2}},
		{name: "TestPrivateKeyList_ByAmount_nil", args: args{amount: 2}, s: nil, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetKeyByAmount(tt.args.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ByAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}
