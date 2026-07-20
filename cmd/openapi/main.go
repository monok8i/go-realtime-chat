package main

import (
	"os"

	"go-realtime-chat/internal/api"
	"go-realtime-chat/internal/api/handlers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()
	humaApi := humachi.New(router, huma.DefaultConfig("Go Realtime Chat", "1.0.0"))

	chatHandler := handlers.NewChatHandler(nil)
	api.RegisterRoutes(humaApi, router, chatHandler)

	spec, err := humaApi.OpenAPI().YAML()
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("openapi.yaml", spec, 0600); err != nil {
		panic(err)
	}
}
