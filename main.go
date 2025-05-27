package main

import (
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	casbin_mw "github.com/alexferl/echo-casbin"
	"github.com/casbin/casbin/v2"
	"github.com/sbekti/www/handlers"
	"github.com/sbekti/www/util"
)

type (
	Host struct {
		Echo *echo.Echo
	}
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func setupInternApp() *echo.Echo {
	intern := echo.New()

	enforcer, err := casbin.NewEnforcer("model.conf", "policy.csv")
	if err != nil {
		log.Fatalf("Failed to create Casbin enforcer: %v", err)
	}
	config := casbin_mw.Config{
		Enforcer:          enforcer,
		EnableRolesHeader: true,
		RolesHeader:       "Remote-Groups",
	}
	intern.Use(casbin_mw.CasbinWithConfig(config))

	// Add auth middleware to parse reverse proxy headers
	intern.Use(util.AuthMiddleware())

	// Load templates using the new template module
	internTemplates, err := util.LoadTemplatesFromDirectory("templates/intern")
	if err != nil {
		log.Fatalf("Failed to load intern templates: %v", err)
	}

	internRenderer := &TemplateRenderer{
		templates: internTemplates,
	}
	intern.Renderer = internRenderer

	intern.GET("/", func(c echo.Context) error {
		authInfo := util.GetAuthInfo(c)
		data := struct {
			Title    string
			AuthInfo *util.AuthInfo
		}{
			Title:    "Intern Home",
			AuthInfo: authInfo,
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	// Device management routes
	deviceHandler := handlers.NewDeviceHandler()
	intern.GET("/devices", deviceHandler.ListDevices)
	intern.GET("/devices/add", deviceHandler.ShowAddForm)
	intern.POST("/devices/add", deviceHandler.AddDevice)
	intern.GET("/devices/edit/:mac", deviceHandler.ShowEditForm)
	intern.POST("/devices/edit/:mac", deviceHandler.UpdateDevice)
	intern.POST("/devices/delete/:mac", deviceHandler.DeleteDevice)

	return intern
}

func setupPublicApp() *echo.Echo {
	public := echo.New()

	// Load templates using the new template module
	publicTemplates, err := util.LoadTemplatesFromDirectory("templates/public")
	if err != nil {
		log.Fatalf("Failed to load public templates: %v", err)
	}

	publicRenderer := &TemplateRenderer{
		templates: publicTemplates,
	}
	public.Renderer = publicRenderer

	// Public routes using handler
	publicHandler := handlers.NewPublicHandler()
	public.GET("/", publicHandler.Home)
	public.GET("/resume", publicHandler.Resume)
	public.GET("/blog", publicHandler.Blog)

	return public
}

func main() {
	// Initialize database connection
	if err := util.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer util.CloseDB()

	// Hosts
	hosts := map[string]*Host{}

	// Intern
	hosts["intern.corp.bekti.com"] = &Host{setupInternApp()}
	hosts["intern-dev.corp.bekti.com"] = &Host{setupInternApp()}

	// Server
	e := echo.New()

	// Use Echo's built-in logger middleware, which is already consistent.
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/static", "static")
	e.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		host := hosts[req.Host]
		if host == nil {
			// Log if a host is not found, could indicate misconfiguration or unexpected traffic
			c.Logger().Warnf("Host not found, falling back to public: %s", req.Host)
			host = &Host{setupPublicApp()}
		}

		host.Echo.ServeHTTP(res, req)
		return
	})

	// Start server - Echo's e.Logger.Fatal already handles this well.
	log.Printf("Starting server on :3000")
	e.Logger.Fatal(e.Start(":3000"))
}
