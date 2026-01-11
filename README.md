# GatheringTheBulk

GatheringTheBulk is a locally hosted inventory management system for Magic: The Gathering (MTG), prioritizing data integrity via Scryfall UUIDs.

## Features
- **Strict Scryfall Integration**: Uses Scryfall UUIDs as the source of truth.
- **Offline Search**: Fast, local prefix search using a synced SQLite database.
- **Async Jobs**: Sync and processing happens in the background.
- **Bulk Import**: Upload `.csv` files to import cards. Ambiguous items trigger a review workflow.
- **Review Queue**: Manually resolve import conflicts or missing data.
- **Pure Go**: No external runtime dependencies (Node/Python) required for the backend.
- **HTMX**: Modern, responsive UI without heavy client-side frameworks.

## Getting Started

### Prerequisites
- Go 1.22+

### Installation
1. Clone the repository.
2. Run `go mod download`.

### Running the Application
```bash
go run cmd/server/main.go
```
The server will start at [http://localhost:8080](http://localhost:8080).

### First Run Setup
1. Go to **Settings**.
2. Click **Update Card Database** to download the latest Scryfall data.
   - This downloads ~400MB of JSON and ingests it. It may take 1-2 minutes.
   - You can monitor progress on the page.
3. Once synced, go to **Dashboard** and click **+ Add Card** to start managing your inventory.

### Importing Cards
1. Go to **Settings** -> **Bulk Import**.
2. Upload a CSV file. It must have headers.
   - Required: `set` and `cn` (Collector Number) OR `name`.
   - Optional: `quantity`, `condition`, `foil`, `language`.
3. Monitor the import job.
4. If items are flagged for review, go to the **Review Queue** tab to resolve them.

## Project Structure
- `cmd/server/`: Main entry point.
- `internal/`: Core application logic (Database, Workers, Scryfall Client).
- `web/templates/`: HTML templates using HTMX and Pico.css.
