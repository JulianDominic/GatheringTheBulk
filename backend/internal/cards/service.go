package cards

import "log/slog"

type CardsService struct {
	repo *CardsRepo
}

func NewCardsService(repo *CardsRepo) *CardsService {
	return &CardsService{repo: repo}
}

func (s *CardsService) AddCardBySetAndCollectorNumber(set string, collector_number int, quantity int) {
	slog.Info("Adding card into database", "set", set, "collector_number", collector_number, "quantity", quantity)
	s.repo.AddCardBySetAndCollectorNumber(set, collector_number, quantity)
	slog.Info("Successfully added card into database", "set", set, "collector_number", collector_number, "quantity", quantity)
}
