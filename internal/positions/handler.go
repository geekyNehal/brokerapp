package positions

import (
	"encoding/json"
	"net/http"

	"brokerapp/internal/db"

	"github.com/go-chi/chi/v5"
)

type Position struct {
	Symbol        string  `json:"symbol"`
	Quantity      int     `json:"quantity"`
	EntryPrice    float64 `json:"entry_price"`
	CurrentPrice  float64 `json:"current_price"`
	UnrealizedPNL float64 `json:"unrealized_pnl"`
}

type PositionsResponse struct {
	Positions []Position `json:"positions"`
	Summary   struct {
		TotalUnrealizedPNL float64 `json:"total_unrealized_pnl"`
		TotalRealizedPNL   float64 `json:"total_realized_pnl"`
		TotalPNL           float64 `json:"total_pnl"`
	} `json:"summary"`
}

type Handler struct {
	db *db.MySQL
}

func NewHandler(db *db.MySQL) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/positions", h.GetPositions)
}

func (h *Handler) GetPositions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	query := `
		SELECT symbol, quantity, entry_price, current_price, unrealized_pnl
		FROM positions
		WHERE user_id = ?
	`

	rows, err := h.db.Query(r.Context(), query, userID)
	if err != nil {
		http.Error(w, "Failed to fetch positions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var positions []Position
	for rows.Next() {
		var p Position
		if err := rows.Scan(&p.Symbol, &p.Quantity, &p.EntryPrice, &p.CurrentPrice, &p.UnrealizedPNL); err != nil {
			http.Error(w, "Failed to scan positions", http.StatusInternalServerError)
			return
		}
		positions = append(positions, p)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to process positions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}
