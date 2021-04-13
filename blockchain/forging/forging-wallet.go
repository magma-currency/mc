package forging

import (
	"bytes"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/store"
	"sync"
	"sync/atomic"
)

type ForgingWallet struct {
	addresses    []*ForgingWalletAddress
	addressesMap map[string]*ForgingWalletAddress
	sync.RWMutex `json:"-"`

	updates      atomic.Value //[]*ForgingWalletAddressUpdate
	updatesMutex sync.Mutex
}

type ForgingWalletAddressUpdate struct {
	delegatedPriv []byte
	pubKeyHash    []byte
}

type ForgingWalletAddress struct {
	delegatedPrivateKey    *addresses.PrivateKey
	delegatedPublicKeyHash []byte //20 byte
	publicKeyHash          []byte //20byte
	account                *account.Account
}

func (w *ForgingWallet) AddWallet(delegatedPriv []byte, pubKeyHash []byte) {

	w.updatesMutex.Lock()
	defer w.updatesMutex.Unlock()

	updates := w.updates.Load().([]*ForgingWalletAddressUpdate)
	updates = append(updates, &ForgingWalletAddressUpdate{
		delegatedPriv,
		pubKeyHash,
	})

	w.updates.Store(updates)

	return
}

func (w *ForgingWallet) RemoveWallet(delegatedPublicKeyHash []byte) { //20 byte
	w.AddWallet(nil, delegatedPublicKeyHash)
}

func (w *ForgingWallet) processUpdates() (err error) {

	w.updatesMutex.Lock()
	updates := w.updates.Load().([]*ForgingWalletAddressUpdate)
	w.updates.Store([]*ForgingWalletAddressUpdate{}) //reset with empty
	w.updatesMutex.Unlock()

	w.Lock()
	defer w.Unlock()

	for _, update := range updates {

		key := string(update.pubKeyHash)

		//let's delete it
		if update.delegatedPriv == nil {

			if w.addressesMap[key] != nil {
				delete(w.addressesMap, key)
				for i, address := range w.addresses {
					if bytes.Equal(address.publicKeyHash, update.pubKeyHash) {
						w.addresses = append(w.addresses[:i], w.addresses[:i+1]...)
						break
					}
				}
			}

		} else {

			delegatedPrivateKey := &addresses.PrivateKey{Key: update.delegatedPriv}

			var delegatedPublicKey []byte
			if delegatedPublicKey, err = delegatedPrivateKey.GeneratePublicKey(); err != nil {
				return err
			}

			delegatedPublicKeyHash := cryptography.ComputePublicKeyHash(delegatedPublicKey)

			//let's read the balance
			if err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {

				accs := accounts.NewAccounts(boltTx)
				acc := accs.GetAccount(update.pubKeyHash)

				if acc.DelegatedStake == nil || !bytes.Equal(acc.DelegatedStake.DelegatedPublicKeyHash, delegatedPublicKeyHash) {
					return errors.New("Delegated stake is not matching")
				}

				address := w.addressesMap[key]
				if address == nil {
					address = &ForgingWalletAddress{
						delegatedPrivateKey,
						delegatedPublicKeyHash,
						update.pubKeyHash,
						acc,
					}
					w.addresses = append(w.addresses, address)
					w.addressesMap[key] = address
				} else {
					address.delegatedPrivateKey = delegatedPrivateKey
					address.delegatedPublicKeyHash = delegatedPublicKeyHash
				}

				return
			}); err != nil {
				return
			}

		}

	}

	return
}

func (w *ForgingWallet) UpdateBalanceChanges(accs *accounts.Accounts) {

	w.processUpdates()

	w.Lock()
	defer w.Unlock()

	for k, v := range accs.HashMap.Committed {
		if w.addressesMap[k] != nil {

			if v.Commit == "update" {
				w.addressesMap[k].account = new(account.Account)
				w.addressesMap[k].account.Deserialize(v.Data)
			} else if v.Commit == "delete" {
				w.addressesMap[k].account = nil
			}

		}
	}

}

func (w *ForgingWallet) loadBalances() error {

	w.Lock()
	defer w.Unlock()

	return store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {

		accs := accounts.NewAccounts(boltTx)

		for _, address := range w.addresses {
			account := accs.GetAccount(address.publicKeyHash)
			address.account = account
		}

		return nil
	})

}
