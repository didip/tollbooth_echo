package main

import (
	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_echo"
	"github.com/labstack/echo/v4"
	"net/http"
)

func main() {
	e := echo.New()

	// Create a limiter struct.
	limiter := tollbooth.NewLimiter(1, nil)
	// or
	// var tbOptions limiter.ExpirableOptions
	// tbOptions.DefaultExpirationTTL = time.Second
	// tbOptions.ExpireJobInterval = 0
	// limiter := tollbooth.NewLimiter(1, &tbOptions)

	e.GET("/", echo.HandlerFunc(func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusOK, "OK")
	}), tollbooth_echo.LimitHandler(limiter))

	e.Start(":4444")
}
