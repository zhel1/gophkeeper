package v2

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"gophkeeper/internal/domain"
	"net/http"
	"time"
)

func (h *Handler) initMaterialsRoutes(gr *echo.Group) {
	materialsGr := gr.Group("/materials")

	authGr := materialsGr.Group("", h.checkUserIdentity)
	authGr.GET("/text", h.getAllTextData)
	authGr.POST("/text", h.updateTextDataByID)
	authGr.PUT("/text", h.createNewTextData)

	authGr.GET("/card", h.getAllCardData)
	authGr.POST("/card", h.updateCardDataByID)
	authGr.PUT("/card", h.createNewCardData)

	authGr.GET("/cred", h.getAllCredData)
	authGr.POST("/cred", h.updateCredDataByID)
	authGr.PUT("/cred", h.createNewCredData)
}

//**********************************************************************************************************************
// Text
//**********************************************************************************************************************
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

func (h Handler) updateTextDataByID(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	var inp domain.TextData
	if err := json.NewDecoder(c.Request().Body).Decode(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.services.Materials.UpdateTextDataByID(c.Request().Context(), userID, inp)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

type newTextDataInput struct {
	Text     string `json:"text"`
	Metadata string `json:"metadata"`
}

func (h Handler) createNewTextData(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	var inp newTextDataInput
	if err := json.NewDecoder(c.Request().Body).Decode(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.services.Materials.CreateNewTextData(c.Request().Context(), userID, domain.TextData{
		ID:       -1, //this field fill be ignored
		Text:     inp.Text,
		Metadata: inp.Metadata,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

//**********************************************************************************************************************
// Credit card
//**********************************************************************************************************************
type newCardDataInput struct {
	CardNumber string    `json:"card_number"`
	ExpDate    time.Time `json:"exp_data"`
	CVV        string    `json:"cvv"`
	Name       string    `json:"name"`
	Surname    string    `json:"surname"`
	Metadata   string    `json:"metadata"`
}

func (h Handler) getAllCardData(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	dataArray, err := h.services.Materials.GetAllCardData(c.Request().Context(), userID)
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

func (h Handler) updateCardDataByID(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	var inp domain.CardData
	if err := json.NewDecoder(c.Request().Body).Decode(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.services.Materials.UpdateCardDataByID(c.Request().Context(), userID, inp)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h Handler) createNewCardData(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	var inp newCardDataInput
	if err := json.NewDecoder(c.Request().Body).Decode(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.services.Materials.CreateNewCardData(c.Request().Context(), userID, domain.CardData{
		ID:         -1, //this field fill be ignored
		CardNumber: inp.CardNumber,
		ExpDate:    inp.ExpDate,
		CVV:        inp.CVV,
		Name:       inp.Name,
		Surname:    inp.Surname,
		Metadata:   inp.Metadata,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

//**********************************************************************************************************************
// Creds
//**********************************************************************************************************************
func (h Handler) getAllCredData(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	dataArray, err := h.services.Materials.GetAllCredData(c.Request().Context(), userID)
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

func (h Handler) updateCredDataByID(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	var inp domain.CredData
	if err := json.NewDecoder(c.Request().Body).Decode(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.services.Materials.UpdateCredDataByID(c.Request().Context(), userID, inp)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

type newCredDataInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Metadata string `json:"metadata"`
}

func (h Handler) createNewCredData(c echo.Context) error {
	userID := c.Get(UserIDCtxName.String()).(int)

	var inp newCredDataInput
	if err := json.NewDecoder(c.Request().Body).Decode(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.services.Materials.CreateNewCredData(c.Request().Context(), userID, domain.CredData{
		ID:       -1, //this field fill be ignored
		Login:    inp.Login,
		Password: inp.Password,
		Metadata: inp.Metadata,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
