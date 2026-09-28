package main

import (
	"context"
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/tiagogb/ssa-io-t/pkg/plugin"
)

func main() {
	// Instantiate the app once. Our NewApp ignores AppInstanceSettings so a
	// singleton is fine. Using backend.Serve (instead of app.Manage) avoids the
	// instance manager trying to extract AppInstanceSettings from datasource
	// requests, which panics with "app instance settings cannot be nil".
	inst, err := plugin.NewApp(context.Background(), backend.AppInstanceSettings{})
	if err != nil {
		log.DefaultLogger.Error("failed to create app", "error", err)
		os.Exit(1)
	}
	app := inst.(*plugin.App)

	if err := backend.Serve(backend.ServeOpts{
		CheckHealthHandler:  app,
		CallResourceHandler: app,
	}); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
