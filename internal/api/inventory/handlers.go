package inventory

import (
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

func (h *Handler) HandleAdd(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	if qty < 1 {
		qty = 1
	}

	item := models.InventoryItem{
		ScryfallID: r.FormValue("scryfall_id"),
		Quantity:   qty,
		Condition:  r.FormValue("condition"),
		IsFoil:     r.FormValue("is_foil") == "on",
		Language:   r.FormValue("language"),
		Location:   "Binder",
	}

	if err := h.Store.AddInventory(item); err != nil {
		log.Printf("Failed to add inventory: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) HandleEdit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	pathID := r.PathValue("id")
	id, err := strconv.Atoi(pathID)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	if qty < 1 {
		qty = 1
	}

	item := models.InventoryItem{
		ID:        id,
		Quantity:  qty,
		Condition: r.FormValue("condition"),
		IsFoil:    r.FormValue("is_foil") == "on",
		Language:  r.FormValue("language"),
		Location:  "Binder",
	}

	if loc := r.FormValue("location"); loc != "" {
		item.Location = loc
	}

	if err := h.Store.UpdateInventory(item); err != nil {
		log.Printf("Failed to update inventory: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteInventory(id); err != nil {
		log.Printf("Failed to delete inventory: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleEditModal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)
	item, err := h.Store.GetInventoryByID(id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	h.Renderer.RenderPartial(w, "partials/edit_modal.html", item)
}

func (h *Handler) HandleAddDetails(w http.ResponseWriter, r *http.Request) {
	scryfallID := r.PathValue("scryfall_id")
	card, err := h.Store.GetCardByScryfallID(scryfallID)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}
	h.Renderer.RenderPartial(w, "partials/add_card_details.html", card)
}

func (h *Handler) HandleAutocomplete(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		return
	}

	names, err := h.Store.SearchInventoryNames(q)
	if err != nil {
		log.Printf("Autocomplete error: %v", err)
		return
	}

	for _, n := range names {
		fmt.Fprintf(w, `<li role="option" class="search-result-item" hx-get="/?q=%s" hx-target="body" hx-push-url="true">%s</li>`, n, n)
	}
}

func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	set := r.URL.Query().Get("set")

	if len(q) < 2 {
		return
	}

	results, err := h.Store.SearchCards(q, set)
	if err != nil {
		log.Printf("Search error: %v", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	mode := r.URL.Query().Get("mode")
	qID := r.URL.Query().Get("q_id")

	for _, c := range results {
		var hxAttr string
		if mode == "resolve" {
			hxAttr = fmt.Sprintf(`hx-get="/review/resolve-select/%s?q_id=%s" hx-target="#res-confirmation"`, c.ScryfallID, qID)
		} else {
			hxAttr = fmt.Sprintf(`hx-get="/inventory/add-details/%s" hx-target="#selected-card-container"`, c.ScryfallID)
		}

		fmt.Fprintf(w, `
			<li role="option" 
				class="search-result-item"
				style="display:flex; gap:1rem; align-items:center;"
				%s
				hx-trigger="click"
				hx-on:click="document.getElementById('res-confirmation').style.display='block'">
				<img src="%s" style="height:40px; border-radius:4px; flex-shrink:0;">
				<div style="flex-grow:1; overflow:hidden;">
					<strong style="color:var(--text-primary); display:block; white-space:nowrap; overflow:hidden; text-overflow:ellipsis;">%s</strong>
					<small style="display:block; color:var(--text-secondary); text-transform: uppercase;">%s #%s</small>
				</div>
			</li>`,
			hxAttr, c.ImageURI, c.Name, c.SetCode, c.CollectorNumber)
	}
}
