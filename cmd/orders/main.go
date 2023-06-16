package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"pmozhchil/orders/internal/app/repository"

	gw "pmozhchil/orders/api/pmozhchil/orders"
	"pmozhchil/orders/internal/app/orders"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Config struct {
	GRPC  string
	HTTP  string
	DbDSN string
}

var (
	Version = ""
	Cfg     = Config{
		GRPC:  "localhost:8070",
		HTTP:  "localhost:8080",
		DbDSN: "host=127.0.0.1 port=5432 user=test password=test database=test sslmode=disable",
	}
)

const MaxReceiveMessageSize = 50000000

func main() {
	log := logrus.New()
	//log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter) //default
	//log.Formatter.(*logrus.TextFormatter).DisableColors = true    // remove colors
	log.Formatter.(*logrus.TextFormatter).DisableTimestamp = true // remove timestamp from test output
	log.Level = logrus.TraceLevel
	log.Out = os.Stdout
	//flag.Parse()
	//defer glog.Flush()

	if err := run(log); err != nil {
		log.Fatal(err)
	}
}

func run(logger *logrus.Logger) error { // nolint: funlen
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var group errgroup.Group

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", Cfg.GRPC)
	if err != nil {
		logger.Fatal("Failed to listen", err)
	}

	// Logrus entry is used, allowing pre-definition of certain fields by the user.
	logrusEntry := logrus.NewEntry(logger)
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}

	// Create a gRPC server object
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_logrus.UnaryServerInterceptor(logrusEntry, opts...),
		),
	)

	logger.Debugf("Connect to Db on %s", Cfg.DbDSN)
	orm, err := repository.NewORM(Cfg.DbDSN)
	if err != nil {
		logger.Fatal("cannot init db")
	}

	app, err := orders.NewOrdersService(Version, orm)
	if err != nil {
		logger.Fatal("cannot init app")
	}

	gw.RegisterOrdersServiceServer(s, app)

	// Serve gRPC server
	logger.Debugf("Serving gRPC on %s", Cfg.GRPC)
	group.Go(func() error {
		return s.Serve(lis)
	})

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		ctx,
		Cfg.GRPC,
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxReceiveMessageSize)),
	)
	if err != nil {
		logger.Fatal("Failed to dial server", err)
	}

	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{OrigName: true, EmitDefaults: true},
		),
	)

	// Register Orders service
	err = gw.RegisterOrdersServiceHandler(ctx, gwMux, conn)
	if err != nil {
		return err
	}

	gwServer := &http.Server{
		Addr:    Cfg.HTTP,
		Handler: gwMux,
	}

	logger.Debugf("Serving gRPC-Gateway on http://%s", Cfg.HTTP)
	group.Go(func() error {
		return gwServer.ListenAndServe()
	})

	return group.Wait()
}
