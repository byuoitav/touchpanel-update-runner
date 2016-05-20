package main

import (
	"flag"
	"fmt"

	"github.com/byuoitav/touchpanel-update-runner/controllers"
	"github.com/byuoitav/touchpanel-update-runner/helpers"
	"github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/middleware"
)

func main() {
	helpers.TouchpanelStatusMap = make(map[string]helpers.TouchpanelStatus)
	helpers.ValidationStatus = make(map[string]helpers.TouchpanelStatus)

	flag.Parse()

	// Build our channels
	submissionChannel := make(chan helpers.TouchpanelStatus, 50)
	helpers.UpdateChannel = make(chan helpers.TouchpanelStatus, 150)
	helpers.ValidationChannel = make(chan helpers.TouchpanelStatus, 150)

	go helpers.ChannelUpdater()
	go helpers.ValidateHelper()

	// Build a couple controllers--to have access to channels, controllers must be wrapped
	touchpanelUpdateController := controllers.BuildControllerStartTouchpanelUpdate(submissionChannel)
	multipleTouchpanelUpdatesController := controllers.BuildControllerStartMultipleTPUpdate(submissionChannel)

	port := ":8004"
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())

	e.Get("/health", health.Check)

	// Touchpanels
	e.Get("/touchpanel/status", controllers.GetAllTPStatus)
	e.Get("/touchpanel/status/concise", controllers.GetAllTPStatusConcise)
	e.Get("/touchpanel/:address/status", controllers.GetTPStatus)

	e.Post("/touchpanel", multipleTouchpanelUpdatesController)
	e.Post("/touchpanel/:address", touchpanelUpdateController)

	e.Put("/touchpanel", multipleTouchpanelUpdatesController)
	e.Put("/touchpanel/:address", touchpanelUpdateController)

	// Callback
	e.Post("/callback/afterWait", controllers.PostWait)
	e.Post("/callback/afterFTP", controllers.AfterFTPHandle)

	// Validation
	e.Get("/validate/touchpanels/status", controllers.GetValidationStatus)

	e.Post("/validate/touchpanels", controllers.Validate)

	fmt.Printf("The Touchpanel Update Runner is listening on %s\n", port)
	e.Run(fasthttp.New(port))
}
