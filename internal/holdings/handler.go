package holdings

import (
	"encoding/json"
	"net/http"

	"brokerapp/internal/db"

	"github.com/go-chi/chi/v5"
)

type Holding struct {
	Symbol   string  `json:"symbol"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Value    float64 `json:"value"`
}

type CreateHoldingRequest struct {
	Symbol   string  `json:"symbol"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type Handler struct {
	db *db.MySQL
}

func NewHandler(db *db.MySQL) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/holdings", h.GetHoldings)
	r.Post("/holdings", h.CreateHolding)
}

func (h *Handler) GetHoldings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	query := `
		SELECT symbol, quantity, price, value
		FROM holdings
		WHERE user_id = ?
	`

	rows, err := h.db.Query(r.Context(), query, userID)
	if err != nil {
		http.Error(w, "Failed to fetch holdings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var holdings []Holding
	for rows.Next() {
		var h Holding
		if err := rows.Scan(&h.Symbol, &h.Quantity, &h.Price, &h.Value); err != nil {
			http.Error(w, "Failed to scan holdings", http.StatusInternalServerError)
			return
		}
		holdings = append(holdings, h)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to process holdings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(holdings)
}

func (h *Handler) CreateHolding(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	var req CreateHoldingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate value
	value := float64(req.Quantity) * req.Price

	query := `
		INSERT INTO holdings (user_id, symbol, quantity, price, value)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(r.Context(), query,
		userID,
		req.Symbol,
		req.Quantity,
		req.Price,
		value,
	)
	if err != nil {
		http.Error(w, "Failed to create holding", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
