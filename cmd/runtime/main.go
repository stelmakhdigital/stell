package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/budaev/stell/pkg/config"
	"github.com/budaev/stell/runtime/api"
	"github.com/budaev/stell/runtime/executor"
	"github.com/budaev/stell/runtime/sandbox"
)

func main() {
	cfgPath := flag.String("config", "configs/runtime.yaml", "runtime config path")
	addr := flag.String("addr", "", "listen address override")
	flag.Parse()

	cfg, err := config.LoadRuntime(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	listen := cfg.Addr
	if *addr != "" {
		listen = *addr
	}

	hmacKey := cfg.HMACKey
	if hmacKey == "" {
		hmacKey = os.Getenv("STELL_HMAC_KEY")
	}

	policy := sandbox.DefaultPolicy()
	if cfg.Production {
		policy = sandbox.ProductionPolicy()
		if hmacKey == "" {
			log.Fatal("production runtime requires hmac_key or STELL_HMAC_KEY")
		}
	}
	if cfg.Sandbox.Image != "" {
		policy.Image = cfg.Sandbox.Image
	}
	if cfg.Sandbox.Network != "" {
		policy.Network = cfg.Sandbox.Network
	}
	if cfg.Sandbox.Memory != "" {
		policy.Memory = cfg.Sandbox.Memory
	}
	if cfg.Sandbox.User != "" {
		policy.User = cfg.Sandbox.User
	}
	if cfg.Sandbox.PidsLimit != "" {
		policy.PidsLimit = cfg.Sandbox.PidsLimit
	}
	if cfg.Sandbox.ReadOnlyRoot {
		policy.ReadOnlyRoot = true
	}
	if cfg.Production && policy.Network != "none" {
		log.Fatal("production sandbox network must be none")
	}

	sb := sandbox.NewDocker(policy)
	exec := executor.New(sb)
	if cfg.MaxOutputBytes > 0 {
		exec.MaxOutput = cfg.MaxOutputBytes
	}

	srv := &api.Server{
		Exec:        exec,
		HMACKey:     []byte(hmacKey),
		RequireHMAC: cfg.RequireHMAC || cfg.Production || hmacKey != "",
	}
	fmt.Printf("runtime listening on %s\n", listen)
	log.Fatal(http.ListenAndServe(listen, srv.Handler()))
}
