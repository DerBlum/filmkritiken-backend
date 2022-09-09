package filmkritiken_test

import (
	"context"
	"errors"
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/DerBlum/filmkritiken-backend/mocks"
	"github.com/golang/mock/gomock"
	"testing"
)

//go:generate mockgen -source=FilmkritikenService.go -destination=../../mocks/FilmkritikenService.go -package mocks

func TestFilmkritikenServiceImpl_CreateFilm(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)

	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	film := &filmkritiken.Film{
		Image: &filmkritiken.Image{
			Copyright: "IMDb",
		},
	}
	details := &filmkritiken.FilmkritikenDetails{}
	image := []byte("img")

	expectedImageId := "image_1"
	expectedFilmkritiken := &filmkritiken.Filmkritiken{
		Details:     details,
		Film:        film,
		Bewertungen: make([]*filmkritiken.Bewertung, 0),
	}

	imageRepository.EXPECT().SaveImage(ctx, &image).Return(expectedImageId, nil)
	filmkritikenRepository.EXPECT().SaveFilmkritiken(ctx, gomock.Eq(expectedFilmkritiken)).
		DoAndReturn(func(c context.Context, f *filmkritiken.Filmkritiken) error {
			if f.Film.Image.Id != expectedImageId {
				t.Errorf("expected imageId to be %s but was %s", expectedImageId, f.Film.Image.Id)
			}

			return nil
		})

	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// when
	response, err := service.CreateFilm(ctx, film, details, &image)

	// then
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if !gomock.Eq(expectedFilmkritiken).Matches(response) {
		t.Errorf("expected filmkritiken to be %+v but was %+v", expectedFilmkritiken, response)
	}
}

func TestFilmkritikenServiceImpl_CreateFilm_ErrorSaveImage(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)

	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	film := &filmkritiken.Film{
		Image: &filmkritiken.Image{
			Copyright: "IMDb",
		},
	}
	details := &filmkritiken.FilmkritikenDetails{}
	image := []byte("img")

	imageRepository.EXPECT().SaveImage(ctx, &image).Return("", errors.New(""))

	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// when
	_, err := service.CreateFilm(ctx, film, details, &image)

	// then
	if err == nil {
		t.Error("expected error but got none")
		return
	}
	var re *filmkritiken.RepositoryError
	if !errors.As(err, &re) {
		t.Errorf("Expected RepositoryError but got %v", err)
	}
}

func TestFilmkritikenServiceImpl_CreateFilm_ErrorSaveFilmkritiken(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)

	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	film := &filmkritiken.Film{
		Image: &filmkritiken.Image{
			Copyright: "IMDb",
		},
	}
	details := &filmkritiken.FilmkritikenDetails{}
	image := []byte("img")

	expectedImageId := "image_1"
	expectedFilmkritiken := &filmkritiken.Filmkritiken{
		Details:     details,
		Film:        film,
		Bewertungen: make([]*filmkritiken.Bewertung, 0),
	}

	imageRepository.EXPECT().SaveImage(ctx, &image).Return(expectedImageId, nil)
	filmkritikenRepository.EXPECT().SaveFilmkritiken(ctx, gomock.Eq(expectedFilmkritiken)).Return(errors.New(""))
	imageRepository.EXPECT().DeleteImage(ctx, expectedImageId).Return(nil)

	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// when
	_, err := service.CreateFilm(ctx, film, details, &image)

	// then
	if err == nil {
		t.Error("expected error but got none")
		return
	}
	var re *filmkritiken.RepositoryError
	if !errors.As(err, &re) {
		t.Errorf("Expected RepositoryError but got %v", err)
	}
}
