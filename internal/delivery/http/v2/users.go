package v2

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"gophkeeper/internal/domain"
	"gophkeeper/internal/service"
	"net/http"
	"time"
)

//TODO make all handlers private?

func (h *Handler) initUserRoutes(gr *echo.Group) {
	userGr := gr.Group("/user")

	//TODO use echo jwt tokens
	userGr.POST("/auth/sign-up", h.userSignUp)
	userGr.POST("/auth/sign-in", h.userSignIn)
	userGr.POST("/auth/refresh", h.userRefresh)
}

//TODO tags
type signInInput struct {
	Login    string `json:"login" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

//user registration
func (h Handler) userSignUp(c echo.Context) error {
	var inp signInInput
	if err := json.NewDecoder(c.Request().Body).Decode(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	//registration
	err := h.services.Users.SignUp(c.Request().Context(), service.UserSignUpInput {
		Login:    inp.Login,
		Password: inp.Password,
	})

	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	//log in
	tokens, err := h.services.Users.SignIn(c.Request().Context(), service.UserSignInInput {
		Login:    inp.Login,
		Password: inp.Password,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.SetCookie(&http.Cookie{
		Name: "RefreshToken", 						  //TODO make constant
		Value: tokens.RefreshToken,
		Path:  "/api/user/auth/refresh", 				  //TODO construct it somehow automatically (routers may change)
		Domain: "",									  //TODO set domen from configuration
		Expires: time.Now().Add(40 * 24 * time.Hour), //TODO Expire time for RefreshToken from config
	})

	return c.JSON(http.StatusOK, tokens)
}

type signUpInput struct {
	Login    string `json:"login" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

//user authorization
func (h Handler) userSignIn(c echo.Context) error {
	var inp signUpInput
	if err := json.NewDecoder(c.Request().Body).Decode(&inp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	//log in
	tokens, err := h.services.Users.SignIn(c.Request().Context(), service.UserSignInInput {
		Login:    inp.Login,
		Password: inp.Password,
	})

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.SetCookie(&http.Cookie{
		Name: "RefreshToken", 						  //TODO make constant
		Value: tokens.RefreshToken,
		Path:  "/api/user/auth/refresh", 			  //TODO construct it somehow automatically (routers may change)
		Domain: "",									  //TODO set domen from configuration
		Expires: time.Now().Add(40 * 24 * time.Hour), //TODO Expire time for RefreshToken from config
	})

	return c.JSON(http.StatusOK, tokens)
}

//TODO tags
type refreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

//TODO logout
//TODO add logout tokens to the black list until they are spoiled

func (h Handler) userRefresh(c echo.Context) error {
	//TODO add old tokens to the black list until they are spoiled

	// For browsers we can use cookie.
	// Those apps, which don't support cookie, should provide refresh token in body
	var inp refreshInput
	for _, cookie := range c.Cookies() {
		if cookie.Name == "RefreshToken" { //TODO make constant
			inp.RefreshToken = cookie.Value
			break
		}
	}

	if inp.RefreshToken == "" {
		if err := json.NewDecoder(c.Request().Body).Decode(&inp); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	tokens, err := h.services.Users.RefreshTokens(c.Request().Context(), inp.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.SetCookie(&http.Cookie{
		Name: "RefreshToken", 						  //TODO make constant
		Value: tokens.RefreshToken,
		Path:  "/api/user/auth/refresh", 				  //TODO construct it somehow automatically (routers may change)
		Domain: "",									  //TODO set domen from configuration
		Expires: time.Now().Add(40 * 24 * time.Hour), //TODO Expire time for RefreshToken from config
	})

	return c.JSON(http.StatusOK, tokens)
}