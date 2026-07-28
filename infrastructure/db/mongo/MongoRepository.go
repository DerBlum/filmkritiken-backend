package mongo

import (
	"context"
	"regexp"
	"time"

	"github.com/DerBlum/filmkritiken-backend/domain/errors"
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	filmkritikenCollectionName = "filmkritiken"
	imagesCollectionName       = "images"
)

type image struct {
	ImageId string `bson:"_id"`
	Image   *[]byte
}

var updateOpts = options.UpdateOne().SetUpsert(true)

type Config struct {
	ConnectionString string `env:"MONGODB_CONNECTION_URI,unset"`
	Database         string `env:"MONGODB_DATABASE"`
}

type mongoDbRepository struct {
	database *mongo.Database
}

func NewMongoDbRepository(ctx context.Context, config *Config) (*mongoDbRepository, error) {
	mongoDbRepository := &mongoDbRepository{}
	return mongoDbRepository, mongoDbRepository.init(ctx, config)
}

func (repo *mongoDbRepository) init(ctx context.Context, config *Config) error {
	client, err := mongo.Connect(options.Client().ApplyURI(config.ConnectionString))
	if err != nil {
		return err
	}

	repo.database = client.Database(config.Database)
	return repo.ensureIndexes(ctx)
}

func (repo *mongoDbRepository) FindFilmkritiken(ctx context.Context, filmkritikenId string) (*filmkritiken.Filmkritiken, error) {
	mongoFilter := bson.M{"_id": bson.M{"$eq": filmkritikenId}}
	result := &filmkritiken.Filmkritiken{}

	err := repo.database.Collection(filmkritikenCollectionName).FindOne(ctx, mongoFilter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.NewNotFoundErrorFromString("Filmkritiken konnten nicht gefunden werden.")
		}

		return nil, err
	}

	return result, nil
}

func (repo *mongoDbRepository) GetFilmkritiken(ctx context.Context, filter *filmkritiken.FilmkritikenFilter) ([]*filmkritiken.Filmkritiken, int64, error) {
	mongoFilter := bson.D{}

	search := ""
	if filter != nil {
		search = filter.Suche
		if search == "" {
			search = filter.Titel
		}
	}

	if search != "" {
		escaped := regexp.QuoteMeta(search)
		mongoFilter = append(mongoFilter, bson.E{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "film.titel", Value: bson.D{{Key: "$regex", Value: escaped}, {Key: "$options", Value: "i"}}}},
				bson.D{{Key: "film.originaltitel", Value: bson.D{{Key: "$regex", Value: escaped}, {Key: "$options", Value: "i"}}}},
			},
		})
	}

	if filter != nil && filter.Jahr > 0 {
		startOfYear := time.Date(filter.Jahr, 1, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := time.Date(filter.Jahr, 12, 31, 23, 59, 59, 999999999, time.UTC)
		mongoFilter = append(mongoFilter, bson.E{
			Key: "details.besprochenam",
			Value: bson.D{
				{Key: "$gte", Value: startOfYear},
				{Key: "$lte", Value: endOfYear},
			},
		})
	}

	if filter != nil && filter.BeitragVon != "" {
		escaped := regexp.QuoteMeta(filter.BeitragVon)
		mongoFilter = append(mongoFilter, bson.E{
			Key: "details.beitragvon",
			Value: bson.D{
				{Key: "$regex", Value: "^" + escaped + "$"},
				{Key: "$options", Value: "i"},
			},
		})
	}

	totalCount, err := repo.database.Collection(filmkritikenCollectionName).CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, err
	}

	if filter != nil && filter.Sortierung == "beste" {
		pipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: mongoFilter}},
			bson.D{{Key: "$addFields", Value: bson.D{
				{Key: "avgRating", Value: bson.D{
					{Key: "$avg", Value: "$bewertungen.wertung"},
				}},
			}}},
			bson.D{{Key: "$sort", Value: bson.D{
				{Key: "avgRating", Value: -1},
				{Key: "details.besprochenam", Value: -1},
			}}},
		}
		if filter.Offset > 0 {
			pipeline = append(pipeline, bson.D{{Key: "$skip", Value: int64(filter.Offset)}})
		}
		if filter.Limit > 0 {
			pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(filter.Limit)}})
		}

		cursor, err := repo.database.Collection(filmkritikenCollectionName).Aggregate(ctx, pipeline)
		if err != nil {
			return nil, 0, err
		}
		results := make([]*filmkritiken.Filmkritiken, 0)
		err = cursor.All(ctx, &results)
		if err != nil {
			return nil, 0, err
		}
		return results, totalCount, nil
	}

	sortDir := -1
	if filter != nil && filter.Sortierung == "aelteste" {
		sortDir = 1
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "details.besprochenam", Value: sortDir}})

	if filter != nil {
		if filter.Limit > 0 {
			findOptions.SetLimit(int64(filter.Limit))
		}
		if filter.Offset > 0 {
			findOptions.SetSkip(int64(filter.Offset))
		}
	}

	cursor, err := repo.database.Collection(filmkritikenCollectionName).Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	results := make([]*filmkritiken.Filmkritiken, 0)

	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

