package plugin

import (
	"context"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
)

var (
	_ backend.CallResourceHandler   = (*App)(nil)
	_ instancemgmt.InstanceDisposer = (*App)(nil)
	_ backend.CheckHealthHandler    = (*App)(nil)
)

type App struct {
	backend.CallResourceHandler
}

func NewApp(_ context.Context, _ backend.AppInstanceSettings) (instancemgmt.Instance, error) {
	var app App
	mux := http.NewServeMux()
	app.registerRoutes(mux)
	app.CallResourceHandler = httpadapter.New(mux)
	return &app, nil
}

func (a *App) Dispose() {}

// CallResource routes datasource requests to the datasource handler and
// everything else to the app's httpadapter mux.
func (a *App) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	if req.PluginContext.DataSourceInstanceSettings != nil {
		return handleDatasourceCallResource(ctx, req, sender)
	}
	return a.CallResourceHandler.CallResource(ctx, req, sender)
}

// CheckHealth handles health checks for both the app plugin and the datasource sub-plugin.
func (a *App) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if req.PluginContext.DataSourceInstanceSettings != nil {
		return checkDatasourceHealth(ctx, req)
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "ok",
	}, nil
}
