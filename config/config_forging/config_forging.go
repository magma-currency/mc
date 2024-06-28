package config_forging

import "mc/config/arguments"

var (
	FORGING_ENABLED = true
)

func InitConfig() (err error) {

	if arguments.Arguments["--forging"] == false {
		FORGING_ENABLED = false
	}

	return
}
