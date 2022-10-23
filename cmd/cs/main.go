package main

import (
	"flag"
	"log"
	"net"

	"github.com/nyiyui/qrystal/cs"
	"github.com/nyiyui/qrystal/node/api"
	"github.com/nyiyui/qrystal/util"
	"github.com/tidwall/buntdb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "config", "", "config file path")
	flag.Parse()

	util.L, _ = zap.NewDevelopment()
	defer util.L.Sync()
	util.S = util.L.Sugar()

	util.ShowCurrent()

	config, err := cs.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("load config: %s", err)
	}
	log.Print("config file loaded")
	creds, err := credentials.NewServerTLSFromFile(config.TLSCertPath, config.TLSKeyPath)
	if err != nil {
		log.Fatalf("server tls: %s", err)
	}
	log.Printf("TLS creds read")

	db, err := buntdb.Open(config.DBPath)
	if err != nil {
		log.Fatalf("open db: %s", err)
	}

	server, err := cs.New(*config.CC, config.BackportPath, db)
	if err != nil {
		log.Fatalf("new: %s", err)
	}
	err = server.AddTokens(config.Tokens.Raw)
	if err != nil {
		log.Fatalf("add tokens: %s", err)
	}
	if config.BackportPath != "" && false {
		err = server.ReadBackport()
		if err != nil {
			log.Fatalf("read backport: %s", err)
		}
		log.Printf("read backport from %s", config.BackportPath)
	}
	gs := grpc.NewServer(grpc.Creds(creds))
	api.RegisterCentralSourceServer(gs, server)
	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		log.Fatalf("listen: %s", err)
	}
	log.Print("will serve…")
	err = gs.Serve(lis)
	if err != nil {
		log.Fatalf("serve: %s", err)
	}
}
