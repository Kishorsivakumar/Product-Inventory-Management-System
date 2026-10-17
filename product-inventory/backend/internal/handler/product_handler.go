package handler

import (
	"net/http"
	"strconv"

	"inventory/internal/model"
	"inventory/internal/service"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func pathID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := model.ListFilter{Search: q.Get("search"), Status: q.Get("status")}
	if c := q.Get("category_id"); c != "" {
		id, err := strconv.ParseInt(c, 10, 64)
		if err != nil {
			badRequest(w, "category_id must be a number")
			return
		}
		f.CategoryID = id
	}
	if f.Status != "" && !model.ValidStatus(f.Status) {
		badRequest(w, "invalid status filter")
		return
	}
	products, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		badRequest(w, "invalid product id")
		return
	}
	p, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.ProductInput
	if err := decode(w, r, &in); err != nil {
		badRequest(w, "invalid JSON body")
		return
	}
	p, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		badRequest(w, "invalid product id")
		return
	}
	var in model.ProductInput
	if err := decode(w, r, &in); err != nil {
		badRequest(w, "invalid JSON body")
		return
	}
	p, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		badRequest(w, "invalid product id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		badRequest(w, "invalid product id")
		return
	}
	var in model.StockUpdate
	if err := decode(w, r, &in); err != nil {
		badRequest(w, "invalid JSON body")
		return
	}
	p, err := h.svc.UpdateStock(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}
