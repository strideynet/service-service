package main

import (
	"context"
	"fmt"
	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/strideynet/service-service/proto/book/v1"
	"google.golang.org/grpc"
	"log/slog"
	"net"
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

	bookSvc := &BookService{
		log: log,
	}
	srv := grpc.NewServer(
		grpc.Creds(
			grpccredentials.MTLSServerCredentials(
				source,
				source,
				// AuthorizeAny is used here so that more advanced authz
				// decisions can be made at the RPC level.
				tlsconfig.AuthorizeAny(),
			),
		),
	)
	bookv1.RegisterBookServiceServer(srv, bookSvc)

	lis, err := net.Listen("tcp", ":1338")
	if err != nil {
		return fmt.Errorf("creating listener: %w", err)
	}

	return srv.Serve(lis)
}

type BookService struct {
	log *slog.Logger
	bookv1.UnimplementedBookServiceServer
}

func (b *BookService) ListBooks(ctx context.Context, req *bookv1.ListBooksRequest) (*bookv1.ListBooksResponse, error) {
	id, _ := grpccredentials.PeerIDFromContext(ctx)
	b.log.Info("ListBooks", slog.String("id", id.String()))

	return &bookv1.ListBooksResponse{
		Books: []*bookv1.Book{
			{
				Isbn:  "9781098131890",
				Title: "Identity-Native Infrastructure Access Management",
			},
		},
	}, nil
}

func (b *BookService) DeleteBook(ctx context.Context, req *bookv1.DeleteBookRequest) (*bookv1.DeleteBookResponse, error) {
	id, _ := grpccredentials.PeerIDFromContext(ctx)
	b.log.Info(
		"DeleteBook",
		slog.String("id", id.String()),
		slog.String("id.path", id.Path()),
	)
	if id.Path() != "admin" {
		return nil, fmt.Errorf("not authorized")
	}
	return &bookv1.DeleteBookResponse{}, nil
}
