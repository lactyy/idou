package main

import (
	"context"
	rand2 "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	atomic2 "github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/google/uuid"
	"github.com/lactyy/idou/internal"
	"github.com/lactyy/idou/internal/conf"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/franchise"
	"github.com/sandertv/gophertunnel/minecraft/franchise/signaling"
	"github.com/sandertv/gophertunnel/minecraft/nethernet"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/room"
	"github.com/sandertv/gophertunnel/xsapi"
	"github.com/sandertv/gophertunnel/xsapi/mpsd"
	"github.com/sandertv/gophertunnel/xsapi/xal"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"golang.org/x/text/message"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
	_ "unsafe"
)

func main() {
	log := logrus.New()
	log.Level = logrus.DebugLevel

	// var setting webrtc.SettingEngine
	// factory := logging.NewDefaultLoggerFactory()
	// factory.DefaultLogLevel = logging.LogLevelDebug
	// factory.Writer = log.WriterLevel(logrus.DebugLevel)
	// setting.LoggerFactory = factory
	// nethernet.DefaultDialer.API = webrtc.NewAPI(webrtc.WithSettingEngine(setting))

	// May be we can make the default value a function, that is only called
	// when the file does not exist.
	cfg, err := conf.Read(conf.TOMLEncoding, "config", defaultConfig())
	if err != nil {
		log.Fatalf("error reading config: %s", err)
	}

	var profiles internal.LoadStorer[uuid.UUID, profile]
	{
		db, err := leveldb.OpenFile(cfg.Database.Folder, nil)
		if err != nil {
			log.Fatalf("error opening database: %s", err)
		}
		defer db.Close()

		log.Infof("Database opened at %q", cfg.Database.Folder)
		profiles = internal.LevelDBProvider[uuid.UUID, profile]{
			DB: db,
		}
	}

	cred, err := conf.Read(conf.JSONEncoding, "auth.cred", &credentials{
		src: auth.TokenSource,
	})
	if err != nil {
		log.Fatalf("error reading credentials: %s", err)
	}
	src := auth.RefreshTokenSource(cred.Token)

	c, err := cfg.Server.Config(log)
	if err != nil {
		log.Fatalf("error creating config: %s", err)
	}

	{
		log.Infof("Publishing session...")
		discovery, err := franchise.Discover(protocol.CurrentVersion)
		if err != nil {
			log.Fatalf("error retrieving discovery: %s", err)
		}
		a := new(franchise.AuthorizationEnvironment)
		if err := discovery.Environment(a, franchise.EnvironmentTypeProduction); err != nil {
			log.Fatalf("error reading environment for authorization: %s", err)
		}
		s := new(signaling.Environment)
		if err := discovery.Environment(s, franchise.EnvironmentTypeProduction); err != nil {
			log.Fatalf("error reading environment for signaling: %s", err)
		}

		refresh, cancel := context.WithCancel(context.Background())
		defer cancel()
		prov := franchise.PlayFabXBLIdentityProvider{
			Environment: a,
			TokenSource: xal.RefreshTokenSourceContext(refresh, src, "http://playfab.xboxlive.com/"),
		}

		d := signaling.Dialer{
			NetworkID: rand.Uint64(),
		}

		dial, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		conn, err := d.DialContext(dial, prov, s)
		if err != nil {
			log.Fatalf("error dialing signaling: %s", err)
		}
		defer conn.Close()

		// A token source that refreshes a token used for generic Xbox Live services.
		x := xal.RefreshTokenSourceContext(refresh, src, "http://xboxlive.com")
		xt, err := x.Token()
		if err != nil {
			log.Fatalf("error refreshing xbox live token: %s", err)
		}
		claimer, ok := xt.(xsapi.DisplayClaimer)
		if !ok {
			log.Fatalf("xbox live token %T does not implement xsapi.DisplayClaimer", xt)
		}
		displayClaims := claimer.DisplayClaims()

		// The name of the session being published. This seems always to be generated
		// randomly, referenced as "GUID" of the session.
		name := strings.ToUpper(uuid.NewString())

		levelID := make([]byte, 8)
		_, _ = rand2.Read(levelID)

		custom, err := json.Marshal(room.Status{
			Joinability: room.JoinabilityJoinableByFriends,
			HostName:    displayClaims.GamerTag,
			OwnerID:     displayClaims.XUID,
			RakNetGUID:  "",
			// This is displayed as the suffix of the world name.
			Version:   protocol.CurrentVersion,
			LevelID:   base64.StdEncoding.EncodeToString(levelID),
			WorldName: c.Name,
			WorldType: room.WorldTypeCreative,
			// The game seems checking this field before joining a session, causes
			// RequestNetworkSettings packet not being even sent to the remote host.
			Protocol:                protocol.CurrentProtocol,
			MemberCount:             1,
			MaxMemberCount:          8,
			BroadcastSetting:        room.BroadcastSettingFriendsOfFriends,
			LanGame:                 true,
			IsEditorWorld:           false,
			TransportLayer:          2,
			WebRTCNetworkID:         d.NetworkID,
			OnlineCrossPlatformGame: true,
			CrossPlayDisabled:       false,
			TitleID:                 0,
			SupportedConnections: []room.Connection{
				{
					ConnectionType:  3, // WebSocketsWebRTCSignaling
					HostIPAddress:   "",
					HostPort:        0,
					NetherNetID:     d.NetworkID,
					WebRTCNetworkID: d.NetworkID,
					RakNetGUID:      "UNASSIGNED_RAKNET_GUID",
				},
			},
		})
		if err != nil {
			log.Fatalf("error encoding custom properties: %s", err)
		}
		cfg := mpsd.PublishConfig{
			Description: &mpsd.SessionDescription{
				Properties: &mpsd.SessionProperties{
					System: &mpsd.SessionPropertiesSystem{
						JoinRestriction: mpsd.SessionRestrictionFollowed,
						ReadRestriction: mpsd.SessionRestrictionFollowed,
					},
					Custom: custom,
				},
			},
		}

		publish, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		session, err := cfg.PublishContext(publish, x, mpsd.SessionReference{
			ServiceConfigID: uuid.MustParse("4fc10100-5f7a-4470-899b-280835760c07"),
			TemplateName:    "MinecraftLobby",
			Name:            name,
		})
		if err != nil {
			log.Fatalf("error publishing session: %s", err)
		}
		defer session.Close()

		log.Debugf("Session Name: %q", name)
		log.Debugf("Network ID: %d", d.NetworkID)

		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))

		minecraft.RegisterNetwork("nethernet", &nethernet.Network{
			Signaling: conn,
		})
	}

	c.WorldProvider = world.NopProvider{}
	c.Generator = func(world.Dimension) world.Generator {
		return world.NopGenerator{}
	}
	c.PlayerProvider = playerProvider{}
	c.Listeners = []func(server.Config) (server.Listener, error){
		cfg.listenerFunc,
	}

	srv := c.New()
	srv.CloseOnProgramEnd()

	srv.World().SetBlock(cube.Pos{0, -2, 0}, block.Barrier{}, nil)
	srv.World().StopWeatherCycle()
	srv.World().StopTime()

	srv.Listen()
	for srv.Accept(func(p *player.Player) {
		plog := log.WithFields(logrus.Fields{
			"uuid":    p.UUID(),
			"address": p.Addr(),
			"name":    p.Name(),
		})

		// TODO
		writePacket(sessionOf(p), &packet.GameRulesChanged{
			GameRules: []protocol.GameRule{
				{
					Name:  "showTags",
					Value: false,
				},
			},
		})

		p.SetImmobile()
		for _, other := range srv.Players() {
			other.HideEntity(p)
			p.HideEntity(other)
		}

		prof, err := internal.LoadDefault(profiles, p.UUID(), profile{
			UUID: p.UUID(),
		})
		if err != nil {
			plog.Errorf("error loading profile: %s", err)
			p.Disconnect("Sorry, We couldn't load your profile")
			return
		}
		if prof.Language == language.Und {
			locale := p.Locale()
			base, _ := locale.Base()
			for _, tag := range message.DefaultCatalog.Languages() {
				if b, _ := tag.Base(); base == b {
					prof.Language = tag
					p.Message(message.NewPrinter(tag).Sprintf("language.auto", display.Self.Name(tag)))
					goto FOUND
				}
			}
			prof.Language = language.English
		FOUND:
		}

		m := message.NewPrinter(prof.Language)
		a := new(atomic.Pointer[message.Printer])
		a.Store(m)

		h := &handler{
			p:   p,
			log: plog.Logger,

			m: a,

			prof:   atomic2.NewValue(prof),
			storer: profiles,
		}
		handlers.Store(p.UUID(), h)
		p.Handle(h)

		p.SendForm(newServerForm(m, h.prof.Load()))
		h.setInventory(p, m)
	}) {
	}
}

