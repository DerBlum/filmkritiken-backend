package filmkritiken

import (
	"context"
	"fmt"
)

type (
	FilmkritikenService interface {
		GetFilmkritiken(ctx context.Context, filter *FilmkritikenFilter) ([]*Filmkritiken, error)
		CreateFilm(ctx context.Context, film *Film, filmkritikenDetails *FilmkritikenDetails, imageBites *[]byte) (*Filmkritiken, error)
		OpenCloseBewertungen(ctx context.Context, filmkritikenId string, offen bool) error
		SetKritik(ctx context.Context, filmkritikenId string, von string, bewertung int) error
		LoadImage(ctx context.Context, imageId string) (*[]byte, error)
	}

	FilmkritikenRepository interface {
		FindFilmkritiken(ctx context.Context, filmkritikenId string) (*Filmkritiken, error)
		GetFilmkritiken(ctx context.Context, filter *FilmkritikenFilter) ([]*Filmkritiken, error)
		SaveFilmkritiken(ctx context.Context, filmkritiken *Filmkritiken) error
	}

	ImageRepository interface {
		FindImage(ctx context.Context, imageId string) (*[]byte, error)
		SaveImage(ctx context.Context, imageBites *[]byte) (string, error)
		DeleteImage(ctx context.Context, id string) error
	}

	filmkritikenServiceImpl struct {
		filmkritikenRepository FilmkritikenRepository
		imageRepository        ImageRepository
	}
)

func NewFilmkritikenService(filmkritikenRepository FilmkritikenRepository, imageRepository ImageRepository) FilmkritikenService {
	return &filmkritikenServiceImpl{
		filmkritikenRepository: filmkritikenRepository,
		imageRepository:        imageRepository,
	}
}

func (f *filmkritikenServiceImpl) GetFilmkritiken(ctx context.Context, filter *FilmkritikenFilter) ([]*Filmkritiken, error) {
	filmkritiken, err := f.filmkritikenRepository.GetFilmkritiken(ctx, filter)
	if err != nil {
		return nil, err
	}

	return filmkritiken, nil
}

func (f *filmkritikenServiceImpl) CreateFilm(ctx context.Context, film *Film, filmkritikenDetails *FilmkritikenDetails, imageBites *[]byte) (*Filmkritiken, error) {
	filmkritiken := &Filmkritiken{
		Film:        film,
		Details:     filmkritikenDetails,
		Bewertungen: make([]*Bewertung, 0),
	}

	imageId, err := f.imageRepository.SaveImage(ctx, imageBites)
	if err != nil {
		return nil, NewRepositoryError(err)
	}
	film.Image.Id = imageId

	err = f.filmkritikenRepository.SaveFilmkritiken(ctx, filmkritiken)
	if err != nil {
		_ = f.imageRepository.DeleteImage(ctx, imageId)
		return nil, NewRepositoryError(err)
	}

	return filmkritiken, nil
}

func (f *filmkritikenServiceImpl) OpenCloseBewertungen(ctx context.Context, filmkritikenId string, offen bool) error {

	filmkritiken, err := f.filmkritikenRepository.FindFilmkritiken(ctx, filmkritikenId)
	if err != nil {
		return err
	}

	filmkritiken.Details.BewertungOffen = offen

	err = f.filmkritikenRepository.SaveFilmkritiken(ctx, filmkritiken)
	if err != nil {
		// TODO: anderer error string?
		return NewRepositoryError(err)
	}

	return nil

}

func (f *filmkritikenServiceImpl) SetKritik(ctx context.Context, filmkritikenId string, von string, wertung int) error {

	if wertung < 1 || wertung > 10 {
		return NewInvalidInputErrorFromString("Wertung muss zwischen 1 und 10 liegen.")
	}

	filmkritiken, err := f.filmkritikenRepository.FindFilmkritiken(ctx, filmkritikenId)
	if err != nil {
		return err
	}

	if !filmkritiken.Details.BewertungOffen {
		return NewInvalidInputErrorFromString(fmt.Sprintf("Die Bewertung von %s ist nicht mehr möglich.", filmkritiken.Film.Titel))
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
		filmkritiken.Bewertungen = append(
			filmkritiken.Bewertungen, &Bewertung{
				Von:        von,
				Wertung:    wertung,
				Enthaltung: false,
			},
		)
	}

	err = f.filmkritikenRepository.SaveFilmkritiken(ctx, filmkritiken)
	if err != nil {
		// TODO: anderer error string?
		return NewRepositoryError(err)
	}

	return nil
}

func (f *filmkritikenServiceImpl) LoadImage(ctx context.Context, imageId string) (*[]byte, error) {
	imageBites, err := f.imageRepository.FindImage(ctx, imageId)
	if err != nil {
		return nil, err
	}

	return imageBites, nil
}
