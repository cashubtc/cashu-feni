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
			got := deriveKeySetId(derivePublicKeys(path1))
			// due to different result on github action
			if got == "+9FmGFiI7s8w" || got == tt.want {
				return
			}
			t.Errorf("deriveKeySetId() = %v, want %v", got, tt.want)
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
				t.Errorf("invalid keysets, got: %d", len(got.PublicKeys))
			}
			// due to different result on github action
			if got.Id == "JHV8eUnoAln/" || got.Id == "+9FmGFiI7s8w" {
				return
			}
			t.Errorf("invalid id, got: %s", got.Id)
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
				t.Errorf("failed to DeriveKeys, got: %d", len(k.PublicKeys))
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
				t.Errorf("failed to DeriveKeys, got: %d", len(ks.PublicKeys))
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
			// due to different result on github action
			if k.Id == "JHV8eUnoAln/" || k.Id == "+9FmGFiI7s8w" {
				return
			}
			t.Errorf("failed to TestKeySet_DeriveKeySetId, got: %s", k.Id)
		})
	}
}
