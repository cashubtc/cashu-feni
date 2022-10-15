package bitcoin

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"reflect"
	"testing"
)

func TestStep1CarolCreateP2SHAddress(t *testing.T) {
	type args struct {
		txInRedeemScript []byte
	}
	tests := []struct {
		name     string
		args     args
		wantAddr []byte
		wantNet  byte
		wantErr  bool
	}{
		{name: "CrateP2SHAddress", args: args{txInRedeemScript: []byte{1, 2, 3, 4}}, wantNet: byte(5), wantAddr: []byte("3PHDvCPPWN6p8cXu2Nfz4AHBeoR5qdDZpY")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Step1CarolCreateP2SHAddress(tt.args.txInRedeemScript)
			if (err != nil) != tt.wantErr {
				t.Errorf("Step1CarolCreateP2SHAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.EncodeAddress() != string(tt.wantAddr) {
				t.Errorf("Step1CarolCreateP2SHAddress() invalid scriptAddress: %s", got.String())
			}
			if !got.IsForNet(&chaincfg.MainNetParams) {
				t.Errorf("Step1CarolCreateP2SHAddress() invalid scriptAddress: %s", got.Hash160())
			}
		})
	}
}

var txInP2SHAddress = "3PHDvCPPWN6p8cXu2Nfz4AHBeoR5qdDZpY"

func createTransaction() *wire.MsgTx {
	h, err := chainhash.NewHashFromStr(TXID)
	if err != nil {
		panic(err)
	}
	script, err := txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData([]byte(txInP2SHAddress)).AddOp(txscript.OP_EQUAL).Script()
	if err != nil {
		panic(err)
	}
	return &wire.MsgTx{
		TxOut:   append(make([]*wire.TxOut, 0), &wire.TxOut{Value: 50000, PkScript: script}),
		TxIn:    append(make([]*wire.TxIn, 0), &wire.TxIn{PreviousOutPoint: wire.OutPoint{Hash: *h}, Sequence: 4294967295}),
		Version: 1}
}
func TestStep1BobCarolCreateTx(t *testing.T) {
	type args struct {
		txInP2SHAddress []byte
	}

	tests := []struct {
		name    string
		args    args
		want    *wire.MsgTx
		wantErr bool
	}{
		{name: "CreateTx", args: args{txInP2SHAddress: []byte(txInP2SHAddress)},
			want:    createTransaction(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Step1BobCarolCreateTx(tt.args.txInP2SHAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("Step1BobCarolCreateTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Step1BobCarolCreateTx() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStep3BobVerifyScript(t *testing.T) {
	type args struct {
		txInSignature    []byte
		txInRedeemScript []byte
		tx               *wire.MsgTx
	}
	script := "2103f97b00bb7b8d03b0e7c52c7ee1632cb969ccc3a2d7f92c04a935f2ea274c149fac"
	sig := "473044022008971a95f886ea09c0041902cb8848cfc9fbec4f03010340a59743b692df203e02203a36fe87b0f8fa80072a8aad35b600f163d44be671050f5d22c25616be4800b101232103f97b00bb7b8d03b0e7c52c7ee1632cb969ccc3a2d7f92c04a935f2ea274c149fac"
	pk := "a9148fe473b44cfa43ebf77ba6544a20b4f91c3f36c687"
	scriptBytes, err := hex.DecodeString(script)
	if err != nil {
		panic(err)
	}
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		panic(err)
	}
	pkScript, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	tx := createTransaction()
	tx.TxOut[0].PkScript = pkScript
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "VerifyScript", args: args{txInRedeemScript: scriptBytes, txInSignature: sigBytes, tx: tx}},
		{name: "InvliceVerifyScript", wantErr: true, args: args{txInRedeemScript: scriptBytes, txInSignature: sigBytes, tx: createTransaction()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := Step3BobVerifyScript(tt.args.txInSignature, tt.args.txInRedeemScript, tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("Step3BobVerifyScript() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
