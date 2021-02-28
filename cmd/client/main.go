package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/rsocket/rsocket-go/logger"
	"github.com/rsocket/rsocket-go/payload"
)

func makeTransport() transport.ClientTransporter {
	return rsocket.WebsocketClient().SetURL("ws://127.0.0.1:7878/events").Build()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	root := zerolog.New(os.Stdout).With().Caller().Logger()

	onClose := func(err error) {
		cancel()
		root.Error().Err(err).Msg("callback")
		os.Exit(1)
	}

	logger.SetLogger(nil)

	cli, err := rsocket.Connect().
		OnClose(onClose).
		ConnectTimeout(time.Second).
		SetupPayload(payload.NewString("data", "metadata")). // must be equal fold setup
		Transport(makeTransport()).
		Start(ctx)
	if err != nil {
		root.Error().Err(err).Msg("connect to server")
		return
	}
	defer func() {
		if err := cli.Close(); err != nil {
			root.Error().Err(err).Msg("close client")
		} else {
			root.Debug().Msg("close ok")
		}
	}()

	// time.Sleep(time.Second * 2)

	select {
	case <-ctx.Done():
		return
	default:
		root.Debug().Msg("ctx must be cancelled")
	}

	response, err := cli.RequestResponse(payload.NewString("data", "metadata")).Block(ctx)
	if err != nil {
		root.Error().Err(err).Msg("do request")
		return
	}

	if response == nil {
		root.Debug().Msg("empty response")
		return
	}

	meta, _ := response.MetadataUTF8()

	root.Debug().Str("data", response.DataUTF8()).Str("metadata", meta).Msg("response")
}
