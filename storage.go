package main

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Storage interface {
	CreateClip() error
}

type PostgresStore struct {
	db *pgx.Conn
}

func NewPostgresStore(dbConnStr string) (*PostgresStore, error) {
	db, err := pgx.Connect(context.Background(), dbConnStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {
	err := s.installUUIDExtension()
	if err != nil {
		return err
	}

	err = s.createGamesTable()
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

func (s *PostgresStore) installUUIDExtension() error {
	query := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createClipsTable() error {
	query := `CREATE TABLE IF NOT EXISTS clips (
		id uuid DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
		playback_id varchar(200) UNIQUE NOT NULL,
		asset_id varchar(200) UNIQUE NOT NULL,
		date_uploaded timestamp NOT NULL,
		user_id uuid NOT NULL,
		game_id uuid NOT NULL,
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
		id uuid DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
		tag_name varchar(20) UNIQUE NOT NULL
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createClipsTagsTable() error {
	query := `CREATE TABLE IF NOT EXISTS clips_tags (
		clip_id uuid NOT NULL,
		tag_id uuid NOT NULL,
		CONSTRAINT fk_clip_id FOREIGN KEY (clip_id) REFERENCES clips(id),
		CONSTRAINT fk_tag_id FOREIGN KEY (tag_id) REFERENCES tags(id)	
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createUsersTable() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id uuid DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
		username varchar(35) UNIQUE NOT NULL,
		email varchar(60) UNIQUE NOT NULL
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createGamesTable() error {
	query := `CREATE TABLE IF NOT EXISTS games (
		id uuid DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
		name varchar(60) NOT NULL,
		PRIMARY KEY ("id")
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) createClipsUsersTable() error {
	query := `CREATE TABLE IF NOT EXISTS clips_users (
		clip_id uuid NOT NULL,
		user_id uuid NOT NULL,
		CONSTRAINT fk_clip_id FOREIGN KEY (clip_id) REFERENCES clips(id),
		CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id)
	)`
	_, err := s.db.Exec(context.Background(), query)
	return err
}
