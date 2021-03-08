package transaction_simple

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/config"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	TxScript TransactionSimpleScriptType
	Nonce    uint64
	Vin      []*TransactionSimpleInput
	Vout     []*TransactionSimpleOutput
	Extra    interface{}
}

func (tx *TransactionSimple) IncludeTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) {

	for i, vin := range tx.Vin {

		acc := accs.GetAccountEvenEmpty(vin.GetPublicKeyHash())

		if i == 0 {
			if acc.Nonce != tx.Nonce {
				panic("Account nonce doesn't match")
			}
			acc.IncrementNonce(true)
			switch tx.TxScript {
			case TxSimpleScriptDelegate:
				tx.Extra.(*transaction_simple_extra.TransactionSimpleDelegate).IncludeTransactionVin0(blockHeight, acc)
			case TxSimpleScriptUnstake:
				tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).IncludeTransactionVin0(blockHeight, acc)
			case TxSimpleScriptWithdraw:
				tx.Extra.(*transaction_simple_extra.TransactionSimpleWithdraw).IncludeTransactionVin0(blockHeight, acc)
			}
		}

		acc.AddBalance(false, vin.Amount, vin.Token)
		accs.UpdateAccount(vin.GetPublicKeyHash(), blockHeight, acc)
	}

	for _, vout := range tx.Vout {
		acc := accs.GetAccountEvenEmpty(vout.PublicKeyHash)
		acc.AddBalance(true, vout.Amount, vout.Token)
		accs.UpdateAccount(vout.PublicKeyHash, blockHeight, acc)
	}

	switch tx.TxScript {
	}

}

func (tx *TransactionSimple) RemoveTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) {

	switch tx.TxScript {
	}

	for i := len(tx.Vout) - 1; i >= 0; i-- {
		vout := tx.Vout[i]
		acc := accs.GetAccountEvenEmpty(vout.PublicKeyHash)
		acc.AddBalance(false, vout.Amount, vout.Token)
		accs.UpdateAccount(vout.PublicKeyHash, blockHeight, acc)
	}

	for i := len(tx.Vin) - 1; i >= 0; i-- {
		vin := tx.Vin[i]
		acc := accs.GetAccountEvenEmpty(vin.GetPublicKeyHash())

		if i == 0 {
			switch tx.TxScript {
			case TxSimpleScriptDelegate:
				tx.Extra.(*transaction_simple_extra.TransactionSimpleWithdraw).RemoveTransactionVin0(blockHeight, acc)
			case TxSimpleScriptUnstake:
				tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).RemoveTransactionVin0(blockHeight, acc)
			case TxSimpleScriptWithdraw:
				tx.Extra.(*transaction_simple_extra.TransactionSimpleWithdraw).RemoveTransactionVin0(blockHeight, acc)
			}

			acc.IncrementNonce(false)
			if acc.Nonce != tx.Nonce {
				panic("Account nonce doesn't match")
			}
		}

		acc.AddBalance(true, vin.Amount, vin.Token)
		accs.UpdateAccount(vin.GetPublicKeyHash(), blockHeight, acc)
	}

}

func (tx *TransactionSimple) ComputeFees(out map[string]uint64) {

	tx.ComputeVin(out)
	tx.ComputeVout(out)

	switch tx.TxScript {
	case TxSimpleScriptUnstake:
		helpers.SafeMapUint64Add(out, config.NATIVE_TOKEN_STRING, tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).UnstakeFeeExtra)
	case TxSimpleScriptWithdraw:
		helpers.SafeMapUint64Add(out, config.NATIVE_TOKEN_STRING, tx.Extra.(*transaction_simple_extra.TransactionSimpleWithdraw).WithdrawFeeExtra)
	}
	return
}

func (tx *TransactionSimple) ComputeVin(out map[string]uint64) {
	for _, vin := range tx.Vin {
		helpers.SafeMapUint64Add(out, string(vin.Token), vin.Amount)
	}
}

