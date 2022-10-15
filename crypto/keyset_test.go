package crypto

import (
	"reflect"
	"testing"
)

func Test_deriveKeys(t *testing.T) {
	type args struct {
		masterKey      string
		derivationPath string
	}
	tests := []struct {
		name string
		args args
		want PrivateKeyList
	}{
		{name: "deriveKeys", args: args{masterKey: "masterkey", derivationPath: "0/0/0/0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path1 := deriveKeys(tt.args.masterKey, tt.args.derivationPath)
			path1Copy := deriveKeys(tt.args.masterKey, tt.args.derivationPath)
			if !reflect.DeepEqual(path1, path1Copy) {
				t.Errorf("unequal derivisions")
			}
		})
	}
}

func Test_deriveKeySetId(t *testing.T) {
	type args struct {
		publicKeys PublicKeyList
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "deriveKeySetId", want: "JHV8eUnoAln/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path1 := deriveKeys("master", "0/0/0/0")
			if got := deriveKeySetId(derivePublicKeys(path1)); got != tt.want {
				t.Errorf("deriveKeySetId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKeySet(t *testing.T) {
	type args struct {
		masterKey      string
		derivationPath string
	}
	tests := []struct {
		name string
		args args
		want *KeySet
	}{
		{name: "NewKeySet", args: args{masterKey: "master", derivationPath: "0/0/0/0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewKeySet(tt.args.masterKey, tt.args.derivationPath)
			if len(got.PublicKeys) != len(got.PrivateKeys) {
				t.Errorf("invalid keysets")
			}
			if got.Id != "JHV8eUnoAln/" {
				t.Errorf("invalid id")
			}
		})
	}
}

func TestKeySet_DeriveKeys(t *testing.T) {

	type args struct {
		masterKey string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "DeriveKeys"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKeySet("master", "0/0/0/0")
			k.DeriveKeys(tt.args.masterKey)
			if len(k.PrivateKeys) == 0 {
				t.Errorf("failed to DeriveKeys")
			}
		})
	}
}

func TestKeySet_DerivePublicKeys(t *testing.T) {

	tests := []struct {
		name string
	}{
		{name: "DerivePublicKeys"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &KeySet{DerivationPath: "0/0/0/0"}
			ks.DeriveKeys("master")
			ks.DerivePublicKeys()
			if len(ks.PublicKeys) == 0 {
				t.Errorf("failed to DeriveKeys")
			}
		})
	}
}

func TestKeySet_DeriveKeySetId(t *testing.T) {

	tests := []struct {
		name string
	}{
		{name: "DeriveKeySetId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KeySet{DerivationPath: "0/0/0/0"}
			k.DeriveKeys("master")
			k.DerivePublicKeys()
			k.DeriveKeySetId()
			if k.Id != "JHV8eUnoAln/" {
				t.Errorf("failed to TestKeySet_DeriveKeySetId")
			}
		})
	}
}
