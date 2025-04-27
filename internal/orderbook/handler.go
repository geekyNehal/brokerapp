package orderbook

import (
	"encoding/json"
	"net/http"

	"brokerapp/internal/db"

	"github.com/go-chi/chi/v5"
)

type Order struct {
	ID        int64   `json:"id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"` // "buy" or "sell"
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

type CreateOrderRequest struct {
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"` // "buy" or "sell"
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type PNL struct {
	Unrealized float64 `json:"unrealized"`
	Realized   float64 `json:"realized"`
	Total      float64 `json:"total"`
}

type OrderbookResponse struct {
	Orders []Order `json:"orders"`
	PNL    PNL     `json:"pnl"`
}

type Handler struct {
	db *db.MySQL
}

func NewHandler(db *db.MySQL) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/orderbook", h.GetOrderbook)
	r.Post("/orders", h.CreateOrder)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate side
	if req.Side != "buy" && req.Side != "sell" {
		http.Error(w, "Invalid side: must be 'buy' or 'sell'", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO orders (user_id, symbol, side, price, quantity, status)
		VALUES (?, ?, ?, ?, ?, 'pending')
	`

	_, err := h.db.Exec(r.Context(), query,
		userID,
		req.Symbol,
		req.Side,
		req.Price,
		req.Quantity,
	)
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetOrderbook(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	query := `
		SELECT id, symbol, side, price, quantity, status, created_at
		FROM orders
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := h.db.Query(r.Context(), query, userID)
	if err != nil {
		http.Error(w, "Failed to fetch orderbook", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Symbol, &o.Side, &o.Price, &o.Quantity, &o.Status, &o.CreatedAt); err != nil {
			http.Error(w, "Failed to scan orders", http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to process orders", http.StatusInternalServerError)
		return
	}

	// Get PNL
	pnlQuery := `
		SELECT 
			COALESCE(SUM(unrealized_pnl), 0) as unrealized,
			COALESCE(SUM(realized_pnl), 0) as realized,
			COALESCE(SUM(total_pnl), 0) as total
		FROM positions
		WHERE user_id = ?
	`

	var pnl PNL
	err = h.db.QueryRow(r.Context(), pnlQuery, userID).Scan(
		&pnl.Unrealized,
		&pnl.Realized,
		&pnl.Total,
	)
	if err != nil {
		http.Error(w, "Failed to fetch PNL", http.StatusInternalServerError)
		return
	}

	response := OrderbookResponse{
		Orders: orders,
		PNL:    pnl,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
