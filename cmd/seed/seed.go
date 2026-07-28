package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetFilmkritiken(ctx context.Context, filter *filmkritiken.FilmkritikenFilter) ([]*filmkritiken.Filmkritiken, int64, error)
	SaveImage(ctx context.Context, imageBites *[]byte) (string, error)
	SaveFilmkritiken(ctx context.Context, filmkritiken *filmkritiken.Filmkritiken) error
}

func seedIfEmpty(ctx context.Context, repo Repository) error {
	existing, _, err := repo.GetFilmkritiken(ctx, &filmkritiken.FilmkritikenFilter{Limit: 1})
	if err != nil {
		log.Warnf("Could not check if database is empty: %v", err)
	} else if len(existing) > 0 {
		log.Infof("Database already contains filmkritiken (%d+ found), skipping seed", len(existing))
		return nil
	}

	log.Info("Database is empty. Populating initial film collection from Bruno requests and posters...")
	return seed(ctx, repo)
}

func seed(ctx context.Context, repo Repository) error {
	// 1x1 transparent PNG fallback image
	dummyImage := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
		0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}

	loadImage := func(filename string) []byte {
		path := filepath.Join("cmd", "seed", "posters", filename)
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			log.Warnf("Could not load poster image %s: %v (using fallback)", path, err)
			return dummyImage
		}
		return data
	}

	parseTime := func(s string) *time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil
		}
		return &t
	}

	type seedItem struct {
		posterFilename string
		fk             *filmkritiken.Filmkritiken
	}

	items := []seedItem{
		{
			posterFilename: "AQuietPlace.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Nico",
					BesprochenAm:   parseTime("2026-08-15T12:00:00Z"),
					BewertungOffen: true,
				},
				Film: &filmkritiken.Film{
					Titel:            "A Quiet Place",
					Altersfreigabe:   16,
					Erscheinungsjahr: 2018,
					Regie:            "John Krasinski",
					Laenge:           90,
					Originalsprache:  "Englisch, Amerikanische Zeichensprache",
					Produktionsland:  "Vereinigte Staaten",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{},
			},
		},
		{
			posterFilename: "KarateKid.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Stefan",
					BesprochenAm:   parseTime("2026-07-17T12:00:00Z"),
					BewertungOffen: false,
				},
				Film: &filmkritiken.Film{
					Titel:            "Karate Kid",
					Originaltitel:    "The Karate Kid",
					Altersfreigabe:   12,
					Erscheinungsjahr: 1984,
					Regie:            "John G. Avildsen",
					Laenge:           126,
					Originalsprache:  "Englisch",
					Produktionsland:  "Vereinigte Staaten",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{
					{Von: "Stefan", Wertung: 9, Enthaltung: false},
					{Von: "Nico", Wertung: 8, Enthaltung: false},
					{Von: "Dani", Wertung: 9, Enthaltung: false},
					{Von: "Tiffy", Wertung: 7, Enthaltung: false},
				},
			},
		},
		{
			posterFilename: "CitizenKane.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Dani",
					BesprochenAm:   parseTime("2021-06-13T12:00:00Z"),
					BewertungOffen: true,
				},
				Film: &filmkritiken.Film{
					Titel:            "Citizen Kane",
					Altersfreigabe:   12,
					Erscheinungsjahr: 1941,
					Regie:            "Orson Welles",
					Laenge:           119,
					Originalsprache:  "Englisch",
					Produktionsland:  "Vereinigte Staaten",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{},
			},
		},
		{
			posterFilename: "TropicThunder.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Flo",
					BesprochenAm:   parseTime("2021-06-06T12:00:00Z"),
					BewertungOffen: true,
				},
				Film: &filmkritiken.Film{
					Titel:            "Tropic Thunder",
					Altersfreigabe:   16,
					Erscheinungsjahr: 2008,
					Regie:            "Ben Stiller",
					Laenge:           106,
					Originalsprache:  "Englisch",
					Produktionsland:  "Vereinigte Staaten, Vereinigtes Königreich, Deutschland",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{
					{Von: "Flo", Wertung: 8, Enthaltung: false},
					{Von: "Dani", Wertung: 7, Enthaltung: false},
				},
			},
		},
		{
			posterFilename: "MallCop.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Tiffy",
					BesprochenAm:   parseTime("2021-06-19T12:00:00Z"),
					BewertungOffen: true,
				},
				Film: &filmkritiken.Film{
					Titel:            "Der Kaufhaus Cop",
					Originaltitel:    "Paul Blart: Mall Cop",
					Altersfreigabe:   6,
					Erscheinungsjahr: 2009,
					Regie:            "Steve Carr",
					Laenge:           87,
					Originalsprache:  "Englisch",
					Produktionsland:  "Vereinigte Staaten",
					Image:            &filmkritiken.Image{Copyright: "Moviepilot"},
				},
				Bewertungen: []*filmkritiken.Bewertung{},
			},
		},
		{
			posterFilename: "SchindlersListe.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Flo",
					BesprochenAm:   parseTime("2021-07-19T12:00:00Z"),
					BewertungOffen: false,
				},
				Film: &filmkritiken.Film{
					Titel:            "Schindlers Liste",
					Altersfreigabe:   12,
					Erscheinungsjahr: 1993,
					Regie:            "Steven Spielberg",
					Laenge:           195,
					Originalsprache:  "Englisch, Deutsch, Polnisch, Hebräisch",
					Produktionsland:  "Vereinigte Staaten",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{
					{Von: "Stefan", Wertung: 10, Enthaltung: false},
					{Von: "Flo", Wertung: 9, Enthaltung: false},
					{Von: "Nico", Wertung: 10, Enthaltung: false},
				},
			},
		},
		{
			posterFilename: "TaxiDriver.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Dani",
					BesprochenAm:   parseTime("2021-07-24T12:00:00Z"),
					BewertungOffen: false,
				},
				Film: &filmkritiken.Film{
					Titel:            "Taxi Driver",
					Altersfreigabe:   16,
					Erscheinungsjahr: 1976,
					Regie:            "Martin Scorsese",
					Laenge:           114,
					Originalsprache:  "Englisch",
					Produktionsland:  "Vereinigte Staaten",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{
					{Von: "Dani", Wertung: 9, Enthaltung: false},
					{Von: "Stefan", Wertung: 8, Enthaltung: false},
				},
			},
		},
		{
			posterFilename: "GunsAkimbo.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Tiffy",
					BesprochenAm:   parseTime("2021-08-15T12:00:00Z"),
					BewertungOffen: false,
				},
				Film: &filmkritiken.Film{
					Titel:            "Guns Akimbo",
					Altersfreigabe:   16,
					Erscheinungsjahr: 2019,
					Regie:            "Jason Lei Howden",
					Laenge:           97,
					Originalsprache:  "Englisch",
					Produktionsland:  "Neuseeland, Vereinigtes Königreich, Deutschland",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{
					{Von: "Tiffy", Wertung: 7, Enthaltung: false},
				},
			},
		},
		{
			posterFilename: "Alien.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Nico",
					BesprochenAm:   parseTime("2021-08-21T12:00:00Z"),
					BewertungOffen: false,
				},
				Film: &filmkritiken.Film{
					Titel:            "Alien – Das unheimliche Wesen aus einer fremden Welt",
					Originaltitel:    "Alien",
					Altersfreigabe:   16,
					Erscheinungsjahr: 1979,
					Regie:            "Ridley Scott",
					Laenge:           117,
					Originalsprache:  "Englisch",
					Produktionsland:  "Vereinigtes Königreich, Vereinigte Staaten",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{
					{Von: "Nico", Wertung: 9, Enthaltung: false},
					{Von: "Stefan", Wertung: 9, Enthaltung: false},
				},
			},
		},
		{
			posterFilename: "Zombiber.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Flo",
					BesprochenAm:   parseTime("2021-09-19T12:00:00Z"),
					BewertungOffen: true,
				},
				Film: &filmkritiken.Film{
					Titel:            "Zombiber",
					Originaltitel:    "Zombeavers",
					Altersfreigabe:   16,
					Erscheinungsjahr: 2014,
					Regie:            "Jordan Rubin",
					Laenge:           77,
					Originalsprache:  "Englisch",
					Produktionsland:  "Vereinigte Staaten",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{},
			},
		},
		{
			posterFilename: "FistfulOfDollars.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Dani",
					BesprochenAm:   parseTime("2021-10-02T12:00:00Z"),
					BewertungOffen: true,
				},
				Film: &filmkritiken.Film{
					Titel:            "Für eine Handvoll Dollar",
					Originaltitel:    "Per un pugno di dollari",
					Altersfreigabe:   16,
					Erscheinungsjahr: 1964,
					Regie:            "Sergio Leone",
					Laenge:           100,
					Originalsprache:  "Italienisch, Englisch, Spanisch",
					Produktionsland:  "Italien, Spanien, Deutschland",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{},
			},
		},
		{
			posterFilename: "PerAnhalter.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Tiffy",
					BesprochenAm:   parseTime("2021-10-17T12:00:00Z"),
					BewertungOffen: true,
				},
				Film: &filmkritiken.Film{
					Titel:            "Per Anhalter durch die Galaxis",
					Originaltitel:    "The Hitchhiker’s Guide to the Galaxy",
					Altersfreigabe:   6,
					Erscheinungsjahr: 2005,
					Regie:            "Garth Jennings",
					Laenge:           110,
					Originalsprache:  "Englisch",
					Produktionsland:  "Vereinigtes Königreich",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{},
			},
		},
		{
			posterFilename: "AlienVsPredator2.jpg",
			fk: &filmkritiken.Filmkritiken{
				Details: &filmkritiken.FilmkritikenDetails{
					BeitragVon:     "Nico",
					BesprochenAm:   parseTime("2021-10-31T12:00:00Z"),
					BewertungOffen: false,
				},
				Film: &filmkritiken.Film{
					Titel:            "Alien vs. Predator 2",
					Originaltitel:    "Aliens vs. Predator: Requiem",
					Altersfreigabe:   18,
					Erscheinungsjahr: 2007,
					Regie:            "Colin Strause, Greg Strause",
					Laenge:           90,
					Originalsprache:  "Englisch",
					Produktionsland:  "Vereinigte Staaten, Kanada",
					Image:            &filmkritiken.Image{Copyright: "IMDb"},
				},
				Bewertungen: []*filmkritiken.Bewertung{
					{Von: "Nico", Wertung: 4, Enthaltung: false},
					{Von: "Stefan", Wertung: 5, Enthaltung: false},
				},
			},
		},
	}

	for _, item := range items {
		imgBytes := loadImage(item.posterFilename)
		imageId, err := repo.SaveImage(ctx, &imgBytes)
		if err != nil {
			log.Warnf("Could not save image for '%s': %v", item.fk.Film.Titel, err)
		} else {
			item.fk.Film.Image.Id = imageId
		}

		if err := repo.SaveFilmkritiken(ctx, item.fk); err != nil {
			log.Errorf("Failed to save seed filmkritik '%s': %v", item.fk.Film.Titel, err)
		}
	}

	log.Infof("Successfully seeded database with %d filmkritiken with posters!", len(items))
	return nil
}
