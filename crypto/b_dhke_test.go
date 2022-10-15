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