func (repo *mongoDbRepository) SaveImage(ctx context.Context, imageBites *[]byte) (string, error) {
	id := bson.NewObjectID().Hex()

	image := &image{
		ImageId: id,
		Image:   imageBites,
	}

	filter := bson.M{"_id": bson.M{"$eq": image.ImageId}}
	update := bson.D{bson.E{Key: "$set", Value: image}}
	_, err := repo.database.Collection(imagesCollectionName).UpdateOne(ctx, filter, update, updateOpts)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (repo *mongoDbRepository) FindImage(ctx context.Context, imageId string) (*[]byte, error) {
	mongoFilter := bson.M{"_id": bson.M{"$eq": imageId}}
	result := &image{}

	err := repo.database.Collection(imagesCollectionName).FindOne(ctx, mongoFilter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.NewNotFoundErrorFromString("Bild konnte nicht gefunden werden.")
		}

		return nil, err
	}

	return result.Image, nil
}

func (repo *mongoDbRepository) DeleteImage(ctx context.Context, imageId string) error {

	filter := bson.M{"_id": bson.M{"$eq": imageId}}
	_, err := repo.database.Collection(imagesCollectionName).DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	return nil
}

func (repo *mongoDbRepository) SaveFilmkritiken(ctx context.Context, filmkritiken *filmkritiken.Filmkritiken) error {

	if filmkritiken.Id == "" {
		filmkritiken.Id = bson.NewObjectID().Hex()
	}

	filter := bson.M{"_id": bson.M{"$eq": filmkritiken.Id}}
	update := bson.D{bson.E{Key: "$set", Value: filmkritiken}}
	_, err := repo.database.Collection(filmkritikenCollectionName).UpdateOne(ctx, filter, update, updateOpts)

	if err != nil {
		return err
	}

	return nil
}

func (repo *mongoDbRepository) UpdateBesprochenAm(ctx context.Context, filmkritikenId string, besprochenAm time.Time) error {
	filter := bson.M{"_id": bson.M{"$eq": filmkritikenId}}
	update := bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: "details.besprochenam", Value: besprochenAm}}}}
	result, err := repo.database.Collection(filmkritikenCollectionName).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.NewNotFoundErrorFromString("Filmkritiken konnten nicht gefunden werden.")
	}
	return nil
}

func (repo *mongoDbRepository) GetFilterOptions(ctx context.Context) (*filmkritiken.FilterOptions, error) {
	yearsPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "details.besprochenam", Value: bson.D{{Key: "$ne", Value: nil}}}}}},
		bson.D{{Key: "$project", Value: bson.D{{Key: "year", Value: bson.D{{Key: "$year", Value: "$details.besprochenam"}}}}}},
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$year"}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
	}

	yearsCursor, err := repo.database.Collection(filmkritikenCollectionName).Aggregate(ctx, yearsPipeline)
	if err != nil {
		return nil, err
	}
	type yearResult struct {
		ID int `bson:"_id"`
	}
	var yearResults []yearResult
	if err := yearsCursor.All(ctx, &yearResults); err != nil {
		return nil, err
	}

	jahre := make([]int, 0, len(yearResults))
	for _, yr := range yearResults {
		if yr.ID > 0 {
			jahre = append(jahre, yr.ID)
		}
	}

	contribPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "details.beitragvon", Value: bson.D{{Key: "$ne", Value: nil}, {Key: "$ne", Value: ""}}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$details.beitragvon"}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	contribCursor, err := repo.database.Collection(filmkritikenCollectionName).Aggregate(ctx, contribPipeline)
	if err != nil {
		return nil, err
	}
	type contribResult struct {
		ID string `bson:"_id"`
	}
	var contribResults []contribResult
	if err := contribCursor.All(ctx, &contribResults); err != nil {
		return nil, err
	}

	beitragende := make([]string, 0, len(contribResults))
	for _, c := range contribResults {
		if c.ID != "" {
			beitragende = append(beitragende, c.ID)
		}
	}

	return &filmkritiken.FilterOptions{
		Jahre:       jahre,
		Beitragende: beitragende,
	}, nil
}
