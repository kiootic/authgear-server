package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachResetPasswordHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/reset_password").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newResetPasswordHandler))
}

type resetPasswordProvider interface {
	GetResetPasswordForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	PostResetPasswordForm(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ResetPasswordHandler struct {
	Provider  resetPasswordProvider
	TxContext db.TxContext
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.TxContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetResetPasswordForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			writeResponse, err := h.Provider.PostResetPasswordForm(w, r)
			writeResponse(err)
			return err
		}
		return nil
	})
}
