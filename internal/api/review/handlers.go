package review

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/JulianDominic/GatheringTheBulk/internal/api/common"
	"github.com/JulianDominic/GatheringTheBulk/internal/models"
	"github.com/JulianDominic/GatheringTheBulk/internal/store"
)

type Handler struct {
	Store    store.Store
	Renderer *common.Renderer
}

func (h *Handler) HandleContent(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListReviewItems()
	if err != nil {
		http.Error(w, "Error listing items", http.StatusInternalServerError)
		return
	}

	data := struct {
		Items []models.ReviewItem
	}{
		Items: items,
	}

	h.Renderer.RenderPartial(w, "review.html", data)
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteReviewItem(id); err != nil {
		log.Printf("Failed to delete review item: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "review-count-updated")
	h.HandleContent(w, r)
}

func (h *Handler) HandleResolveModal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)
	item, err := h.Store.GetReviewItem(id)
	if err != nil {
		http.Error(w, "Review item not found", http.StatusNotFound)
		return
	}

	var rawData map[string]string
	json.Unmarshal([]byte(item.RawData), &rawData)

	var proposedValues map[string]interface{}
	json.Unmarshal([]byte(item.ProposedValues), &proposedValues)

	data := struct {
		models.ReviewItem
		RawDataMap        map[string]string
		ProposedValuesMap map[string]interface{}
	}{
		ReviewItem:        *item,
		RawDataMap:        rawData,
		ProposedValuesMap: proposedValues,
	}

	h.Renderer.RenderPartial(w, "partials/resolve_modal.html", data)
}

func (h *Handler) HandleResolveSelect(w http.ResponseWriter, r *http.Request) {
	scryfallID := r.PathValue("scryfall_id")
	qID := r.URL.Query().Get("q_id")
	id, _ := strconv.Atoi(qID)

	card, err := h.Store.GetCardByScryfallID(scryfallID)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	reviewItem, err := h.Store.GetReviewItem(id)
	if err != nil {
		http.Error(w, "Review item not found", http.StatusNotFound)
		return
	}

	var proposedValues map[string]interface{}
	json.Unmarshal([]byte(reviewItem.ProposedValues), &proposedValues)

	data := struct {
		*store.CardSearchResult
		QueueID   int
		Quantity  interface{}
		Condition interface{}
		IsFoil    interface{}
		Language  interface{}
	}{
		CardSearchResult: card,
		QueueID:          id,
		Quantity:         proposedValues["quantity"],
		Condition:        proposedValues["condition"],
		IsFoil:           proposedValues["is_foil"],
		Language:         proposedValues["language"],
	}

	h.Renderer.RenderPartial(w, "partials/resolve_select.html", data)
}

func (h *Handler) HandleResolve(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QueueID    int    `json:"queue_id"`
		ScryfallID string `json:"scryfall_id"`
		Quantity   int    `json:"quantity"`
		Condition  string `json:"condition"`
		IsFoil     bool   `json:"is_foil"`
		Language   string `json:"language"`
	}

	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad JSON", http.StatusBadRequest)
			return
		}
	} else {
		r.ParseForm()
		req.QueueID, _ = strconv.Atoi(r.FormValue("queue_id"))
		req.ScryfallID = r.FormValue("scryfall_id")
		req.Quantity, _ = strconv.Atoi(r.FormValue("quantity"))
		req.Condition = r.FormValue("condition")
		req.IsFoil = r.FormValue("is_foil") == "true" || r.FormValue("is_foil") == "on"
		req.Language = r.FormValue("language")
	}

	item := models.InventoryItem{
		ScryfallID: req.ScryfallID,
		Quantity:   req.Quantity,
		Condition:  req.Condition,
		IsFoil:     req.IsFoil,
		Language:   req.Language,
		Location:   "Imported",
	}

	if err := h.Store.AddInventory(item); err != nil {
		log.Printf("Resolved item add failed: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	if err := h.Store.DeleteReviewItem(req.QueueID); err != nil {
		log.Printf("Resolved item delete failed: %v", err)
	}

	w.Header().Set("HX-Trigger", "review-count-updated")
	h.HandleContent(w, r)
}

func (h *Handler) HandleBadge(w http.ResponseWriter, r *http.Request) {
	count, _ := h.Store.CountReviewItems()
	display := "none"
	if count > 0 {
		display = "inline-flex"
	}
	fmt.Fprintf(w, `<span id="review-badge"
        hx-get="/api/review/badge"
        hx-trigger="review-count-updated from:body"
        hx-target="this"
        hx-push-url="false"
        hx-swap="outerHTML"
        style="background-color: var(--danger); color: white; display: %s; align-items: center; justify-content: center; padding: 0 6px; border-radius: 12px; min-width: 20px; height: 20px; font-size: 11px; font-weight: bold; line-height: 1;">%d</span>`,
		display, count)
}
