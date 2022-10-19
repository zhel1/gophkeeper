package v2

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type CookieConst string

func (c CookieConst) String() string {
	return string(c)
}

var (
	UserIDCtxName CookieConst = "UserID"
)

//**********************************************************************************************************************
//checks that user is authorised and puts his id into context
func (h *Handler) checkUserIdentity(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var token string

		tokenCookie, err := c.Cookie("AccessToken") //TODO make constant
		if errors.Is(err, http.ErrNoCookie) {       //no cookie
			token = c.Request().Header.Get("Authorization")
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Please, sign in first")
			}
		} else if err != nil { //unknown error
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		} else { //cookie found
			token = tokenCookie.Value
		}

		//fmt.Println("TOKEN: "+ token)

		userID, err := h.tokenManager.Parse(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Please, provide valid credentials. "+err.Error())
		}

		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		c.Set(UserIDCtxName.String(), userIDInt)

		return next(c)
	}
}
