package v2

import (
	"github.com/labstack/echo/v4"
	"gophkeeper/internal/service"
	"gophkeeper/pkg/auth"
)

type Handler struct {
	services     *service.Services
	tokenManager auth.TokenManager
}

func NewHandler(services *service.Services, tokenManager auth.TokenManager) *Handler {
	return &Handler{
		services:     services,
		tokenManager: tokenManager,
	}
}

func (h *Handler) Init(g *echo.Group) {
	h.initUserRoutes(g)
	h.initMaterialsRoutes(g)
}