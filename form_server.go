package main

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"golang.org/x/text/message"
	"net"
)

const (
	imageConnect  = "https://cdn.discordapp.com/attachments/1240581906567790632/1276169643375067200/normal.png?ex=66c88d19&is=66c73b99&hm=66a45478e69ccb336a9629b3d2a97c0fd4136d5fc822f23a942e1bfab0635ba0&"
	imageAdd      = "https://cdn.discordapp.com/attachments/1240581906567790632/1276169643693703269/eyes.png?ex=66c88d19&is=66c73b99&hm=1d6ec8fae2f4dc80cb82d985db9582eab85d96abb44aa80b45957b5e5978ad7f&"
	imageRemove   = "https://cdn.discordapp.com/attachments/1240581906567790632/1276169642829807666/nerd.png?ex=66c88d18&is=66c73b98&hm=e46585c3b856d8bc907c6b6712bce4b482c2e161a45301d998dbd95cc4af4342&"
	imageSettings = "https://cdn.discordapp.com/attachments/1240581906567790632/1276169643069018296/40e8823345030871.png?ex=66c88d18&is=66c73b98&hm=fb37af01b153c27f7bb837d56f1ea25e3b6f7a6a54ec3c0db08ccf442821798a&"
)

func newServerForm(p *message.Printer, prof profile) form.Form {
	buttons, servers := serverButtons(p, prof)

	settings := form.NewButton(
		p.Sprintf("menu.settings.button"),
		imageSettings,
	)
	return form.NewMenu(serverForm{
		Connect: form.NewButton(p.Sprintf("menu.connect.button"), imageConnect),
		Add:     form.NewButton(p.Sprintf("menu.add.button"), imageAdd),
		Remove:  form.NewButton(p.Sprintf("menu.remove.button"), imageRemove),

		settings: settings,

		p:       p,
		prof:    prof,
		servers: servers,
	}, p.Sprintf("menu.title")).WithButtons(append(buttons, settings)...)
}

type serverForm struct {
	Connect  form.Button
	Add      form.Button
	Remove   form.Button
	settings form.Button

	p       *message.Printer
	prof    profile
	servers map[string]serverProfile
}

func (f serverForm) Submit(s form.Submitter, pressed form.Button) {
	switch pressed.Text {
	case f.Connect.Text:
		s.SendForm(newConnectForm(f.p))
	case f.Add.Text:
		s.SendForm(newAddForm(f.p))
	case f.Remove.Text:
		s.SendForm(newRemoveForm(f.p, f.prof))
	case f.settings.Text:
		s.SendForm(newSettingsForm(f.p, f.prof))
	default:
		p, ok := s.(*player.Player)
		if !ok {
			return
		}
		server, ok := f.servers[pressed.Text]
		if !ok {
			return
		}
		_, err := net.ResolveUDPAddr("udp", server.RawAddress)
		if err != nil {
			p.Message(f.p.Sprintf("server.error.resolve", err))
			return
		}
		if p.Transfer(server.RawAddress) != nil {
			return
		}
	}
}

func serverButtons(p *message.Printer, prof profile) ([]form.Button, map[string]serverProfile) {
	buttons := make([]form.Button, len(prof.Servers))
	servers := make(map[string]serverProfile, len(prof.Servers))
	for i, server := range prof.Servers {
		var button form.Button
		if server.Name == "" {
			button.Text = p.Sprintf("server.button.text", server.RawAddress)
		} else {
			button.Text = p.Sprintf("server.button.text.named", server.Name, server.RawAddress)
		}
		servers[button.Text] = server
		buttons[i] = button
	}
	return buttons, servers
}
