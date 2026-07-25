package mongo

import (
	"context"
	"time"

	"github.com/DerBlum/filmkritiken-backend/domain/errors"
	"github.com/DerBlum/filmkritiken-backend/domain/session"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	sessionsCollectionName = "sessions"
)

func (repo *mongoDbRepository) ensureIndexes(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	_, err := repo.database.Collection(sessionsCollectionName).Indexes().CreateOne(ctx, indexModel)
	return err
}

func (repo *mongoDbRepository) SaveSession(ctx context.Context, s *session.Session) error {
	if s.ID == "" {
		s.ID = bson.NewObjectID().Hex()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}

	hashedID := session.HashSessionID(s.ID)

	doc := bson.M{
		"_id":         hashedID,
		"name":        s.Name,
		"permissions": s.Permissions,
		"expiresAt":   s.ExpiresAt,
		"createdAt":   s.CreatedAt,
	}

	filter := bson.M{"_id": bson.M{"$eq": hashedID}}
	update := bson.D{bson.E{Key: "$set", Value: doc}}
	_, err := repo.database.Collection(sessionsCollectionName).UpdateOne(ctx, filter, update, updateOpts)
	if err != nil {
		return errors.NewRepositoryError(err)
	}

	return nil
}

func (repo *mongoDbRepository) FindSession(ctx context.Context, sessionID string) (*session.Session, error) {
	if sessionID == "" {
		return nil, errors.NewNotFoundErrorFromString("Session-ID ist leer.")
	}

	hashedID := session.HashSessionID(sessionID)
	mongoFilter := bson.M{"_id": bson.M{"$eq": hashedID}}
	result := &session.Session{}

	err := repo.database.Collection(sessionsCollectionName).FindOne(ctx, mongoFilter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.NewNotFoundErrorFromString("Session nicht gefunden.")
		}
		return nil, errors.NewRepositoryError(err)
	}

	if time.Now().After(result.ExpiresAt) {
		_ = repo.DeleteSession(ctx, sessionID)
		return nil, errors.NewNotFoundErrorFromString("Session ist abgelaufen.")
	}

	result.ID = sessionID
	return result, nil
}

func (repo *mongoDbRepository) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}

	hashedID := session.HashSessionID(sessionID)
	filter := bson.M{"_id": bson.M{"$eq": hashedID}}
	_, err := repo.database.Collection(sessionsCollectionName).DeleteOne(ctx, filter)
	if err != nil {
		return errors.NewRepositoryError(err)
	}

	return nil
}

func (repo *mongoDbRepository) RefreshSession(ctx context.Context, sessionID string, duration time.Duration) error {
	if sessionID == "" {
		return errors.NewNotFoundErrorFromString("Session-ID ist leer.")
	}

	hashedID := session.HashSessionID(sessionID)
	newExpiresAt := time.Now().Add(duration)
	filter := bson.M{"_id": bson.M{"$eq": hashedID}}
	update := bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: "expiresAt", Value: newExpiresAt}}}}

	result, err := repo.database.Collection(sessionsCollectionName).UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.NewRepositoryError(err)
	}

	if result.MatchedCount == 0 {
		return errors.NewNotFoundErrorFromString("Session für Refresh nicht gefunden.")
	}

	return nil
}
