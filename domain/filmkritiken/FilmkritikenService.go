package filmkritiken

import (
	"context"
	"time"
)

type (
	FilmkritikenService interface {
		GetFilmkritiken(ctx context.Context, filter *FilmkritikenFilter) ([]*Filmkritiken, error)
		CreateFilm(ctx context.Context, film *Film, von string, besprochenam *time.Time) (*Filmkritiken, error)
		SetKritik(ctx context.Context, filmkritikenId string, von string, bewertung int) error
	}

	FilmkritikenRepository interface {
		FindFilmkritiken(ctx context.Context, filmkritikenId string) (*Filmkritiken, error)
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
	filmkritiken, err := f.filmkritikenRepository.GetFilmkritiken(ctx, filter)
	if err != nil {
		return nil, err
	}

	return filmkritiken, nil
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
		// TODO: Anderer Error String?
		return nil, NewRepositoryError(err)
	}

	return filmkritiken, nil
}

func (f *filmkritikenServiceImpl) SetKritik(ctx context.Context, filmkritikenId string, von string, wertung int) error {

	if wertung < 1 || wertung > 10 {
		return NewInvalidInputErrorFromString("Wertung muss zwischen 1 und 10 liegen.")
	}

	filmkritiken, err := f.filmkritikenRepository.FindFilmkritiken(ctx, filmkritikenId)
	if err != nil {
		return err
	}

	found := false
	for _, bewertung := range filmkritiken.Bewertungen {
		if bewertung.Von == von {
			bewertung.Wertung = wertung
			found = true
			break
		}
	}
	if !found {
		filmkritiken.Bewertungen = append(filmkritiken.Bewertungen, &Bewertung{
			Von:        von,
			Wertung:    wertung,
			Enthaltung: false,
		})
	}

	err = f.filmkritikenRepository.SaveFilmkritiken(ctx, filmkritiken)
	if err != nil {
		// TODO: anderer error string?
		return NewRepositoryError(err)
	}

	return nil
}
