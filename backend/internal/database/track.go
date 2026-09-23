package database

import (
	"context"
	"database/sql"
	"echora/internal/models"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrackStore struct {
	db *pgxpool.Pool
}

func NewTrackStore(db *pgxpool.Pool) *TrackStore {
	return &TrackStore{db: db}
}

func (s *TrackStore) GetUserTracks(ctx context.Context, userID string) ([]models.Track, error) {
	query := `
		SELECT id, user_id, title, cover_key, object_key, created_at
		FROM tracks
		WHERE user_id = $1
		ORDER BY created_at DESC;
	`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []models.Track

	for rows.Next() {
		var t models.Track
		err := rows.Scan(&t.ID, &t.User_id, &t.Title, &t.Cover_key, &t.Object_key, &t.Created_at)
		if err != nil {
			return nil, err
		}

		tracks = append(tracks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (s *TrackStore) Create(ctx context.Context, input models.TrackCreateInput) (*models.Track, error) {
	var track models.Track

	query := `
		INSERT INTO tracks 
		(user_id, title, cover_key, object_key) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, user_id, title, cover_key, object_key, created_at;
	`

	err := s.db.QueryRow(ctx, query, input.User_id, input.Title, input.Cover_key, input.Object_key).Scan(
		&track.ID,
		&track.User_id,
		&track.Title,
		&track.Cover_key,
		&track.Object_key,
		&track.Created_at,
	)
	if err != nil {
		return nil, err
	}

	return &track, nil
}

func (s *TrackStore) GetTrackByID(ctx context.Context, id int, userID string) (*models.Track, error) {
	var track models.Track

	query := `
		SELECT id, user_id, title, cover_key, object_key, created_at 
		FROM tracks
		WHERE id = $1 AND user_id = $2;
	`

	err := s.db.QueryRow(ctx, query, id, userID).Scan(
		&track.ID,
		&track.User_id,
		&track.Title,
		&track.Cover_key,
		&track.Object_key,
		&track.Created_at,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &track, nil
}
