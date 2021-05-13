package filmkritiken

import (
	"context"
	"time"
)

type (
	FilmkritikenService interface {
		GetFilmkritiken(ctx context.Context, filter *FilmkritikenFilter) ([]*Filmkritiken, error)
		CreateFilm(ctx context.Context, film *Film, von string, besprochenam *time.Time) (*Filmkritiken, error)
		SetKritik(ctx context.Context, filmkritikenId string, von string) error
	}

	FilmkritikenRepository interface {
		GetFilmkritiken(ctx context.Context, filter *FilmkritikenFilter) ([]*Filmkritiken, error)
		SaveFilmkritiken(ctx context.Context, filmkritiken *Filmkritiken) error
	}

	filmkritikenServiceImpl struct {
		filmkritikenRepository FilmkritikenRepository
	}
)

func NewFilmkritikenService(filmkritikenRepository FilmkritikenRepository) FilmkritikenService {
	return &filmkritikenServiceImpl{
		filmkritikenRepository: filmkritikenRepository,
	}
}

func (f *filmkritikenServiceImpl) GetFilmkritiken(ctx context.Context, filter *FilmkritikenFilter) ([]*Filmkritiken, error) {
	return make([]*Filmkritiken, 0), nil
}

func (f *filmkritikenServiceImpl) CreateFilm(ctx context.Context, film *Film, von string, besprochenam *time.Time) (*Filmkritiken, error) {

	filmkritiken := &Filmkritiken{
		Film: film,
		Details: &FilmkritikenDetails{
			BeitragVon:   von,
			BesprochenAm: besprochenam,
		},
		Bewertungen: make([]*Bewertung, 0),
	}

	err := f.filmkritikenRepository.SaveFilmkritiken(ctx, filmkritiken)
	if err != nil {
		return nil, NewRepositoryError(err)
	}

	return filmkritiken, nil
}

func (f *filmkritikenServiceImpl) SetKritik(ctx context.Context, filmkritikenId string, von string) error {

	return nil
}
