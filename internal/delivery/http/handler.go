package http

import (
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	v1 "gophkeeper/internal/delivery/http/v1"
	v2 "gophkeeper/internal/delivery/http/v2"
	"gophkeeper/internal/service"
	"gophkeeper/pkg/auth"
	"net/http"
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

//**********************************************************************************************************************
// chi router
//**********************************************************************************************************************
func (h *Handler) Init() http.Handler {
	router := chi.NewRouter()
	router.Use(
		chimiddleware.Compress(5),
	)

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		//c.String(http.StatusOK, "pong") // TODO Ping
	})

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router chi.Router) {
	handlerV1 := v1.NewHandler(h.services, h.tokenManager)
	router.Route("/api", func(r chi.Router) {
		handlerV1.Init(r)
	})
}

//**********************************************************************************************************************
// echo router
//**********************************************************************************************************************
func (h *Handler) InitEcho() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	//e.Use(middleware.Recover())
	//e.Use(middleware.Gzip())
	//e.Use(middleware.Decompress())

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	h.initAPIEcho(e)

	return e
}

func (h *Handler) initAPIEcho(e *echo.Echo) {
	api := e.Group("/api")
	handlerV2 := v2.NewHandler(h.services, h.tokenManager)
	handlerV2.Init(api)
}
