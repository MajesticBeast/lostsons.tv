package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	CreateClip(Clip) error
	GetClip(string) (Clip, error)
	GetAllClips() ([]Clip, error)
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

func (s *PostgresStore) GetClip(id string) (Clip, error) {
	clip := Clip{}

	query := buildGetClipQuery()

	err := s.db.QueryRow(context.Background(), query, id).Scan(&clip.ID, &clip.PlaybackID, &clip.AssetID,
		&clip.DateUploaded, &clip.UserID, &clip.GameID,
		&clip.Description, &clip.Tags, &clip.FeaturedUsers, &clip.Game, &clip.Username,
	)

	if err != nil {
		err = fmt.Errorf("error running GetClip: %w", err)
		return clip, err
	}

	return clip, nil
}

func (s *PostgresStore) GetAllClips() ([]Clip, error) {
	clips := []Clip{}

	query := buildGetAllClipsQuery()

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		err = fmt.Errorf("error running GetAllClips: %w", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		clip := new(Clip)
		if err := rows.Scan(&clip.ID, &clip.PlaybackID, &clip.AssetID, &clip.DateUploaded, &clip.UserID, &clip.GameID, &clip.Description, &clip.Tags, &clip.FeaturedUsers, &clip.Game, &clip.Username); err != nil {
			fmt.Printf("Scanned values: ID=%v, PlaybackID=%v, AssetID=%v, DateUploaded=%v, UserID=%v, GameID=%v, Description=%v, Tags=%v, FeaturedUsers=%v, Game=%v, Username=%v\n",
				clip.ID, clip.PlaybackID, clip.AssetID, clip.DateUploaded, clip.UserID, clip.GameID, clip.Description, clip.Tags, clip.FeaturedUsers, clip.Game, clip.Username)
			err = fmt.Errorf("error scanning rows: %w", err)
			return nil, err
		}

		clips = append(clips, *clip)
	}
	return clips, nil
}

func (s *PostgresStore) CreateClip(clip Clip) error {

	//
	// SECTION: Clip insertion
	//
	// Check if user exists -> user_id
	user_id, err := s.getIDFromString(clip.Username, "users", "username")
	if err != nil {
		err = fmt.Errorf("error selecting user: %w", err)
		return err
	}

	// Check if game exists -> game_id
	game_id, err := s.getIDFromString(clip.Game, "games", "name")
	if err != nil {
		err = fmt.Errorf("error selecting game: %w", err)
		return err
	}

	// Game and user exists, complete the clip object and insert the clip
	clip.ID = uuid.New().String()
	clip.UserID = user_id
	clip.GameID = game_id
	insertClipQuery := `INSERT INTO clips (id, playback_id, asset_id, date_uploaded, description, user_id, game_id) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.Exec(context.Background(), insertClipQuery,
		clip.ID,
		clip.PlaybackID,
		clip.AssetID,
		clip.DateUploaded,
		clip.Description,
		clip.UserID,
		clip.GameID,
	)

	if err != nil {
		err = fmt.Errorf("error inserting clip: %w", err)
		return err
	}

	//
	// SECTION: Tags insertion
	//
	// Insert the tags into the tags table. If the tag already exists, do nothing.
	tags := strings.Split(clip.Tags, ",")

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)

		// Skip tag if it is just whitespace
		if len(tag) == 0 {
			continue
		}

		tagID := uuid.New().String()

		insertTagsQuery := `INSERT INTO tags (id, tag_name) VALUES ($1, $2) ON CONFLICT (tag_name) DO UPDATE SET tag_name = EXCLUDED.tag_name RETURNING id`
		err = s.db.QueryRow(context.Background(), insertTagsQuery, tagID, tag).Scan(&tagID)

		if err != nil {
			err = fmt.Errorf("error inserting tags: %w", err)
			return err
		}

		// Insert the clip_id and tag_id into the clips_tags table
		insertClipsTagsQuery := `INSERT INTO clips_tags (clip_id, tag_id) VALUES ($1, $2)`
		_, err = s.db.Exec(context.Background(), insertClipsTagsQuery,
			clip.ID,
			tagID,
		)

		if err != nil {
			err = fmt.Errorf("error inserting clips_tags: %w", err)
			return err
		}
	}

	//
	// SECTION: Clips_Users insertion
	//
	// Insert the clip_id and user_id into the clips_users table
	insertClipsUsersQuery := `INSERT INTO clips_users (clip_id, user_id) VALUES ($1, $2)`
	_, err = s.db.Exec(context.Background(), insertClipsUsersQuery,
		clip.ID,
		clip.UserID,
	)

	if err != nil {
		err = fmt.Errorf("error inserting clips_users: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStore) getIDFromString(name string, table string, column string) (string, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE %s = $1", table, column)

	var uuid string
	err := s.db.QueryRow(context.Background(), query, name).Scan(&uuid)
	if err != nil {
		err = fmt.Errorf("error selecting %s: %s", name, err)
		return "", err
	}

	return uuid, nil
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

// Enter a new user into database
func (s *PostgresStore) CreateUser(user User) error {
	user.ID = uuid.New().String()

	query := `INSERT INTO users (id, username, email) VALUES ($1, $2, $3)`
	_, err := s.db.Exec(context.Background(), query,
		user.ID,
		user.Username,
		user.Email,
	)

	if err != nil {
		err = fmt.Errorf("error inserting user: %w", err)
		return err
	}

	return nil
}

// Get list of all users
func (s *PostgresStore) GetAllUsers() ([]User, error) {
	users := []User{}

	query := `SELECT * FROM users`
	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		err = fmt.Errorf("error getting all users: %w", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := new(User)
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			err = fmt.Errorf("error scanning rows: %w", err)
			return nil, err
		}

		users = append(users, *user)
	}

	return users, nil
}

// Get a user by username
func (s *PostgresStore) GetUserByUsername(username string) (User, error) {
	user := User{}

	query := `SELECT * FROM users WHERE username = $1`
	err := s.db.QueryRow(context.Background(), query, username).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		err = fmt.Errorf("error getting user by username: %w", err)
		return user, err
	}

	return user, nil
}

// Get a user by email
func (s *PostgresStore) GetUserByEmail(email string) (User, error) {
	user := User{}

	query := `SELECT * FROM users WHERE email = $1`
	err := s.db.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		err = fmt.Errorf("error getting user by email: %w", err)
		return user, err
	}

	return user, nil
}

// Get list of all games
func (s *PostgresStore) GetAllGames() ([]Game, error) {
	games := []Game{}

	query := `SELECT * FROM games`
	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		err = fmt.Errorf("error getting all games: %w", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		game := new(Game)
		if err := rows.Scan(&game.ID, &game.Name); err != nil {
			err = fmt.Errorf("error scanning rows: %w", err)
			return nil, err
		}

		games = append(games, *game)
	}

	return games, nil
}

// Get a game by name
func (s *PostgresStore) GetGameByName(name string) (Game, error) {
	game := Game{}

	query := `SELECT * FROM games WHERE name = $1`
	err := s.db.QueryRow(context.Background(), query, name).Scan(&game.ID, &game.Name)
	if err != nil {
		err = fmt.Errorf("error getting game by name: %w", err)
		return game, err
	}

	return game, nil
}

// Enter a new game into database
func (s *PostgresStore) CreateGame(game Game) error {
	game.ID = uuid.New().String()

	query := `INSERT INTO games (id, name) VALUES ($1, $2)`
	_, err := s.db.Exec(context.Background(), query,
		game.ID,
		game.Name,
	)

	if err != nil {
		err = fmt.Errorf("error inserting game: %w", err)
		return err
	}

	return nil
}
