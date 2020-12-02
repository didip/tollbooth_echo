## tollbooth_echo

[Echo](https://github.com/webx-top/echo) middleware for rate limiting HTTP requests.


## Five Minutes Tutorial

```
package main

import (
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_echo"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

func main() {
	e := echo.New()

	// Create a limiter struct.
	limiter := tollbooth.NewLimiter(1, nil)
	// or
	// var tbOptions  limiter.ExpirableOptions
	// tbOptions.DefaultExpirationTTL = time.Second
	// tbOptions.ExpireJobInterval = 0
	// limiter := tollbooth.NewLimiter(1, &tbOptions)

	e.Get("/", echo.HandlerFunc(func(c echo.Context) error {
		return c.String("Hello, World!", 200)
	}), tollbooth_echo.LimitHandler(limiter))

	e.Run(standard.New(":4444"))
}

```