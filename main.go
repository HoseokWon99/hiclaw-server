package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hiclaw-server/config"
	"hiclaw-server/core"
	"hiclaw-server/hub"
	"hiclaw-server/server"
	"hiclaw-server/service"
	"hiclaw-server/storage"
	"hiclaw-server/tailnet"
)

func main() {
	cfg := config.Load()

	// Storage layer (implements core interfaces)
	deviceStore, err := storage.NewSQLiteDeviceStore(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("device store: %v", err)
	}
	defer deviceStore.Close()

	// Chat store: use in-memory by default
	// TODO: swap to CouchBase when gocb/v2 dependency is resolved
	var chatStore core.ChatStore
	chatStore = storage.NewMemoryChatStore()
	log.Println("using in-memory chat store")

	// Infrastructure
	h := hub.New()

	// Service layer (implements core interfaces, depends on core store interfaces)
	deviceSvc := service.NewDeviceService(deviceStore)
	agentSvc := service.NewAgentService("http://"+cfg.AgentDeviceName, cfg.AgentWebhookPath)
	chatSvc := service.NewChatService(chatStore, h, agentSvc)

	// Tailscale discovery (depends on core.DeviceService)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	disc := tailnet.NewDiscovery(deviceSvc, cfg.AgentDeviceName)
	if err := disc.SyncOnce(ctx); err != nil {
		log.Printf("initial tailnet sync (continuing): %v", err)
	}
	go disc.RunLoop(ctx, 30*time.Second)

	// HTTP layer (depends on core.ChatService)
	handler := server.NewHandler(chatSvc, h)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/messages", handler.GetMessages)
	mux.HandleFunc("POST /api/messages", handler.PostMessage)
	mux.HandleFunc("POST /api/agent/reply", handler.HandleAgentReply)
	mux.HandleFunc("GET /ws", handler.HandleWebSocket)

	log.Printf("hiclaw-server starting on %s", cfg.HTTPAddr)
	go func() {
		if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
}
