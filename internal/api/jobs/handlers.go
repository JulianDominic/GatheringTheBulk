package jobs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/JulianDominic/GatheringTheBulk/internal/models"
	"github.com/JulianDominic/GatheringTheBulk/internal/store"
	"github.com/JulianDominic/GatheringTheBulk/internal/worker"
	"github.com/google/uuid"
)

type Handler struct {
	Store      store.Store
	Dispatcher *worker.Dispatcher
}

func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	job, err := h.Store.GetJob(id)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		if job.Status == models.JobStatusCompleted {
			var completionHTML string
			var triggers string

			if job.Type == models.JobTypeSyncDB {
				completionHTML = fmt.Sprintf(`
					<div style="background-color:#2e7d32; color:white; padding:1rem; border-radius:4px; margin-top:1rem;">
						<strong>Scryfall Sync Complete!</strong>
						<p style="margin-bottom:0; margin-top:0.5rem;">%s</p>
					</div>`, job.ResultSummary)
				triggers = "document.body.dispatchEvent(new CustomEvent('scryfall-synced'));"
			} else {
				var res struct {
					Success int `json:"success"`
					Review  int `json:"review"`
				}
				json.Unmarshal([]byte(job.ResultSummary), &res)

				completionHTML = fmt.Sprintf(`
					<div style="background-color:#2e7d32; color:white; padding:1rem; border-radius:4px; margin-top:1rem;">
						<strong>Import Complete!</strong>
						<ul style="margin-bottom:0; margin-top:0.5rem;">
							<li>Successfully Added: <strong>%d</strong></li>
							<li>Sent to Review: <strong>%d</strong></li>
						</ul>
					</div>`, res.Success, res.Review)
				triggers = `
					document.body.dispatchEvent(new CustomEvent('review-count-updated'));
					const reviewLoader = document.querySelector('#review-tab > div');
					if(reviewLoader) { htmx.trigger(reviewLoader, 'reveal'); }
				`
			}

			w.Write([]byte(fmt.Sprintf(`
				%s
				<script>
					(function() {
						%s
					})();
				</script>
			`, completionHTML, triggers)))
		} else if job.Status == models.JobStatusFailed {
			fmt.Fprintf(w, `<div class="pico-color-red"><strong>Failed:</strong> %s</div>`, job.ResultSummary)
		} else {
			content := fmt.Sprintf(`<p>Processing... %d items</p><progress value="%d" max="%d"></progress>`,
				job.ProgressCurrent, job.ProgressCurrent, job.ProgressTotal)

			if job.Type == models.JobTypeSyncDB {
				content = fmt.Sprintf(`
					<div style="display:flex; align-items:center; gap:1rem; padding:1rem; background:var(--surface-color); border:1px solid var(--border-color); border-radius:8px;">
						<div aria-busy="true"></div>
						<div>
							<strong style="display:block;">Syncing Scryfall Database...</strong>
							<small style="color:var(--text-secondary);">Processed %d cards so far. This may take a few minutes.</small>
						</div>
					</div>`, job.ProgressCurrent)
			}

			fmt.Fprintf(w, `<div hx-get="/api/jobs/%s" hx-trigger="every 1s" hx-swap="outerHTML">
				%s
			</div>`, job.ID, content)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}
}

func (h *Handler) HandleSync(w http.ResponseWriter, r *http.Request) {
	jobID := uuid.New().String()
	job := &models.Job{
		ID:        jobID,
		Type:      models.JobTypeSyncDB,
		Status:    models.JobStatusPending,
		CreatedAt: time.Now(),
	}

	if err := h.Store.CreateJob(job); err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	h.Dispatcher.QueueJob(worker.JobRequest{
		Job:     job,
		Handler: worker.SyncDatabaseTask,
	})

	fmt.Fprintf(w, `<div hx-get="/api/jobs/%s" hx-trigger="load delay:500ms, every 1s" hx-swap="outerHTML">
		<div style="display:flex; align-items:center; gap:1rem; padding:1rem; background:var(--surface-color); border:1px solid var(--border-color); border-radius:8px;">
			<div aria-busy="true"></div>
			<div>
				<strong style="display:block;">Initializing Sync...</strong>
				<small style="color:var(--text-secondary);">Connecting to Scryfall API</small>
			</div>
		</div>
    </div>`, jobID)
}

func (h *Handler) HandleImport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("csv_file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	jobID := uuid.New().String()

	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	dstPath := filepath.Join("uploads", jobID+".csv")
	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("Failed to create upload file: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		log.Printf("Failed to save file: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	dst.Close()

	job := &models.Job{
		ID:        jobID,
		Type:      models.JobTypeCSVImport,
		Status:    models.JobStatusPending,
		CreatedAt: time.Now(),
	}

	if err := h.Store.CreateJob(job); err != nil {
		os.Remove(dstPath)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	h.Dispatcher.QueueJob(worker.JobRequest{
		Job:     job,
		Handler: worker.ImportCSVTask,
	})

	fmt.Fprintf(w, `<div hx-get="/api/jobs/%s" hx-trigger="load delay:500ms, every 1s" hx-swap="outerHTML">
        <p>Importing...</p>
        <progress></progress>
    </div>`, jobID)
}
