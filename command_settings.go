package main

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
)

type settingsCommand struct{}

func (c settingsCommand) Run(src cmd.Source, o *cmd.Output) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	v, ok := handlers.Load(p.UUID())
	if !ok {
		return
	}
	h := v.(*handler)

	p.SendForm(newSettingsForm(h.m.Load(), h.prof.Load()))
}
