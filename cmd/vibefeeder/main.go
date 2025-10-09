package main

import (
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tjanas94/vibefeeder/internal/view"
)

func main() {
	// Get executable directory
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exeDir := filepath.Dir(exe)
	staticDir := filepath.Join(exeDir, "static")

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("2M"))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			return next(c)
		}
	})

	e.Static("/static", staticDir)

	e.GET("/", func(c echo.Context) error {
		return view.Hello("World").Render(c.Request().Context(), c.Response().Writer)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
