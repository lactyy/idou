package main

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"golang.org/x/text/message"
	"net"
)

func newAddForm(p *message.Printer) form.Form {
	return form.New(addForm{
		Address: form.NewInput(
			p.Sprintf("server.address.text"),
			"",
			p.Sprintf("server.address.placeholder"),
		),
		Name: form.NewInput(
			p.Sprintf("server.name.text"),
			"",
			"",
		),

		m: p,
	}, p.Sprintf("menu.add.title"))
}

type addForm struct {
	Address form.Input
	Name    form.Input

	m *message.Printer
}

func (f addForm) Submit(s form.Submitter) {
	p, ok := s.(*player.Player)
	if !ok {
		return
	}
	if address := f.Address.Value(); address != "" {
		address = addressWithPort(address)
		addr, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			p.Message(f.m.Sprintf("server.error.resolve", err))
			return
		}

		v, ok := handlers.Load(p.UUID())
		if !ok {
			return
		}
		h := v.(*handler)

		prof := h.prof.Load()
		if len(prof.Servers) >= 20 {
			h.p.Message(f.m.Sprintf("server.error.limit", 20))
			return
		}
		if prof.shouldRemember(addr, address) {
			prof.Servers = append(prof.Servers, serverProfile{
				Name:       serverProfileName(f.Name.Value()),
				Address:    addr,
				RawAddress: address,
			})
			h.prof.Store(prof)
			h.p.Message(f.m.Sprintf("menu.add.success"))
		} else {
			h.p.Message(f.m.Sprintf("menu.add.alreadyExists"))
		}
	}
}
