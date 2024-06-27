package app

import (
	"mc/address_balance_decryptor"
	"mc/blockchain"
	"mc/blockchain/forging"
	"mc/gui"
	"mc/mempool"
	"mc/settings"
	"mc/store"
	"mc/wallet"
)

var (
	Settings                *settings.Settings
	Wallet                  *wallet.Wallet
	Forging                 *forging.Forging
	Mempool                 *mempool.Mempool
	AddressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor
	Chain                   *blockchain.Blockchain
)

func Close() {
	store.DBClose()
	gui.GUI.Close()
	Forging.Close()
	Chain.Close()
	Wallet.Close()
}
