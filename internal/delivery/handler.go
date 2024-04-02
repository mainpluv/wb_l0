package delivery

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mainpluv/wb_l0/internal/service"
)

type Handler struct {
	OrderService service.OrderService
}
type Report struct {
	Error string `json:"omitempty"`
}

func NewHandler(OrderService service.OrderService) *Handler {
	return &Handler{
		OrderService: OrderService,
	}
}

func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	err := h.renderTemplate(w, "static/index.html", nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("error loading page: %s", err.Error()), http.StatusBadRequest)
		return
	}
}

func (h *Handler) renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		return err
	}
	if err = t.Execute(w, data); err != nil {
		return err
	}
	return nil
}

func (h *Handler) GetOrderByIdHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Получаем данные из URL
	orderUUID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		MyError(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Ищем заказ по UUID
	order, err := h.OrderService.GetOrder(orderUUID)
	if err != nil {
		MyError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем ответ
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		MyError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func MyError(w http.ResponseWriter, errorMessage string, statusCode int) {
	report := Report{Error: errorMessage}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) InitRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/orders", h.HandleHome).Methods("GET")
	router.HandleFunc("/odeers/{id}", h.GetOrderByIdHandler).Methods("GET")
	return router
}
