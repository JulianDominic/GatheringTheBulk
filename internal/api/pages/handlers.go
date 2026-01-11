package pages

import (
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

func (h *Handler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	q := r.URL.Query().Get("q")
	pageSize := 20
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if pNum, err := strconv.Atoi(p); err == nil && pNum > 0 {
			page = pNum
		}
	}
	offset := (page - 1) * pageSize

	items, total, err := h.Store.ListInventory(pageSize, offset, q)
	if err != nil {
		log.Printf("Error listing inventory: %v", err)
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	data := struct {
		Items      []models.InventoryItem
		Total      int
		Query      string
		Page       int
		TotalPages int
		PageSize   int
		HasPrev    bool
		HasNext    bool
	}{
		Items:      items,
		Total:      total,
		Query:      q,
		Page:       page,
		TotalPages: totalPages,
		PageSize:   pageSize,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
	}

	h.Renderer.Render(w, r, "index.html", data)
}

func (h *Handler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	lastSync, err := h.Store.GetSetting("scryfall_last_sync")
	if err != nil {
		log.Printf("Error fetching setting: %v", err)
	}

	data := struct {
		LastSync string
	}{
		LastSync: lastSync,
	}

	h.Renderer.Render(w, r, "settings.html", data)
}

func (h *Handler) HandleImportHub(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "upload"
	}

	count, _ := h.Store.CountReviewItems()

	data := struct {
		ReviewCount int
		IsUpload    bool
		IsReview    bool
	}{
		ReviewCount: count,
		IsUpload:    tab == "upload",
		IsReview:    tab == "review",
	}

	h.Renderer.Render(w, r, "import.html", data)
}
