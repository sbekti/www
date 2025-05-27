package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// PublicHandler handles public website HTTP requests
type PublicHandler struct{}

// NewPublicHandler creates a new public handler
func NewPublicHandler() *PublicHandler {
	return &PublicHandler{}
}

// Home handles GET /
func (h *PublicHandler) Home(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

// Resume handles GET /resume
func (h *PublicHandler) Resume(c echo.Context) error {
	return c.Render(http.StatusOK, "resume.html", nil)
}

// Blog handles GET /blog
func (h *PublicHandler) Blog(c echo.Context) error {
	return c.Render(http.StatusOK, "blog.html", nil)
} 