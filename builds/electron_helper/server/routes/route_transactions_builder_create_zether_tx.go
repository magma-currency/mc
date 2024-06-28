package routes

import (
	"context"
	"encoding/json"
	"mc/builds/builds_data"
	"mc/txs_builder/wizard"
)

func RouteTransactionsBuilderCreateZetherTx(req []byte) (any, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	txData, transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, publicKeyIndexes, feesFinal, err := builds_data.PrepareData(req)
	if err != nil {
		return nil, err
	}

	tx, err := wizard.CreateZetherTx(transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, txData.ChainKernelHeight, txData.ChainKernelHash, publicKeyIndexes, feesFinal, ctx, func(status string) {})
	if err != nil {
		return nil, err
	}

	txJson, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	return []interface{}{
		txJson,
		tx.Bloom.Serialized,
	}, nil
}
