package main

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"golang.org/x/text/message"
	"slices"
)

func newRemoveForm(p *message.Printer, prof profile) form.Form {
	buttons, servers := serverButtons(p, prof)
	return form.NewMenu(removeForm{
		p: p,

		servers: servers,
	}, p.Sprintf("menu.remove.title")).WithButtons(buttons...)
}

type removeForm struct {
	p *message.Printer

	servers map[string]serverProfile
}

func (f removeForm) Submit(s form.Submitter, pressed form.Button) {
	server, ok := f.servers[pressed.Text]
	if !ok {
		return
	}

	p, ok := s.(*player.Player)
	if !ok {
		return
	}
	v, ok := handlers.Load(p.UUID())
	if !ok {
		return
	}
	h := v.(*handler)

	prof := h.prof.Load()
	i := slices.IndexFunc(prof.Servers, func(s serverProfile) bool {
		return addressEqual(s.Address, server.Address) || s.RawAddress == server.RawAddress
	})
	if i < 0 {
		return
	}
	prof.Servers = slices.Delete(prof.Servers, i, i+1)
	h.prof.Store(prof)
	p.Message(f.p.Sprintf("menu.remove.success"))
}
