package main

import (
	"os"
	"os/signal"
	"mc/address_balance_decryptor"
	"mc/builds/electron_helper/server"
	"mc/builds/electron_helper/server/global"
	"mc/config"
	"mc/config/arguments"
	"mc/gui"
	"syscall"
)

func main() {
	var err error

	argv := os.Args[1:]
	if err = arguments.InitArguments(argv); err != nil {
		panic(err)
	}
	if err = config.InitConfig(); err != nil {
		panic(err)
	}

	if err = gui.InitGUI(); err != nil {
		panic(err)
	}

	if global.AddressBalanceDecryptor, err = address_balance_decryptor.NewAddressBalanceDecryptor(false); err != nil {
		return
	}

	if err = server.CreateServer(); err != nil {
		panic(err)
	}

	exitSignal := make(chan os.Signal, 10)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

}
