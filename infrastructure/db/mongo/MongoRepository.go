package mongo

import (
	"context"
	"time"

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
	return nil
}

func (repo *mongoDbRepository) FindFilmkritiken(ctx context.Context, filmkritikenId string) (*filmkritiken.Filmkritiken, error) {
	mongoFilter := bson.M{"_id": bson.M{"$eq": filmkritikenId}}
	result := &filmkritiken.Filmkritiken{}

	err := repo.database.Collection(filmkritikenCollectionName).FindOne(ctx, mongoFilter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, filmkritiken.NewNotFoundErrorFromString("Filmkritiken konnten nicht gefunden werden.")
		}

		return nil, err
	}

	return result, nil
}

func (repo *mongoDbRepository) GetFilmkritiken(ctx context.Context, filter *filmkritiken.FilmkritikenFilter) ([]*filmkritiken.Filmkritiken, error) {
	mongoFilter := bson.D{}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "details.besprochenam", Value: -1}}).
		SetLimit(int64(filter.Limit)).
		SetSkip(int64(filter.Offset))
	cursor, err := repo.database.Collection(filmkritikenCollectionName).Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, err
	}
	results := make([]*filmkritiken.Filmkritiken, 0)

	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
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
			return nil, filmkritiken.NewNotFoundErrorFromString("Bild konnte nicht gefunden werden.")
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
		return filmkritiken.NewNotFoundErrorFromString("Filmkritiken konnten nicht gefunden werden.")
	}
	return nil
}
