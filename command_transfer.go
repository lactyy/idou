package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"net"
	"reflect"
)

type transferCommand struct {
	Address rememberedAddress `cmd:"address"`
}

func (c transferCommand) Run(src cmd.Source, o *cmd.Output) {
	p := src.(*player.Player)
	v, _ := handlers.Load(p.UUID())
	h := v.(*handler)
	m := h.m.Load()

	if err := p.Transfer(addressWithPort(string(c.Address))); err != nil {
		o.Errorf(m.Sprintf("server.error.resolve", err))
	}
}

func (c transferCommand) Allow(src cmd.Source) bool {
	p, ok := src.(*player.Player)
	if !ok {
		o := new(cmd.Output)
		o.Errorf("This command can only be executed from in-game.")
		src.SendCommandOutput(o)
		return false
	}
	_, ok = handlers.Load(p.UUID())
	return ok
}

type rememberedAddress string

func (rememberedAddress) Type() string { return "address" }

func (param rememberedAddress) Parse(line *cmd.Line, v reflect.Value) error {
	arg, ok := line.Next()
	if !ok {
		return cmd.ErrInsufficientArgs
	}
	if _, _, err := net.SplitHostPort(addressWithPort(arg)); err != nil {
		return fmt.Errorf("unable to parse server address: %w", err)
	}
	v.SetString(arg)
	return nil
}

func (param rememberedAddress) Options(src cmd.Source) []string {
	p, ok := src.(*player.Player)
	if !ok {
		return nil
	}
	v, ok := handlers.Load(p.UUID())
	if !ok {
		return nil
	}
	h := v.(*handler)
	servers := h.prof.Load().Servers
	s := make([]string, len(servers))
	for i, server := range servers {
		s[i] = server.Address.String()
	}
	return s
}
