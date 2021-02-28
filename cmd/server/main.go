package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	ws "github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := zerolog.New(os.Stdout)

	if err := rsocket.Receive().
		OnStart(onStart(logger)).
		Acceptor(makeAcceptor(logger, makeServer(logger))).
		Transport(makeTransport()).
		Serve(ctx); err != nil {
		logger.Fatal().Err(err).Send()
	}
}

func makeTransport() transport.ServerTransporter {
	return rsocket.WebsocketServer().SetAddr("127.0.0.1:7878").SetPath("/events").SetUpgrader(&ws.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Build()
}

func onStart(logger zerolog.Logger) func() {
	return func() {
		logger.Debug().Msg("start")
	}
}

func makeServer(logger zerolog.Logger) *server {
	return &server{logger: logger}
}

func makeAcceptor(logger zerolog.Logger, server rsocket.RSocket) rsocket.ServerAcceptor {
	return func(_ context.Context, setup payload.SetupPayload, sending rsocket.CloseableRSocket) (rsocket.RSocket, error) {
		data := setup.DataUTF8()

		logger.Debug().
			Str("version", setup.Version().String()).
			Str("data", data).
			Str("metadata_mime_type", setup.MetadataMimeType()).
			Msg("setup")

		sending.OnClose(func(err error) {
			logger.Error().Err(err).Msg("close")
		})

		if !strings.EqualFold(data, "setup") {
			logger.Debug().Msg("reject")
			return nil, errors.New("unknown message")
		}

		return server, nil
	}
}

type server struct {
	logger zerolog.Logger
}

func (s *server) FireAndForget(message payload.Payload) {
	s.logger.Debug().Str("method", "FireAndForget").Send()
}

func (s *server) MetadataPush(message payload.Payload) {
	s.logger.Debug().Str("method", "MetadataPush").Send()
}

func (s *server) RequestResponse(message payload.Payload) mono.Mono {
	data := message.DataUTF8()
	meta, _ := message.MetadataUTF8()

	s.logger.Debug().Str("method", "RequestResponse").
		Str("meta", meta).
		Str("data", data).
		Send()

	return mono.Just(payload.NewString("data", "meta"))
}

func (s *server) RequestStream(message payload.Payload) flux.Flux {
	s.logger.Debug().Str("method", "RequestStream").Send()

	return nil
}
func (s *server) RequestChannel(messages flux.Flux) flux.Flux {

	s.logger.Debug().Str("method", "RequestChannel").Send()

	return nil
}
