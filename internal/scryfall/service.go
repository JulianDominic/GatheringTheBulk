package scryfall

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/JulianDominic/GatheringTheBulk/pkg/models"
)

type default_cards_uri struct {
	Id           string
	Download_uri string
}

type ScryfallService struct {
	repo *ScryfallRepo
}

func NewScryfallService(repo *ScryfallRepo) *ScryfallService {
	return &ScryfallService{repo: repo}
}

func (s *ScryfallService) PullMasterData() {
	// get the download uri for today's uri
	default_cards_uri := s.getDownloadURI()

	// check if we have already pulled
	if s.repo.HasPulledToday(default_cards_uri.Id) {
		return
	}

	// get all of the card information
	slog.Info("Pulling all of the card data from Scryfall")
	cards := s.getCards(default_cards_uri.Download_uri)
	slog.Info("Finished pulling all of the card data from Scryfall")

	// see what we already have
	existingIDs, err := s.repo.GetAllCardIDs()
	if err != nil {
		slog.Error("Failed to fetch existing IDs", "error", err)
		panic(err)
	}

	// find only the new cards
	var newCards []models.Card
	for _, card := range cards {
		if !existingIDs[card.Id] {
			newCards = append(newCards, card)
		}
	}

	// load all of the (new) cards into the database
	if len(newCards) > 0 {
		slog.Info("Inserting new cards", "count", len(newCards))
		if err := s.repo.AddCardsBatch(newCards); err != nil {
			slog.Error("Failed to batch insert", "error", err)
			panic(err)
		}
	} else {
		slog.Info("No new cards to insert.")
	}

	// write to db that we have pulled for today
	s.repo.WritePullLog(default_cards_uri.Id)
}

func (s *ScryfallService) getDownloadURI() default_cards_uri {
	// get the today's uri
	resp, err := http.Get("https://api.scryfall.com/bulk-data/default_cards")
	if err != nil {
		slog.Error("Could not get bulk-data from Scryfall", "error", err)
		panic(err)
	}
	defer resp.Body.Close()
	default_cards_uri := new(default_cards_uri)
	err = json.NewDecoder(resp.Body).Decode(default_cards_uri)
	if err != nil {
		slog.Error("Failed to parse bulk-data json from Scryfall", "error", err)
		panic(err)
	}

	return *default_cards_uri
}

func (s *ScryfallService) getCards(download_uri string) []models.Card {
	// get the actual card data
	resp, err := http.Get(download_uri)
	if err != nil {
		slog.Error("Could not get bulk-data from Scryfall", "error", err)
		panic(err)
	}
	defer resp.Body.Close()
	cards := new([]models.Card)
	// TODO: Dual faced cards support so that I can get all their info
	err = json.NewDecoder(resp.Body).Decode(cards)
	if err != nil {
		slog.Error("Failed to parse default_cards json from Scryfall", "error", err)
		panic(err)
	}
	return *cards
}
