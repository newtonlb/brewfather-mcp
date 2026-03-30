package handler

import (
	"context"

	"brewfather-mcp/internal/client"
	"brewfather-mcp/internal/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListBatchesInput struct {
	Status           string `json:"status,omitempty" jsonschema:"Filter by batch status. One of: Planning, Brewing, Fermenting, Conditioning, Completed, Archived"`
	Limit            int    `json:"limit,omitempty" jsonschema:"Number of results to return (1-50). Defaults to 10"`
	StartAfter       string `json:"start_after,omitempty" jsonschema:"The _id of the last item from the previous page for pagination"`
	Include          string `json:"include,omitempty" jsonschema:"Comma-separated list of additional fields to include (e.g. recipe.fermentation or recipe.mash)"`
	Complete         bool   `json:"complete,omitempty" jsonschema:"If true returns all fields for each batch. Defaults to false"`
	OrderBy          string `json:"order_by,omitempty" jsonschema:"Field to order results by. Defaults to _id"`
	OrderByDirection string `json:"order_by_direction,omitempty" jsonschema:"Sort direction: asc or desc. Defaults to asc"`
}

type GetBatchInput struct {
	ID      string `json:"id" jsonschema:"The unique _id of the batch to retrieve"`
	Include string `json:"include,omitempty" jsonschema:"Comma-separated list of additional fields to include (e.g. recipe.fermentation or recipe.mash)"`
}

type UpdateBatchInput struct {
	ID                       string   `json:"id" jsonschema:"The unique _id of the batch to update"`
	Status                   *string  `json:"status,omitempty" jsonschema:"Set batch status. One of: Planning, Brewing, Fermenting, Conditioning, Completed, Archived"`
	MeasuredMashPh           *float64 `json:"measured_mash_ph,omitempty" jsonschema:"Mash pH (0-14)"`
	MeasuredBoilSize         *float64 `json:"measured_boil_size,omitempty" jsonschema:"Pre-Boil Volume in liters"`
	MeasuredFirstWortGravity *float64 `json:"measured_first_wort_gravity,omitempty" jsonschema:"Pre-Sparge Gravity in SG (e.g. 1.055)"`
	MeasuredPreBoilGravity   *float64 `json:"measured_pre_boil_gravity,omitempty" jsonschema:"Pre-Boil Gravity in SG (e.g. 1.055)"`
	MeasuredPostBoilGravity  *float64 `json:"measured_post_boil_gravity,omitempty" jsonschema:"Post-Boil Gravity in SG (e.g. 1.055)"`
	MeasuredKettleSize       *float64 `json:"measured_kettle_size,omitempty" jsonschema:"Post-Boil Volume in liters"`
	MeasuredOg               *float64 `json:"measured_og,omitempty" jsonschema:"Original Gravity in SG (e.g. 1.050)"`
	MeasuredFermenterTopUp   *float64 `json:"measured_fermenter_top_up,omitempty" jsonschema:"Fermenter Top-Up Volume in liters"`
	MeasuredBatchSize        *float64 `json:"measured_batch_size,omitempty" jsonschema:"Fermenter Volume in liters"`
	MeasuredFg               *float64 `json:"measured_fg,omitempty" jsonschema:"Final Gravity in SG (e.g. 1.011)"`
	MeasuredBottlingSize     *float64 `json:"measured_bottling_size,omitempty" jsonschema:"Final Bottling/Kegging Volume in liters"`
	CarbonationTemp          *float64 `json:"carbonation_temp,omitempty" jsonschema:"Carbonation Temperature in Celsius (-50 to 100)"`
}

type BatchIDInput struct {
	ID string `json:"id" jsonschema:"The unique _id of the batch"`
}

func RegisterBatchTools(server *mcp.Server, svc *service.BatchService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_batches",
		Description: "List brewing batches from Brewfather. Returns batch names, IDs, status, brew dates, and recipe names. Use status filter to find batches in a specific phase. Supports pagination via start_after.",
	}, newListBatchesHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_batch",
		Description: "Get full details of a specific brewing batch by its ID. Returns measured values, estimated values, recipe details, ingredients, dates, carbonation info, and notes.",
	}, newGetBatchHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_batch",
		Description: "Update a batch's status and/or measured values (gravity readings, volumes, pH, temperatures). All fields are optional - only send the values you want to change.",
	}, newUpdateBatchHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_batch_last_reading",
		Description: "Get the most recent sensor or manual reading for a batch. Returns gravity, temperature, device info, and timestamp. Useful for checking current fermentation progress.",
	}, newGetBatchLastReadingHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_batch_readings",
		Description: "Get all readings (sensor and manual) recorded for a batch. Returns a time series of gravity, temperature, and device data. Useful for analyzing fermentation trends.",
	}, newGetBatchReadingsHandler(svc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_batch_brew_tracker",
		Description: "Get the brew tracker state for a batch. Shows the current stage, steps, timers, and progress through the brew day process.",
	}, newGetBatchBrewTrackerHandler(svc))
}

func newListBatchesHandler(svc *service.BatchService) mcp.ToolHandlerFor[ListBatchesInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListBatchesInput) (*mcp.CallToolResult, any, error) {
		params := &client.ListBatchesParams{
			Status:           input.Status,
			Limit:            input.Limit,
			StartAfter:       input.StartAfter,
			Include:          input.Include,
			Complete:         input.Complete,
			OrderBy:          input.OrderBy,
			OrderByDirection: input.OrderByDirection,
		}
		text, err := svc.ListBatches(ctx, params)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newGetBatchHandler(svc *service.BatchService) mcp.ToolHandlerFor[GetBatchInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetBatchInput) (*mcp.CallToolResult, any, error) {
		params := &client.GetItemParams{Include: input.Include}
		text, err := svc.GetBatch(ctx, input.ID, params)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newUpdateBatchHandler(svc *service.BatchService) mcp.ToolHandlerFor[UpdateBatchInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateBatchInput) (*mcp.CallToolResult, any, error) {
		body := &client.UpdateBatchBody{
			Status:                   input.Status,
			MeasuredMashPh:           input.MeasuredMashPh,
			MeasuredBoilSize:         input.MeasuredBoilSize,
			MeasuredFirstWortGravity: input.MeasuredFirstWortGravity,
			MeasuredPreBoilGravity:   input.MeasuredPreBoilGravity,
			MeasuredPostBoilGravity:  input.MeasuredPostBoilGravity,
			MeasuredKettleSize:       input.MeasuredKettleSize,
			MeasuredOg:               input.MeasuredOg,
			MeasuredFermenterTopUp:   input.MeasuredFermenterTopUp,
			MeasuredBatchSize:        input.MeasuredBatchSize,
			MeasuredFg:               input.MeasuredFg,
			MeasuredBottlingSize:     input.MeasuredBottlingSize,
			CarbonationTemp:          input.CarbonationTemp,
		}
		text, err := svc.UpdateBatch(ctx, input.ID, body)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newGetBatchLastReadingHandler(svc *service.BatchService) mcp.ToolHandlerFor[BatchIDInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchIDInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.GetBatchLastReading(ctx, input.ID)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newGetBatchReadingsHandler(svc *service.BatchService) mcp.ToolHandlerFor[BatchIDInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchIDInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.GetBatchReadings(ctx, input.ID)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}

func newGetBatchBrewTrackerHandler(svc *service.BatchService) mcp.ToolHandlerFor[BatchIDInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchIDInput) (*mcp.CallToolResult, any, error) {
		text, err := svc.GetBatchBrewTracker(ctx, input.ID)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return textResult(text), nil, nil
	}
}
