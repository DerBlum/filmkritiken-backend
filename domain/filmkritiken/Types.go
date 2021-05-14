package filmkritiken

import (
	"time"
)

const (
	Context_Username = "username"
	Context_TraceId  = "traceId"
)

type (
	Filmkritiken struct {
		Id          string               `json:"id" bson:"_id"`
		Details     *FilmkritikenDetails `json:"details"`
		Film        *Film                `json:"film"`
		Bewertungen []*Bewertung         `json:"bewertungen"`
	}

	Film struct {
		//Id               UUID   `json:"id" bson:"_id"`
		Titel            string `json:"titel"`
		Altersfreigabe   int    `json:"altersfreigabe"`
		Erscheinungsjahr int    `json:"erscheinungsjahr"`
		Regie            string `json:"regie"`
		Laenge           int    `json:"laenge"`
		Originaltitel    string `json:"originaltitel"`
		Originalsprache  string `json:"originalsprache"`
		Produktionsland  string `json:"produktionsland"`
		Image            *Image `json:"image"`
	}

	Bewertung struct {
		Von        string `json:"von"`
		Wertung    int    `json:"wertung"`
		Enthaltung bool   `json:"enthaltung"`
	}

	Image struct {
		Source    string
		Copyright string
	}

	FilmkritikenDetails struct {
		BeitragVon   string     `json:"beitragvon"`
		BesprochenAm *time.Time `json:"besprochenam"`
	}

	FilmkritikenFilter struct {
		Limit  int
		Offset int
	}
)
