package main

import (
	"context"
	"io"
	"os"
	"os/signal"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/dnnrly/gsl-lang/lsp"
)

type readWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	rwc := readWriteCloser{os.Stdin, os.Stdout, os.Stdin}
	stream := jsonrpc2.NewStream(rwc)
	conn := jsonrpc2.NewConn(stream)
	defer conn.Close()

	client := protocol.ClientDispatcher(conn)

	server := lsp.NewServer(client)

	handler := protocol.ServerHandler(server, func(ctx context.Context, req *jsonrpc2.Request) (any, error) {
		return nil, jsonrpc2.ErrNotHandled
	})

	conn.Go(ctx, handler)

	<-conn.Done()
}
