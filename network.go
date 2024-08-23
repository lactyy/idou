package main

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sirupsen/logrus"
	"log"
	"math/rand"
	"strconv"
)

func (cfg config) listenerFunc(c server.Config) (server.Listener, error) {
	listen := minecraft.ListenConfig{
		MaximumPlayers:         c.MaxPlayers,
		StatusProvider:         minecraft.NewStatusProvider(c.Name, "idou"),
		AuthenticationDisabled: c.AuthDisabled,
		ResourcePacks:          c.Resources,
		Biomes:                 biomes(),
		TexturePacksRequired:   true,
	}
	if l, ok := c.Log.(*logrus.Logger); ok {
		listen.ErrorLog = log.Default()
		log.SetOutput(l.WithField("src", "gophertunnel").WriterLevel(logrus.DebugLevel))
	}
	if cfg.Listener.NetworkID == 0 {
		cfg.Listener.NetworkID = rand.Uint64()
	}
	l, err := listen.Listen("nethernet", strconv.FormatUint(cfg.Listener.NetworkID, 10))
	if err != nil {
		return nil, err
	}
	c.Log.Infof("Server listening on %v.", l.Addr())
	return listener{l}, nil
}

// listener is a Listener implementation that wraps around a minecraft.Listener so that it can be listened on by
// Server.
type listener struct {
	*minecraft.Listener
}

// Accept blocks until the next connection is established and returns it. An error is returned if the Listener was
// closed using Close.
func (l listener) Accept() (session.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return conn.(session.Conn), err
}

// Disconnect disconnects a connection from the Listener with a reason.
func (l listener) Disconnect(conn session.Conn, reason string) error {
	return l.Listener.Disconnect(conn.(*minecraft.Conn), reason)
}
