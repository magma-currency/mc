package wallet

import (
	"errors"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
)

type Wallet struct {
	Version   Version
	Mnemonic  string
	Seed      [32]byte
	SeedIndex uint32
	Count     int
	Addresses []*WalletAddress `json:"-"`
}

var wallet Wallet

func GetAddresses() []*WalletAddress {
	return wallet.Addresses
}

func addNewAddress() (err error) {

	masterKey, _ := bip32.NewMasterKey(wallet.Seed[:])

	var key *bip32.Key
	if key, err = masterKey.NewChildKey(wallet.SeedIndex); err != nil {
		gui.Fatal("Couldn't derivate the marker key", err)
	}

	privateKey := addresses.PrivateKey{Key: key.Key}

	var publicKey []byte
	if publicKey, err = privateKey.GeneratePublicKey(); err != nil {
		gui.Fatal("Generating Public Key from Private key raised an error", err)
	}

	var address *addresses.Address
	if address, err = privateKey.GenerateAddress(true, 0, []byte{}); err != nil {
		gui.Fatal("Generating Address raised an error", err)
	}

	var publicKeyHash [20]byte
	copy(publicKeyHash[:], crypto.ComputePublicKeyHash(publicKey))

	var finalPublicKey [33]byte
	copy(finalPublicKey[:], publicKey)
	walletAddress := WalletAddress{
		Name:          "Addr " + strconv.Itoa(wallet.Count),
		PrivateKey:    &privateKey,
		PublicKeyHash: publicKeyHash,
		PublicKey:     finalPublicKey,
		Address:       address,
		SeedIndex:     wallet.SeedIndex,
	}

	wallet.Addresses = append(wallet.Addresses, &walletAddress)
	wallet.Count += 1
	wallet.SeedIndex += 1

	updateWallet()
	return saveWallet()
}

func removeAddress(index int) error {
	if index < 0 || index > len(wallet.Addresses) {
		return errors.New("Invalid Address Index")
	}
	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	wallet.Count -= 1

	updateWallet()
	return saveWallet()
}

func showPrivateKey(index int) ([]byte, error) {
	if index < 0 || index > len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return wallet.Addresses[index].PrivateKey.Key, nil
}

func createSeed() (err error) {

	var entropy []byte
	if entropy, err = bip39.NewEntropy(256); err != nil {
		return gui.Error("Entropy of the address raised an error", err)
	}

	var mnemonic string
	if mnemonic, err = bip39.NewMnemonic(entropy); err != nil {
		return gui.Error("Mnemonic couldn't be created", err)
	}
	wallet.Mnemonic = mnemonic

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "SEED Secret Passphrase")
	wallet.Seed = *helpers.Byte32(seed)

	return nil
}

func createEmptyWallet() error {
	wallet = Wallet{}

	if err := createSeed(); err != nil {
		return gui.Error("Error creating seed", err)
	}
	return addNewAddress()
}

func updateWallet() {
	gui.InfoUpdate("Wallet", walletSaved.Encrypted.String())
	gui.InfoUpdate("Wallet Addrs", strconv.Itoa(wallet.Count))
}
