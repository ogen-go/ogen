package sampleserver

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Register(r chi.Router, s Server) {
	// paths.K
	r.Route("/pets", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			data, err := s.PetsGet(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			e := json.NewEncoder(w)
			_ = e.Encode(data)
		})
	})
}
