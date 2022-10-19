package bitcoin

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lightningnetwork/lnd/input"
)

const (
	TXID = "bff785da9f8169f49be92fa95e31f0890c385bfb1bd24d6b94d7900057c617ae"
	COIN = 100_000_000
)

func Step0CarolPrivateKey() *secp256k1.PrivateKey {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil
	}
	return key
}
func Step0CarolCheckSigRedeemScript(C_ secp256k1.PublicKey) []byte {
	TxInRedeemScript, err := txscript.NewScriptBuilder().AddData(C_.SerializeCompressed()).AddOp(txscript.OP_CHECKSIG).Script()
	if err != nil {
		return nil
	}
	return TxInRedeemScript
}

/*
step1_carol_create_p2sh_address
*/
func Step1CarolCreateP2SHAddress(txInRedeemScript []byte) (*btcutil.AddressScriptHash, error) {
	script, err := txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(btcutil.Hash160(txInRedeemScript)).AddOp(txscript.OP_EQUAL).Script()
	if err != nil {
		return nil, err
	}
	return btcutil.NewAddressScriptHashFromHash(script[2:22], &chaincfg.MainNetParams)
}

/*
step1_bob_carol_create_tx
*/
func Step1BobCarolCreateTx(txInP2SHAddress []byte) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(1)
	txin := make([]*wire.TxIn, 0)
	h, err := chainhash.NewHashFromStr(TXID)
	if err != nil {
		return nil, err
	}
	// set the sequence number to uint32 max, because python btc library does this as well.
	txin = append(txin, &wire.TxIn{PreviousOutPoint: *wire.NewOutPoint(h, 0), Sequence: 4294967295})
	tx.TxIn = txin

	txout := make([]*wire.TxOut, 0)
	script, err := txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(txInP2SHAddress).AddOp(txscript.OP_EQUAL).Script()
	if err != nil {
		return nil, err
	}
	txout = append(txout, &wire.TxOut{Value: int64(0.0005 * COIN), PkScript: script})
	tx.TxOut = txout

	return tx, nil

}

func Step2CarolSignTx(txInRedeemScript []byte, privateKey *btcec.PrivateKey) (*wire.TxIn, error) {
	txInP2SHAddress, err := Step1CarolCreateP2SHAddress(txInRedeemScript)
	if err != nil {
		return nil, err
	}
	tx, err := Step1BobCarolCreateTx(txInP2SHAddress.ScriptAddress())
	if err != nil {
		return nil, err
	}
	sig, err := txscript.RawTxInSignature(tx, 0, txInRedeemScript, txscript.SigHashAll, privateKey)
	if err != nil {
		return nil, err
	}
	tx.TxIn[0].SignatureScript = sig
	return tx.TxIn[0], nil

}
func Step3BobVerifyScript(txInSignature, txInRedeemScript []byte, tx *wire.MsgTx) error {
	txInScriptPubKey, err := input.GenerateP2SH(txInRedeemScript)
	if err != nil {
		return err
	}
	// set the received signature script
	tx.TxIn[0].SignatureScript = txInSignature
	if txscript.IsPayToScriptHash(txInScriptPubKey) {
		vm, err := txscript.NewEngine(
			txInScriptPubKey, tx, 0,
			txscript.ScriptBip16, nil, nil,
			0.0005*COIN, txscript.NewCannedPrevOutputFetcher(
				txInRedeemScript, int64(0.0005*COIN),
			))
		if err != nil {
			return err
		}
		err = vm.Execute()
		if err != nil {
			return err
		}
	}
	return nil
}

func VerifyScript(pubScriptKey, sig []byte) (txInP2SHAddress *btcutil.AddressScriptHash, err error) {
	// create p2sh address from public script key
	txInP2SHAddress, err = Step1CarolCreateP2SHAddress(pubScriptKey)
	if err != nil {
		return
	}
	// create transaction
	tx, err := Step1BobCarolCreateTx(txInP2SHAddress.ScriptAddress())
	if err != nil {
		return
	}
	// verify the script
	err = Step3BobVerifyScript(sig, pubScriptKey, tx)
	if err != nil {
		return nil, err
	}
	return
}
