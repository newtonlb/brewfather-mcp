package handler

import (
	"context"

	"brewfather-mcp/internal/client"
	"brewfather-mcp/internal/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListInventoryInput struct {
	InventoryExists  *bool  `json:"inventory_exists,omitempty" jsonschema:"If true only include items with stock greater than zero"`
	InventoryNegative *bool  `json:"inventory_negative,omitempty" jsonschema:"If true only include items with negative stock (owed/overdrawn)"`
	Limit            int    `json:"limit,omitempty" jsonschema:"Number of results to return (1-50). Defaults to 10"`
	StartAfter       string `json:"start_after,omitempty" jsonschema:"The _id of the last item from the previous page for pagination"`
	Include          string `json:"include,omitempty" jsonschema:"Comma-separated list of additional fields to include"`
	Complete         bool   `json:"complete,omitempty" jsonschema:"If true returns all fields for each item. Defaults to false"`
	OrderBy          string `json:"order_by,omitempty" jsonschema:"Field to order results by. Defaults to _id"`
	OrderByDirection string `json:"order_by_direction,omitempty" jsonschema:"Sort direction: asc or desc. Defaults to asc"`
}

type GetInventoryItemInput struct {
	ID      string `json:"id" jsonschema:"The unique _id of the inventory item to retrieve"`
	Include string `json:"include,omitempty" jsonschema:"Comma-separated list of additional fields to include. When omitted all fields are included"`
}

type UpdateInventoryInput struct {
	ID              string   `json:"id" jsonschema:"The unique _id of the inventory item to update"`
	Inventory       *float64 `json:"inventory,omitempty" jsonschema:"Set the inventory amount to this absolute value. Takes precedence over inventory_adjust if both are provided"`
	InventoryAdjust *float64 `json:"inventory_adjust,omitempty" jsonschema:"Adjust the current inventory by this amount (positive to add or negative to subtract)"`
}

func toListInventoryParams(input ListInventoryInput) *client.ListInventoryParams {
	return &client.ListInventoryParams{
		InventoryExists:  input.InventoryExists,
		InventoryNegative: input.InventoryNegative,
		Limit:            input.Limit,
		StartAfter:       input.StartAfter,
		Include:          input.Include,
		Complete:         input.Complete,
		OrderBy:          input.OrderBy,
		OrderByDirection: input.OrderByDirection,
	}
}

func toGetItemParams(include string) *client.GetItemParams {
	return &client.GetItemParams{Include: include}
}

func toUpdateInventoryBody(input UpdateInventoryInput) *client.UpdateInventoryBody {
	return &client.UpdateInventoryBody{
		Inventory:       input.Inventory,
		InventoryAdjust: input.InventoryAdjust,
	}
}

func RegisterInventoryTools(server *mcp.Server, svc *service.InventoryService) {
	// Fermentables
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_fermentables",
		Description: "List fermentable ingredients from inventory (malts, sugars, extracts, adjuncts). Shows name, type, supplier, and stock level. Use inventory_exists to filter to items in stock.",
	}, newListFermentablesHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_fermentable",
		Description: "Get full details of a specific fermentable ingredient by ID. Returns type, color, potential gravity, attenuation, origin, supplier, and current stock.",
	}, newGetFermentableHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_fermentable_inventory",
		Description: "Update the inventory amount of a specific fermentable. Use inventory to set an absolute value, or inventory_adjust to add/subtract from the current amount.",
	}, newUpdateFermentableHandler(svc))

	// Hops
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_hops",
		Description: "List hop varieties from inventory. Shows name, alpha acid, form (pellet/whole/cryo), use, and stock level in grams.",
	}, newListHopsHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_hop",
		Description: "Get full details of a specific hop variety by ID. Returns alpha/beta acids, oil profile, origin, usage role, and current stock.",
	}, newGetHopHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_hop_inventory",
		Description: "Update the inventory amount of a specific hop. Use inventory to set an absolute value in grams, or inventory_adjust to add/subtract.",
	}, newUpdateHopHandler(svc))

	// Miscs
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_miscs",
		Description: "List miscellaneous ingredients from inventory (spices, finings, water agents, herbs, flavors). Shows name, type, use, and stock level.",
	}, newListMiscsHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_misc",
		Description: "Get full details of a specific miscellaneous ingredient by ID. Returns type, use timing, unit, concentration, and current stock.",
	}, newGetMiscHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_misc_inventory",
		Description: "Update the inventory amount of a specific misc ingredient. Use inventory to set an absolute value, or inventory_adjust to add/subtract.",
	}, newUpdateMiscHandler(svc))

	// Yeasts
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_yeasts",
		Description: "List yeast strains from inventory. Shows name, type (ale/lager/hybrid), form (dry/liquid), attenuation, laboratory, and stock level.",
	}, newListYeastsHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_yeast",
		Description: "Get full details of a specific yeast strain by ID. Returns type, form, lab, attenuation range, temperature range, flocculation, ABV tolerance, best-for styles, and current stock.",
	}, newGetYeastHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_yeast_inventory",
		Description: "Update the inventory amount of a specific yeast. Use inventory to set an absolute value, or inventory_adjust to add/subtract.",
	}, newUpdateYeastHandler(svc))
}

