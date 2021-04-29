package filmkritiken

type FilmkritikenService interface {
	GetFilmkritiken(limit int, offset int) ([]*Filmkritiken, error)
}

type filmkritikenServiceImpl struct{}

func NewFilmkritikenService() FilmkritikenService {
	return &filmkritikenServiceImpl{}
}

func (f *filmkritikenServiceImpl) GetFilmkritiken(limit int, offset int) ([]*Filmkritiken, error) {
	return make([]*Filmkritiken, 0), nil
}
