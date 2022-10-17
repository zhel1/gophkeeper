package v2

import (
	"errors"
	"github.com/labstack/echo/v4"
	"gophkeeper/internal/domain"
	"net/http"
)

func (h *Handler) initMaterialsRoutes(gr *echo.Group) {
	materialsGr := gr.Group("/materials")

	authGr := materialsGr.Group("", h.checkUserIdentity)
	authGr.GET("/text", h.getAllTextData)
	authGr.POST("/text", h.UpdateTextDataByID)
	authGr.PUT("/text", h.CreateNewTextData)
}

func (h Handler) getAllTextData(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	dataArray, err := h.services.Materials.GetAllTextData(c.Request().Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDataNotFound):
			return c.NoContent(http.StatusNoContent)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return c.JSON(http.StatusOK, dataArray)
}

func (h Handler) UpdateTextDataByID(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	var inp domain.TextData
	if err := c.Bind(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.services.Materials.UpdateTextDataByID(c.Request().Context(), userID, inp)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

type newTextDataInput struct {
	Text    string 	`json:"login"`
	Metadata string `json:"password"`
}

func (h Handler) CreateNewTextData(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	var inp newTextDataInput
	if err := c.Bind(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.services.Materials.CreateNewTextData(c.Request().Context(), userID, domain.TextData{
		ID: -1,					//this field fill be ignored
		Text: inp.Text,
		Metadata: inp.Metadata,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}