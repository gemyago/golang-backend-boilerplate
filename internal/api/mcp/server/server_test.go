package server

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/lifecycle"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServer(t *testing.T) {
	fake := faker.New()
	makeMockDeps := func() MCPServerDeps {
		return MCPServerDeps{
			RootLogger:    telemetry.RootTestLogger(),
			ShutdownHooks: lifecycle.NewTestShutdownHooks(),
			Controllers:   []ToolsFactory{},
			IDGen:         ident.NewDefaultGenerator(),
		}
	}

	makeToolCallRequest := func() mcp.CallToolRequest {
		return mcp.CallToolRequest{
			Request: mcp.Request{
				Method: "tools/call",
			},
			Header: http.Header{},
			Params: mcp.CallToolParams{
				Name: "tool-1-" + fake.Lorem().Word(),
			},
		}
	}

	newToolCallResult := func() *mcp.CallToolResult {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fake.Lorem().Sentence(10)),
			},
		}
	}

	newToolsFactories := func(
		toolName string,
		handler mcpserver.ToolHandlerFunc,
	) []ToolsFactory {
		return []ToolsFactory{
			ToolsFactoryFunc(func() []mcpserver.ServerTool {
				return []mcpserver.ServerTool{
					{
						Tool:    mcp.Tool{Name: toolName},
						Handler: handler,
					},
				}
			}),
		}
	}

	t.Run("middleware", func(t *testing.T) {
		t.Run("should process success tool call", func(t *testing.T) {
			deps := makeMockDeps()

			wantCall := makeToolCallRequest()
			wantResult := newToolCallResult()

			deps.Controllers = newToolsFactories(
				wantCall.Params.Name,
				func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					assert.NotNil(t, ctx)
					assert.Equal(t, wantCall, req)
					return wantResult, nil
				})
			srv := NewMCPServer(deps)
			ctx := t.Context()
			testServer := newTestMCPServer()
			err := testServer.Start(ctx, srv.mcpServer)
			require.NoError(t, err)

			client := testServer.Client()

			gotResult, err := client.CallTool(ctx, wantCall)
			require.NoError(t, err)
			assert.Equal(t, wantResult, gotResult)
		})

		t.Run("should setup correlation id in context", func(t *testing.T) {
			deps := makeMockDeps()

			wantCall := makeToolCallRequest()

			contextChecked := false
			deps.Controllers = newToolsFactories(
				wantCall.Params.Name,
				func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					diagCtx := telemetry.GetLogAttributesFromContext(ctx)
					assert.NotEmpty(t, diagCtx.CorrelationID)
					contextChecked = true
					return newToolCallResult(), nil
				})
			srv := NewMCPServer(deps)
			ctx := t.Context()
			testServer := newTestMCPServer()
			err := testServer.Start(ctx, srv.mcpServer)
			require.NoError(t, err)

			client := testServer.Client()

			_, err = client.CallTool(ctx, wantCall)
			require.NoError(t, err)
			assert.True(t, contextChecked)
		})

		t.Run("should reuse correlation id from context", func(t *testing.T) {
			deps := makeMockDeps()

			wantCorrelationID := fake.UUID().V4()
			wantCall := makeToolCallRequest()

			callCtx := telemetry.SetLogAttributesToContext(t.Context(), telemetry.LogAttributes{
				CorrelationID: slog.StringValue(wantCorrelationID),
			})

			contextChecked := false
			deps.Controllers = newToolsFactories(
				wantCall.Params.Name,
				func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					diagCtx := telemetry.GetLogAttributesFromContext(ctx)
					assert.Equal(t, wantCorrelationID, diagCtx.CorrelationID.String())
					contextChecked = true
					return newToolCallResult(), nil
				})
			srv := NewMCPServer(deps)
			testServer := newTestMCPServer()
			err := testServer.Start(callCtx, srv.mcpServer)
			require.NoError(t, err)

			client := testServer.Client()

			_, err = client.CallTool(callCtx, wantCall)
			require.NoError(t, err)
			assert.True(t, contextChecked)
		})

		t.Run("should respond with error if tool call fails", func(t *testing.T) {
			deps := makeMockDeps()

			wantCall := makeToolCallRequest()
			wantError := errors.New(fake.Lorem().Sentence(10))

			deps.Controllers = newToolsFactories(
				wantCall.Params.Name,
				func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					assert.NotNil(t, ctx)
					assert.Equal(t, wantCall, req)
					return nil, wantError
				})
			srv := NewMCPServer(deps)
			ctx := t.Context()
			testServer := newTestMCPServer()
			err := testServer.Start(ctx, srv.mcpServer)
			require.NoError(t, err)

			client := testServer.Client()

			_, err = client.CallTool(ctx, wantCall)
			require.Error(t, err)
			assert.Contains(t, err.Error(), wantError.Error())
		})
	})
}
