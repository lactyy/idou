package main

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"golang.org/x/text/message"
)

func newSettingsForm(m *message.Printer, prof profile) form.Form {
	return form.NewMenu(settingsForm{
		Language: form.NewButton(
			m.Sprintf("menu.language.button"),
			"",
		),

		m:    m,
		prof: prof,
	}, m.Sprintf("menu.settings.title"))
}

type settingsForm struct {
	Language form.Button

	m    *message.Printer
	prof profile
}

func (f settingsForm) Submit(s form.Submitter, pressed form.Button) {
	switch pressed.Text {
	case f.Language.Text:
		s.SendForm(newLanguageForm(f.m, f.prof))
	}
}
