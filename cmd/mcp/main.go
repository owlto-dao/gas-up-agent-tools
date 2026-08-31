package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/owlto-dao/gas-up-agent-tools/internal/mcpserver"
)

func main() {
	host := env("MCP_HOST", "0.0.0.0")
	port := env("MCP_PORT", "4010")

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpserver.NewHTTPHandler(mcpserver.OptionsFromEnv()))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	addr := host + ":" + port
	log.Printf("gas_up_mcp_listen addr=%s mcp_path=/mcp backend=%s", addr, env("GAS_UP_API_BASE_URL", "http://localhost:4000"))
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
