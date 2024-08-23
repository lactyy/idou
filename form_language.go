package main

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"golang.org/x/text/language/display"
	"golang.org/x/text/message"
)

func newLanguageForm(p *message.Printer, prof profile) form.Form {
	languages := message.DefaultCatalog.Languages()
	dropdown := form.Dropdown{
		Text:    p.Sprintf("menu.language.dropdown.text"),
		Options: make([]string, len(languages)),
	}
	for i, tag := range languages {
		if tag == prof.Language {
			dropdown.DefaultIndex = i
		}
		dropdown.Options[i] = p.Sprintf("menu.language.dropdown.option",
			display.Self.Name(tag),
			display.English.Languages().Name(tag),
		)
	}
	return form.New(languageForm{
		Dropdown: dropdown,
	}, p.Sprintf("menu.language.title"))
}

type languageForm struct {
	Dropdown form.Dropdown
}

func (f languageForm) Submit(s form.Submitter) {
	if p, ok := s.(*player.Player); ok {
		v, ok := handlers.Load(p.UUID())
		if !ok {
			return
		}
		h := v.(*handler)
		prof := h.prof.Load()

		tag := message.DefaultCatalog.Languages()[f.Dropdown.Value()]
		if prof.Language == tag {
			return
		}

		m := message.NewPrinter(tag)
		h.m.Store(m)

		prof.Language = tag
		h.prof.Store(prof)
		h.setInventory(p, m)

		p.Message(m.Sprintf("language.changed", display.Self.Name(tag)))
	}
}
