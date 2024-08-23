package main

import (
	"github.com/df-mc/dragonfly/server/cmd"
)

func init() {
	cmd.Register(cmd.New("transfer", "Transfers you to a server", nil, transferCommand{}))
	cmd.Register(cmd.New("settings", "Change settings", nil, settingsCommand{}, languageCommand{}))
}
