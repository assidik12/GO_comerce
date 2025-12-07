package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/pkg/response"
	"github.com/assidik12/go-restfull-api/internal/service"
	"github.com/julienschmidt/httprouter"
)

type ProductHandler struct {
	service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (ph *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Baca query parameter 'page'
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1" // Default ke halaman 1 jika tidak ada
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		response.BadRequest(w, "Invalid page parameter")
		return
	}

	// 2. Baca query parameter 'pageSize'
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "10" // Default ke 10 item per halaman
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		response.BadRequest(w, "Invalid pageSize parameter")
		return
	}

	// 3. Panggil service dengan parameter paginasi
	products, err := ph.service.GetAllProducts(r.Context(), page, pageSize)
	if err != nil {
		// Gunakan InternalServerError karena error dari service/repo lebih mungkin masalah server
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, products)
}

func (ph *ProductHandler) GetProductById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	product, err := ph.service.GetProductById(r.Context(), idInt)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	response.OK(w, product)
}

func (ph *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req dto.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	product, err := ph.service.CreateProduct(r.Context(), req)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.Created(w, product)
}

func (ph *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	var req dto.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	product, err := ph.service.UpdateProduct(r.Context(), idInt, req)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, product)
}

func (ph *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	err = ph.service.DeleteProduct(r.Context(), idInt)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}
	response.OK(w, nil)
}
