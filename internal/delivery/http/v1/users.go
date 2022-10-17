package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"gophkeeper/internal/domain"
	"gophkeeper/internal/service"
	"net/http"
)

//TODO make all handlers private?

func (h *Handler) initUserRoutes(r chi.Router) {
	r.Route("/user", func(r chi.Router) {
		r.Post("/sign-up", h.userSignUp())
		r.Post("/sign-in", h.userSignIn())
		r.Post("/refresh", h.userRefresh())

		r.Group(func(r chi.Router) {
			r.Use(h.checkUserIdentity)

			r.Get("/orders", h.Test())
		})
	})
}

//TODO tags
type signInInput struct {
	Login    string `json:"login" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

//user registration
func (h *Handler)userSignUp() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var inp signInInput
		if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := h.services.Users.SignUp(r.Context(), service.UserSignUpInput {
			Login:    inp.Login,
			Password: inp.Password,
		})

		if err != nil {
			if errors.Is(err, domain.ErrUserAlreadyExists) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//authentication
		token, err := h.services.Users.SignIn(r.Context(), service.UserSignInInput {
			Login:    inp.Login,
			Password: inp.Password,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//TODO use headers, not cookie
		cookie := &http.Cookie{
			Name: "AccessToken", //TODO make constant
			Value: token.AccessToken,//TODO add refresh token
			Path:  "/",
		}
		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusOK)
	}
}

type signUpInput struct {
	Login    string `json:"login" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

//user authentication
func (h *Handler)userSignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var inp signUpInput
		if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token, err := h.services.Users.SignIn(r.Context(), service.UserSignInInput {
			Login:    inp.Login,
			Password: inp.Password,
		})

		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie := &http.Cookie{
			Name: "AccessToken", //TODO make constant
			Value: token.AccessToken,//TODO add refresh token
			Path:  "/",
		}
		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusOK)
	}
}

//TODO tags
type refreshInput struct {
	Token string `json:"token" binding:"required"`
}

func (h *Handler)userRefresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var inp refreshInput
		if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := h.services.Users.RefreshTokens(r.Context(), inp.Token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie := &http.Cookie{
			Name: "AccessToken", //TODO make constant
			Value: res.AccessToken,//TODO add refresh token
			Path:  "/",
		}
		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler)Test() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}