package main

import (
	"context"
	"fmt"
	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	bookv1 "github.com/strideynet/service-service/proto/book/v1"
	"google.golang.org/grpc"
	"log/slog"
	"os"
)

func main() {
	err := tryMain()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func tryMain() error {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		return fmt.Errorf("creating workloadapi src: %w", err)
	}

	conn, err := grpc.Dial(
		"127.0.0.1:1338",
		grpc.WithTransportCredentials(
			grpccredentials.MTLSClientCredentials(source, source, tlsconfig.AuthorizeAny()),
		),
	)
	if err != nil {
		return fmt.Errorf("dialling server: %w", err)
	}
	defer conn.Close()

	bookv1c := bookv1.NewBookServiceClient(conn)
	res, err := bookv1c.ListBooks(ctx, &bookv1.ListBooksRequest{})
	if err != nil {
		return fmt.Errorf("listing books: %w", err)
	}
	log.Info("listed books", slog.Any("books", res.Books))

	return nil
}
