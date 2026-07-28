package filmkritiken_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainErrors "github.com/DerBlum/filmkritiken-backend/domain/errors"
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/DerBlum/filmkritiken-backend/mocks"
	"github.com/golang/mock/gomock"
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
	var re *domainErrors.RepositoryError
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
	var re *domainErrors.RepositoryError
	if !errors.As(err, &re) {
		t.Errorf("Expected RepositoryError but got %v", err)
	}
}

func TestFilmkritikenServiceImpl_UpdateBesprochenAm(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)

	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	filmkritikenId := "fk_1"
	besprochenAm := time.Date(2024, 10, 18, 20, 0, 0, 0, time.UTC)

	filmkritikenRepository.EXPECT().UpdateBesprochenAm(ctx, filmkritikenId, besprochenAm).Return(nil)

	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// when
	err := service.UpdateBesprochenAm(ctx, filmkritikenId, besprochenAm)

	// then
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFilmkritikenServiceImpl_UpdateBesprochenAm_NotFound(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)

	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	filmkritikenId := "fk_doesnotexist"
	besprochenAm := time.Date(2024, 10, 18, 20, 0, 0, 0, time.UTC)

	filmkritikenRepository.EXPECT().
		UpdateBesprochenAm(ctx, filmkritikenId, besprochenAm).
		Return(domainErrors.NewNotFoundErrorFromString("Filmkritiken konnten nicht gefunden werden."))

	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// when
	err := service.UpdateBesprochenAm(ctx, filmkritikenId, besprochenAm)

	// then
	if err == nil {
		t.Error("expected error but got none")
		return
	}
	var nfe *domainErrors.NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("Expected NotFoundError but got %v", err)
	}
}

func TestFilmkritikenServiceImpl_SetKritik_Success(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	fkID := "fk_1"
	user := "Stefan"
	existingFK := &filmkritiken.Filmkritiken{
		Id: fkID,
		Film: &filmkritiken.Film{Titel: "Test Film"},
		Details: &filmkritiken.FilmkritikenDetails{BewertungOffen: true},
		Bewertungen: make([]*filmkritiken.Bewertung, 0),
	}

	filmkritikenRepository.EXPECT().FindFilmkritiken(ctx, fkID).Return(existingFK, nil)
	filmkritikenRepository.EXPECT().SaveFilmkritiken(ctx, gomock.Any()).DoAndReturn(func(c context.Context, fk *filmkritiken.Filmkritiken) error {
		if len(fk.Bewertungen) != 1 {
			t.Errorf("expected 1 bewertung, got %d", len(fk.Bewertungen))
		}
		if fk.Bewertungen[0].Wertung != 8 || fk.Bewertungen[0].Enthaltung != false {
			t.Errorf("unexpected bewertung values: %+v", fk.Bewertungen[0])
		}
		return nil
	})

	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// when
	err := service.SetKritik(ctx, fkID, user, 8, false)

	// then
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFilmkritikenServiceImpl_SetKritik_Enthaltung(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	fkID := "fk_1"
	user := "Stefan"
	existingFK := &filmkritiken.Filmkritiken{
		Id: fkID,
		Film: &filmkritiken.Film{Titel: "Test Film"},
		Details: &filmkritiken.FilmkritikenDetails{BewertungOffen: true},
		Bewertungen: make([]*filmkritiken.Bewertung, 0),
	}

	filmkritikenRepository.EXPECT().FindFilmkritiken(ctx, fkID).Return(existingFK, nil)
	filmkritikenRepository.EXPECT().SaveFilmkritiken(ctx, gomock.Any()).DoAndReturn(func(c context.Context, fk *filmkritiken.Filmkritiken) error {
		if len(fk.Bewertungen) != 1 {
			t.Errorf("expected 1 bewertung, got %d", len(fk.Bewertungen))
		}
		if fk.Bewertungen[0].Enthaltung != true {
			t.Errorf("expected enthaltung=true, got %+v", fk.Bewertungen[0])
		}
		return nil
	})

	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// when
	err := service.SetKritik(ctx, fkID, user, 0, true)

	// then
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFilmkritikenServiceImpl_SetKritik_InvalidWertung(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// when
	err := service.SetKritik(ctx, "fk_1", "Stefan", 15, false)

	// then
	if err == nil {
		t.Error("expected error for wertung 15, got none")
	}
}

func TestFilmkritikenServiceImpl_GetFilmkritiken(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	filter := &filmkritiken.FilmkritikenFilter{
		Limit:      10,
		Suche:      "Matrix",
		Sortierung: "neueste",
	}

	expectedResult := []*filmkritiken.Filmkritiken{
		{Id: "fk_1", Film: &filmkritiken.Film{Titel: "Matrix"}},
	}

	filmkritikenRepository.EXPECT().GetFilmkritiken(ctx, filter).Return(expectedResult, int64(1), nil)

	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// when
	result, totalCount, err := service.GetFilmkritiken(ctx, filter)

	// then
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if totalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", totalCount)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestFilmkritikenServiceImpl_GetFilterOptions_Caching(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	filmkritikenRepository := mocks.NewMockFilmkritikenRepository(ctrl)
	imageRepository := mocks.NewMockImageRepository(ctrl)

	ctx := context.Background()
	expectedOpts := &filmkritiken.FilterOptions{
		Jahre:       []int{2024, 2023},
		Beitragende: []string{"Alice", "Bob"},
	}

	// Expect GetFilterOptions to be called ONLY ONCE on repo due to caching
	filmkritikenRepository.EXPECT().GetFilterOptions(ctx).Return(expectedOpts, nil).Times(1)

	service := filmkritiken.NewFilmkritikenService(filmkritikenRepository, imageRepository)

	// First call -> fetches from repo
	opts1, err1 := service.GetFilterOptions(ctx)
	if err1 != nil {
		t.Fatalf("unexpected error 1: %v", err1)
	}

	// Second call -> returns cached result without hitting repo
	opts2, err2 := service.GetFilterOptions(ctx)
	if err2 != nil {
		t.Fatalf("unexpected error 2: %v", err2)
	}

	if len(opts1.Jahre) != len(opts2.Jahre) || len(opts1.Beitragende) != len(opts2.Beitragende) {
		t.Errorf("cached options mismatch: %+v vs %+v", opts1, opts2)
	}
}
