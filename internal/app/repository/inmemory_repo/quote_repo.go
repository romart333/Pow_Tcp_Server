package inmemory_repo

import (
	"Pow_Tcp_Server/internal/app/domain"
	"math/rand"
)

type InMemoryQuoteRepo struct {
	quotes []*domain.Quote
}

func NewInMemoryQuoteRepo() *InMemoryQuoteRepo {
	quotes := []*domain.Quote{
		{
			Text:   "It is not the strongest of the species that survives, but the most adaptable",
			Author: "Charles Darwin",
		},
		{
			Text:   "The journey of a thousand miles begins with one step",
			Author: "Lao Tzu",
		},
		{
			Text:   "To be, or not to be, that is the question",
			Author: "William Shakespeare",
		},
		{
			Text:   "All that we see or seem is but a dream within a dream",
			Author: "Edgar Allan Poe",
		},
		{
			Text:   "The only thing we have to fear is fear itself",
			Author: "Franklin D. Roosevelt",
		},
		{
			Text:   "Life is what happens when you're busy making other plans",
			Author: "John Lennon",
		},
		{
			Text:   "You miss 100% of the shots you don't take",
			Author: "Wayne Gretzky",
		},
		{
			Text:   "The unexamined life is not worth living",
			Author: "Socrates",
		},
	}
	return &InMemoryQuoteRepo{quotes: quotes}
}

func (r *InMemoryQuoteRepo) GetRandom() (*domain.Quote, error) {
	if len(r.quotes) == 0 {
		return nil, domain.ErrQuoteNotFound
	}
	return r.quotes[rand.Intn(len(r.quotes))], nil
}
