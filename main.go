package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"brewfather-mcp/internal/client"
	"brewfather-mcp/internal/handler"
	"brewfather-mcp/internal/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	userID := os.Getenv("BREWFATHER_USER_ID")
	apiKey := os.Getenv("BREWFATHER_API_KEY")

	if userID == "" || apiKey == "" {
		fmt.Fprintln(os.Stderr, "BREWFATHER_USER_ID and BREWFATHER_API_KEY environment variables are required")
		os.Exit(1)
	}

	c := client.NewClient(userID, apiKey)

	batchSvc := service.NewBatchService(c)
	recipeSvc := service.NewRecipeService(c)
	inventorySvc := service.NewInventoryService(c)

	server := mcp.NewServer(
		&mcp.Implementation{Name: "brewfather", Version: "v1.0.0"},
		nil,
	)

	handler.RegisterBatchTools(server, batchSvc)
	handler.RegisterRecipeTools(server, recipeSvc)
	handler.RegisterInventoryTools(server, inventorySvc)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
