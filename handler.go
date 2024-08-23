package main

import (
	atomic2 "github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/lactyy/idou/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/message"
	"strings"
	"sync"
	"sync/atomic"
	_ "unsafe"
)

type handler struct {
	p   *player.Player
	log *logrus.Logger

	m *atomic.Pointer[message.Printer]

	prof   *atomic2.Value[profile]
	storer internal.Storer[uuid.UUID, profile] // Store that profile on quit

	player.NopHandler
}

func (h *handler) setInventory(p *player.Player, m *message.Printer) {
	_ = p.Inventory().SetItem(
		4,
		item.NewStack(
			compass{},
			1,
		).WithCustomName(
			"§r§f"+m.Sprintf("item.menu"),
		).WithValue(
			"menu",
			true,
		),
	)
}

func (h *handler) HandleChat(ctx *event.Context, message *string) {
	ctx.Cancel()
}

func (h *handler) HandleItemUse(ctx *event.Context) {
	held, _ := h.p.HeldItems()
	if _, ok := held.Value("menu"); ok {
		h.p.SendForm(newServerForm(h.m.Load(), h.prof.Load()))
	}
	ctx.Cancel()
}

func (h *handler) HandleQuit() {
	if err := h.storer.Store(h.p.UUID(), h.prof.Load()); err != nil {
		h.log.Errorf("error storing profile: %s", err)
	}
	handlers.Delete(h.p.UUID())
}

func addressWithPort(s string) string {
	if strings.Count(s, ":") == 0 {
		s += ":19132"
	}
	return s
}

var handlers sync.Map

type visitor struct{}

func (visitor) AllowsEditing() bool      { return false }
func (visitor) AllowsTakingDamage() bool { return false }
func (visitor) CreativeInventory() bool  { return false }
func (visitor) HasCollision() bool       { return true } // Returning false will make this game mode spectator when sent to players, and we don't want to make the slots invisible.
func (visitor) AllowsFlying() bool       { return false }
func (visitor) AllowsInteraction() bool  { return true }
func (visitor) Visible() bool            { return false }

type playerProvider struct{ player.NopProvider }

func (playerProvider) Load(uuid.UUID, func(world.Dimension) *world.World) (player.Data, error) {
	return player.Data{
		GameMode:  visitor{},
		Health:    20,
		MaxHealth: 20,
		Hunger:    20,
	}, nil
}

type compass struct{ item.Compass }

func (compass) EncodeNBT() map[string]any {
	return map[string]any{
		"minecraft:item_lock": byte(0x1), // 0x1 locks in slot, 0x2 locks in inventory
	}
}

func (c compass) DecodeNBT(map[string]any) any {
	return c
}

//go:linkname sessionOf github.com/df-mc/dragonfly/server/player.(*Player).session
func sessionOf(*player.Player) *session.Session

//go:linkname writePacket github.com/df-mc/dragonfly/server/session.(*Session).writePacket
func writePacket(*session.Session, packet.Packet)
