package main

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"golang.org/x/text/message"
	"net"
)

func newConnectForm(p *message.Printer) form.Form {
	return form.New(connectForm{
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
		Remember: form.NewToggle(
			p.Sprintf("menu.connect.remember.text"),
			true,
		),

		m: p,
	}, p.Sprintf("menu.connect.title"))
}

type connectForm struct {
	Address  form.Input
	Name     form.Input
	Remember form.Toggle

	m *message.Printer
}

func (f connectForm) Submit(s form.Submitter) {
	if p, ok := s.(*player.Player); ok {
		if address := f.Address.Value(); address != "" {
			address = addressWithPort(address)
			addr, err := net.ResolveUDPAddr("udp", address)
			if err != nil {
				p.Message(f.m.Sprintf("server.error.resolve", err))
				return
			}
			if p.Transfer(address) != nil {
				return
			}

			if f.Remember.Value() {
				v, ok := handlers.Load(p.UUID())
				if !ok {
					return
				}
				h := v.(*handler)

				prof := h.prof.Load()
				if len(prof.Servers) >= 20 {
					return
				}
				if prof.shouldRemember(addr, address) {
					prof.Servers = append(prof.Servers, serverProfile{
						Name:       serverProfileName(f.Name.Value()),
						Address:    addr,
						RawAddress: address,
					})
					h.prof.Store(prof)
				}
			}
		}
	}
}