func (tx *TransactionSimple) ComputeVout(out map[string]uint64) {
	for _, vout := range tx.Vout {
		token := string(vout.Token)
		helpers.SafeMapUint64Sub(out, token, vout.Amount)
		if out[token] == 0 {
			delete(out, token)
		}
	}
}

func (tx *TransactionSimple) VerifySignature(hash helpers.Hash) bool {
	if len(tx.Vin) == 0 {
		return false
	}

	for _, vin := range tx.Vin {
		if ecdsa.VerifySignature(vin.PublicKey[:], hash[:], vin.Signature[0:64]) == false {
			return false
		}
	}
	return true
}

func (tx *TransactionSimple) Validate() {

	switch tx.TxScript {
	case TxSimpleScriptNormal:
		if len(tx.Vin) == 0 || len(tx.Vin) > 255 {
			panic("Invalid vin")
		}
		if len(tx.Vout) == 0 || len(tx.Vout) > 255 {
			panic("Invalid vout")
		}
	case TxSimpleScriptDelegate, TxSimpleScriptUnstake, TxSimpleScriptWithdraw:
		if len(tx.Vin) != 1 {
			panic("Invalid vin")
		}
		if len(tx.Vout) != 0 {
			panic("Invalid vout")
		}
	default:
		panic("Invalid TxScript")
	}

	switch tx.TxScript {
	case TxSimpleScriptDelegate:
		tx.Extra.(*transaction_simple_extra.TransactionSimpleDelegate).Validate()
	case TxSimpleScriptUnstake:
		tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).Validate()
	case TxSimpleScriptWithdraw:
		tx.Extra.(*transaction_simple_extra.TransactionSimpleWithdraw).Validate()
	}

	final := make(map[string]uint64)
	tx.ComputeVin(final)
	tx.ComputeVout(final)
}

func (tx *TransactionSimple) Serialize(writer *helpers.BufferWriter, inclSignature bool) {

	writer.WriteUvarint(uint64(tx.TxScript))
	writer.WriteUvarint(tx.Nonce)

	writer.WriteUvarint(uint64(len(tx.Vin)))
	for _, vin := range tx.Vin {
		vin.Serialize(writer, inclSignature)
	}

	writer.WriteUvarint(uint64(len(tx.Vout)))
	for _, vout := range tx.Vout {
		vout.Serialize(writer)
	}

	switch tx.TxScript {
	case TxSimpleScriptDelegate:
		tx.Extra.(*transaction_simple_extra.TransactionSimpleDelegate).Serialize(writer)
	case TxSimpleScriptUnstake:
		tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).Serialize(writer)
	case TxSimpleScriptWithdraw:
		tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).Serialize(writer)
	}
}

func (tx *TransactionSimple) Deserialize(reader *helpers.BufferReader) {

	n := reader.ReadUvarint()
	tx.TxScript = TransactionSimpleScriptType(n)
	tx.Nonce = reader.ReadUvarint()

	n = reader.ReadUvarint()
	for i := 0; i < int(n); i++ {
		vin := &TransactionSimpleInput{}
		vin.Deserialize(reader)
		tx.Vin = append(tx.Vin, vin)
	}

	//vout only TransactionTypeSimple
	n = reader.ReadUvarint()
	for i := 0; i < int(n); i++ {
		vout := &TransactionSimpleOutput{}
		vout.Deserialize(reader)
		tx.Vout = append(tx.Vout, vout)
	}

	switch tx.TxScript {
	case TxSimpleScriptDelegate:
		extra := &transaction_simple_extra.TransactionSimpleDelegate{}
		extra.Deserialize(reader)
		tx.Extra = extra
	case TxSimpleScriptUnstake:
		extra := &transaction_simple_extra.TransactionSimpleUnstake{}
		extra.Deserialize(reader)
		tx.Extra = extra
	case TxSimpleScriptWithdraw:
		extra := &transaction_simple_extra.TransactionSimpleWithdraw{}
		extra.Deserialize(reader)
		tx.Extra = extra
	}

	return
}