// --- Fermentable handlers ---

func newListFermentablesHandler(svc *service.InventoryService) mcp.ToolHandlerFor[ListInventoryInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListInventoryInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.ListFermentables(ctx, toListInventoryParams(input))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newGetFermentableHandler(svc *service.InventoryService) mcp.ToolHandlerFor[GetInventoryItemInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetInventoryItemInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.GetFermentable(ctx, input.ID, toGetItemParams(input.Include))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newUpdateFermentableHandler(svc *service.InventoryService) mcp.ToolHandlerFor[UpdateInventoryInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateInventoryInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.UpdateFermentableInventory(ctx, input.ID, toUpdateInventoryBody(input))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

// --- Hop handlers ---

func newListHopsHandler(svc *service.InventoryService) mcp.ToolHandlerFor[ListInventoryInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListInventoryInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.ListHops(ctx, toListInventoryParams(input))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newGetHopHandler(svc *service.InventoryService) mcp.ToolHandlerFor[GetInventoryItemInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetInventoryItemInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.GetHop(ctx, input.ID, toGetItemParams(input.Include))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newUpdateHopHandler(svc *service.InventoryService) mcp.ToolHandlerFor[UpdateInventoryInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateInventoryInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.UpdateHopInventory(ctx, input.ID, toUpdateInventoryBody(input))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

// --- Misc handlers ---

func newListMiscsHandler(svc *service.InventoryService) mcp.ToolHandlerFor[ListInventoryInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListInventoryInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.ListMiscs(ctx, toListInventoryParams(input))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newGetMiscHandler(svc *service.InventoryService) mcp.ToolHandlerFor[GetInventoryItemInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetInventoryItemInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.GetMisc(ctx, input.ID, toGetItemParams(input.Include))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newUpdateMiscHandler(svc *service.InventoryService) mcp.ToolHandlerFor[UpdateInventoryInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateInventoryInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.UpdateMiscInventory(ctx, input.ID, toUpdateInventoryBody(input))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

// --- Yeast handlers ---

func newListYeastsHandler(svc *service.InventoryService) mcp.ToolHandlerFor[ListInventoryInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListInventoryInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.ListYeasts(ctx, toListInventoryParams(input))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newGetYeastHandler(svc *service.InventoryService) mcp.ToolHandlerFor[GetInventoryItemInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetInventoryItemInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.GetYeast(ctx, input.ID, toGetItemParams(input.Include))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newUpdateYeastHandler(svc *service.InventoryService) mcp.ToolHandlerFor[UpdateInventoryInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateInventoryInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.UpdateYeastInventory(ctx, input.ID, toUpdateInventoryBody(input))
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}
