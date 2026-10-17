package handler

import "net/http"

func Routes(p *ProductHandler, c *CategoryHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/categories", c.List)
	mux.HandleFunc("POST /api/categories", c.Create)

	mux.HandleFunc("GET /api/products", p.List)
	mux.HandleFunc("POST /api/products", p.Create)
	mux.HandleFunc("GET /api/products/{id}", p.Get)
	mux.HandleFunc("PUT /api/products/{id}", p.Update)
	mux.HandleFunc("DELETE /api/products/{id}", p.Delete)
	mux.HandleFunc("PATCH /api/products/{id}/stock", p.UpdateStock)
	return mux
}
