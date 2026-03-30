package handler

import (
	"context"

	"brewfather-mcp/internal/client"
	"brewfather-mcp/internal/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListRecipesInput struct {
	Limit            int    `json:"limit,omitempty" jsonschema:"Number of results to return (1-50). Defaults to 10"`
	StartAfter       string `json:"start_after,omitempty" jsonschema:"The _id of the last item from the previous page for pagination"`
	Include          string `json:"include,omitempty" jsonschema:"Comma-separated list of additional fields to include (e.g. fermentation or mash)"`
	Complete         bool   `json:"complete,omitempty" jsonschema:"If true returns all fields for each recipe. Defaults to false"`
	OrderBy          string `json:"order_by,omitempty" jsonschema:"Field to order results by. Defaults to _id"`
	OrderByDirection string `json:"order_by_direction,omitempty" jsonschema:"Sort direction: asc or desc. Defaults to asc"`
}

type GetRecipeInput struct {
	ID      string `json:"id" jsonschema:"The unique _id of the recipe to retrieve"`
	Include string `json:"include,omitempty" jsonschema:"Comma-separated list of additional fields to include (e.g. fermentation or mash). When omitted all fields are included"`
}

func RegisterRecipeTools(server *mcp.Server, svc *service.RecipeService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_recipes",
		Description: "List beer recipes from Brewfather. Returns recipe names, IDs, types, authors, equipment, and style names. Supports pagination via start_after.",
	}, newListRecipesHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_recipe",
		Description: "Get full details of a specific recipe by its ID. Returns style guidelines, volumes, gravity/IBU/color stats, all ingredients (fermentables, hops, yeast, misc), mash profile, fermentation profile, equipment, and notes.",
	}, newGetRecipeHandler(svc))
}

func newListRecipesHandler(svc *service.RecipeService) mcp.ToolHandlerFor[ListRecipesInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListRecipesInput) (*mcp.CallToolResult, any, error) {
		params := &client.ListRecipesParams{
			Limit:            input.Limit,
			StartAfter:       input.StartAfter,
			Include:          input.Include,
			Complete:         input.Complete,
			OrderBy:          input.OrderBy,
			OrderByDirection: input.OrderByDirection,
		}
		text, err := svc.ListRecipes(ctx, params)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newGetRecipeHandler(svc *service.RecipeService) mcp.ToolHandlerFor[GetRecipeInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetRecipeInput) (*mcp.CallToolResult, any, error) {
		params := &client.GetItemParams{Include: input.Include}
		text, err := svc.GetRecipe(ctx, input.ID, params)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}
