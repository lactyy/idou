package main

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"golang.org/x/text/message"
	"slices"
)

type languageCommand struct {
	Sub cmd.SubCommand `cmd:"language"`

	Value cmd.Optional[supportedLanguage] `cmd:"value"`
}

func (c languageCommand) Run(src cmd.Source, o *cmd.Output) {
	p := src.(*player.Player)
	v, _ := handlers.Load(p.UUID())
	h := v.(*handler)
	m := h.m.Load()

	if value, ok := c.Value.Load(); ok {
		tag, err := language.Parse(string(value))
		if err != nil {
			o.Error(m.Sprintf("command.language.error.parse", err))
			return
		}
		if !slices.Contains(message.DefaultCatalog.Languages(), tag) {
			o.Error(m.Sprintf("language.unsupported", display.Self.Name(tag)))
			return
		}
		h.m.Store(message.NewPrinter(tag))

		prof := h.prof.Load()
		prof.Language = tag
		h.prof.Store(prof)

		o.Printf(h.m.Load().Sprintf("language.changed", display.Self.Name(tag)))
	} else {
		p.SendForm(newLanguageForm(m, h.prof.Load()))
	}
}

func (c languageCommand) Allow(src cmd.Source) bool {
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

type supportedLanguage string

func (supportedLanguage) Type() string { return "language" }

func (e supportedLanguage) Options(cmd.Source) []string {
	supported := message.DefaultCatalog.Languages()
	s := make([]string, len(supported))
	for i, tag := range supported {
		s[i] = tag.String()
	}
	return s
}