type profile struct {
	UUID     uuid.UUID       `json:"uuid"`
	Language language.Tag    `json:"language"`
	Servers  []serverProfile `json:"servers"`
}

type serverProfile struct {
	Name       string       `json:"name,omitempty"`
	Address    *net.UDPAddr `json:"address"`
	RawAddress string       `json:"rawAddress"`
}

func serverProfileName(s string) string {
	if len(s) > 20 {
		s = ""
	}
	return s
}

func addressEqual(a, b *net.UDPAddr) bool {
	if a.String() == b.String() {
		return true
	}
	if a.IP.Equal(b.IP) && a.Port == b.Port {
		return true
	}
	return false
}

func (prof profile) shouldRemember(addr *net.UDPAddr, rawAddr string) bool {
	for _, srv := range prof.Servers {
		if addressEqual(srv.Address, addr) || rawAddr == srv.RawAddress {
			return false
		}
	}
	return true
}

func defaultConfig() (cfg config) {
	cfg.Database.Folder = "db"

	cfg.Server = server.DefaultConfig()
	cfg.Server.World.SaveData = false
	cfg.Server.Players.SaveData = false
	cfg.Server.Server.JoinMessage = ""
	cfg.Server.Server.QuitMessage = ""

	cfg.Server.Server.Name = "Testing"
	return cfg
}

type config struct {
	Listener struct {
		NetworkID uint64 `json:"networkID" toml:"networkID"`
	} `json:"listener" toml:"listener"`
	Database struct {
		Folder string `json:"folder" toml:"folder"`
	} `json:"db" toml:"db"`
	Server server.UserConfig `json:"server" toml:"server"`
}

//go:linkname biomes github.com/df-mc/dragonfly/server.biomes
func biomes() map[string]any

type credentials struct {
	*oauth2.Token `json:"token"`
	src           oauth2.TokenSource
}

func (cred *credentials) MarshalJSON() ([]byte, error) {
	if cred.Token == nil {
		t, err := cred.src.Token()
		if err != nil {
			return nil, fmt.Errorf("request token: %w", err)
		}
		cred.Token = t
	}
	return json.Marshal(*cred)
}
