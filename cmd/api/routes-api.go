package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

// routes returns a chi router with all the routes defined.
func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	mux.Post("/api/payment-intent", app.GetPaymentIntent)

	mux.Get("/api/maize/{id}", app.GetMaizeByID)

	mux.Post("/api/create-customer-and-subscribe-to-plan", app.CreateCustomerAndSubscribeToPlan)

	mux.Post("/api/authenticate", app.CreateAuthToken)
	mux.Post("/api/is-authenticated", app.CheckAuthentication)
	mux.Post("/api/forgot-password", app.SendPasswordResetEmail)
	mux.Post("/api/reset-password", app.ResetPassword)

	mux.Route("/api/admin", func(mux chi.Router) {
		mux.Use(app.Auth)

		mux.Post("/virtual-terminal-succeeded", app.VirtualTerminalPaymentSucceeded)
		mux.Post("/all-sales", app.AllSales)
		mux.Post("/all-subs", app.AllSubs)

		mux.Post("/get-sale/{id}", app.GetSale)
		mux.Post("/refund", app.RefundPayment)
		mux.Post("/cancel-sub", app.CancelSub)

		mux.Post("/all-users", app.AllUsers)
		mux.Post("/all-users/{id}", app.OneUser)
		mux.Post("/all-users/edit/{id}", app.EditUser)
		mux.Post("/all-users/delete/{id}", app.DeleteUser)

	})

	return mux
}
