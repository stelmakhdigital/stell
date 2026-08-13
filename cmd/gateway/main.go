package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	pubapi "github.com/budaev/agent/internal/api"
	"github.com/budaev/agent/internal/bootstrap"
)

func main() {
	cfgPath := flag.String("config", "configs/agent.yaml", "agent config")
	addr := flag.String("addr", "", "listen address override")
	runtimeMode := flag.String("runtime", "local", "hands mode: local|http")
	runtimeURL := flag.String("runtime-url", "", "hands URL for http mode")
	workspace := flag.String("workspace", "", "workspace root")
	flag.Parse()

	rt, err := bootstrap.New(bootstrap.Options{
		ConfigPath:  *cfgPath,
		RuntimeMode: *runtimeMode,
		RuntimeURL:  *runtimeURL,
		Workspace:   *workspace,
	})
	if err != nil {
		log.Fatal(err)
	}
	token := rt.Config.Agent.APIToken
	if token == "" {
		token = os.Getenv("AGENT_API_TOKEN")
	}
	if token == "" {
		log.Fatal("set agent.api_token or AGENT_API_TOKEN")
	}
	listen := rt.Config.Agent.APIAddr
	if *addr != "" {
		listen = *addr
	}
	if listen == "" {
		listen = "127.0.0.1:8080"
	}
	fmt.Printf("gateway listening on %s\n", listen)
	log.Fatal(http.ListenAndServe(listen, pubapi.New(token, rt.Spawn).Handler()))
}
