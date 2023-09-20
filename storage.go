package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	CreateClip(Clip) error
}

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(dbConnStr string) (*PostgresStore, error) {
	db, err := pgxpool.New(context.Background(), dbConnStr)
	if err != nil {
		err = fmt.Errorf("error creating db conn pool: %w", err)
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) CreateClip(clip Clip) error {

	//
	// SECTION: Clip insertion
	//
	// Check if user exists
	selectUserQuery := `SELECT id FROM users WHERE users.username = $1`
	var user_id string
	err := s.db.QueryRow(context.Background(), selectUserQuery, clip.User).Scan(&user_id)
	if err != nil {
		err = fmt.Errorf("error selecting user: %w", err)
		return err
	}

	// User exists at this point, check if game exists
	selectGameQuery := `SELECT id FROM games WHERE games.name = $1`
	var game_id string
	err = s.db.QueryRow(context.Background(), selectGameQuery, clip.Game).Scan(&game_id)
	if err != nil {
		err = fmt.Errorf("error selecting game: %w", err)
		return err
	}

	// Game and user exists, complete the clip object and insert the clip
	clip.Id = uuid.New().String()
	clip.User = user_id
	clip.Game = game_id
	insertClipQuery := `INSERT INTO clips (id, playback_id, asset_id, date_uploaded, description, user_id, game_id) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.Exec(context.Background(), insertClipQuery,
		clip.Id,
		clip.Playback_id,
		clip.Asset_id,
		clip.Date_uploaded,
		clip.Description,
		clip.User,
		clip.Game,
	)

	if err != nil {
		err = fmt.Errorf("error inserting clip: %w", err)
		return err
	}

	//
	// SECTION: Tags insertion
	//
	// Insert the tags into the tags table. If the tag already exists, do nothing.
	tagId := uuid.New().String()

	insertTagsQuery := `INSERT INTO tags (id, tag_name) VALUES ($1, $2) ON CONFLICT (tag_name) DO UPDATE SET tag_name = EXCLUDED.tag_name RETURNING id`
	err = s.db.QueryRow(context.Background(), insertTagsQuery, tagId, "creepin").Scan(&tagId)

	if err != nil {
		err = fmt.Errorf("error inserting tags: %w", err)
		return err
	}

	// Insert the clip_id and tag_id into the clips_tags table
	insertClipsTagsQuery := `INSERT INTO clips_tags (clip_id, tag_id) VALUES ($1, $2)`
	_, err = s.db.Exec(context.Background(), insertClipsTagsQuery,
		clip.Id,
		tagId,
	)

	if err != nil {
		err = fmt.Errorf("error inserting clips_tags: %w", err)
		return err
	}

	//
	// SECTION: Clips_Users insertion
	//
	// Insert the clip_id and user_id into the clips_users table
	insertClipsUsersQuery := `INSERT INTO clips_users (clip_id, user_id) VALUES ($1, $2)`
	_, err = s.db.Exec(context.Background(), insertClipsUsersQuery,
		clip.Id,
		clip.User,
	)

	if err != nil {
		err = fmt.Errorf("error inserting clips_users: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) Init() error {

	err := s.createGamesTable()
	if err != nil {
		return err
	}

	err = s.createUsersTable()
	if err != nil {
		return err
	}

	err = s.createTagsTable()
	if err != nil {
		return err
	}

	err = s.createClipsTable()
	if err != nil {
		return err
	}

	err = s.createClipsTagsTable()
	if err != nil {
		return err
	}

	err = s.createClipsUsersTable()
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) createClipsTable() error {
	query := `CREATE TABLE IF NOT EXISTS clips (
		id varchar(128) UNIQUE NOT NULL,
		playback_id varchar(200) UNIQUE NOT NULL,
		asset_id varchar(200) UNIQUE NOT NULL,
		date_uploaded timestamp NOT NULL,
		user_id varchar(128) NOT NULL,
		game_id varchar(128) NOT NULL,
		description varchar(120) NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT fk_game_id FOREIGN KEY (game_id) REFERENCES games(id),
		CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id)
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createTagsTable() error {
	query := `CREATE TABLE IF NOT EXISTS tags (
		id varchar(128) UNIQUE NOT NULL,
		tag_name varchar(20) UNIQUE NOT NULL
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createClipsTagsTable() error {
	query := `CREATE TABLE IF NOT EXISTS clips_tags (
		clip_id varchar(128) NOT NULL,
		tag_id varchar(128) NOT NULL,
		CONSTRAINT fk_clip_id FOREIGN KEY (clip_id) REFERENCES clips(id),
		CONSTRAINT fk_tag_id FOREIGN KEY (tag_id) REFERENCES tags(id)	
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createUsersTable() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id varchar(128) UNIQUE NOT NULL,
		username varchar(35) UNIQUE NOT NULL,
		email varchar(60) UNIQUE NOT NULL
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createGamesTable() error {
	query := `CREATE TABLE IF NOT EXISTS games (
		id varchar(128) UNIQUE NOT NULL,
		name varchar(60) NOT NULL,
		PRIMARY KEY ("id")
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createClipsUsersTable() error {
	query := `CREATE TABLE IF NOT EXISTS clips_users (
		clip_id varchar(128) NOT NULL,
		user_id varchar(128) NOT NULL,
		CONSTRAINT fk_clip_id FOREIGN KEY (clip_id) REFERENCES clips(id),
		CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id)
	)`
	_, err := s.db.Exec(context.Background(), query)
	return err
}
