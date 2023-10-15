package wallet

import (
	"reflect"
	"testing"
)

func Test_generateSecrets(t *testing.T) {
	type args struct {
		secret string
		n      int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		//{name: "generateSecretsP2SH", args: args{secret: "P2SH:test", n: 2}, want: []string{"0:test", "1:test"}},
		{name: "generateSecrets", args: args{secret: "test", n: 2}, want: []string{"0:test", "1:test"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateSecrets(tt.args.secret, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateSecrets() = %v, want %v", got, tt.want)
			}
		})
	}
}
