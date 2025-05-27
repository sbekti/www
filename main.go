package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	casbin_mw "github.com/alexferl/echo-casbin"
	"github.com/casbin/casbin/v2"
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
		panic(err)
	}
	config := casbin_mw.Config{
		Enforcer:          enforcer,
		EnableRolesHeader: true,
		RolesHeader:       "Remote-Groups",
	}
	intern.Use(casbin_mw.CasbinWithConfig(config))

	internRenderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/intern/*.html")),
	}
	intern.Renderer = internRenderer

	intern.GET("/", func(c echo.Context) error {
		data := struct {
			Title string
			Name  string
		}{
			Title: "Intern Home",
			Name:  "Intern User",
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	return intern
}

func setupPublicApp() *echo.Echo {
	public := echo.New()

	publicRenderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/public/*.html")),
	}
	public.Renderer = publicRenderer

	public.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	public.GET("/resume", func(c echo.Context) error {
		return c.Render(http.StatusOK, "resume.html", nil)
	})

	public.GET("/blog", func(c echo.Context) error {
		return c.Render(http.StatusOK, "blog.html", nil)
	})

	return public
}

func main() {
	// Hosts
	hosts := map[string]*Host{}

	// Intern
	hosts["intern.corp.bekti.com"] = &Host{setupInternApp()}

	// Server
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/static", "static")
	e.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		host := hosts[req.Host]
		if host == nil {
			host = &Host{setupPublicApp()}
		}

		host.Echo.ServeHTTP(res, req)
		return
	})

	// Start server
	e.Logger.Fatal(e.Start(":3000"))
}
