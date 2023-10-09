package main

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/majesticbeast/lostsons.tv/mux"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/chi/v5"
)

func (s *APIServer) clipsRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/new", makeHTTPHandleFunc(s.handleCreateClip))
	r.Post("/delete", makeHTTPHandleFunc(s.handleDeleteClip))

	return r
}

// Route for submitting a clip
func (s *APIServer) handleCreateClip(w http.ResponseWriter, r *http.Request) error {
	newForm := new(NewClipForm)
	if err := r.ParseMultipartForm(45 << 20); err != nil {
		err = fmt.Errorf("error parsing multipart form: %w", err)
		return err
	}

	newForm.Description = r.FormValue("description")
	newForm.Game = r.FormValue("game")
	newForm.Username = r.FormValue("username")
	newForm.Tags = r.FormValue("tags")
	newForm.FeaturedUsers = r.FormValue("featured_users")

	// Get file from form
	file, handler, err := r.FormFile("clip")
	if err != nil {
		err = fmt.Errorf("error getting file from form: %w", err)
		return err
	}
	defer file.Close()

	// Upload file to DigitalOcean Spaces
	sess, err := NewDigitalOceanSession()
	if err != nil {
		err = fmt.Errorf("error creating new digital ocean session: %w", err)
		return err
	}

	svc, err := NewS3Client(sess)
	if err != nil {
		err = fmt.Errorf("error creating new s3 client: %w", err)
		return err
	}

	err = UploadFileToSpaces(svc, file, handler)
	if err != nil {
		err = fmt.Errorf("error uploading file to spaces: %w", err)
		return err
	}

	// Create a new Mux client and asset
	client := mux.NewMuxClient()
	asset, err := mux.CreateAsset(client, handler.Filename)
	if err != nil {
		err = fmt.Errorf("error creating mux asset: %w", err)
		return err
	}

	// Create a new clip object
	clip := Clip{
		PlaybackID:    asset.Data.PlaybackIds[0].Id,
		AssetID:       asset.Data.Id,
		Description:   newForm.Description,
		Game:          newForm.Game,
		Username:      newForm.Username,
		Tags:          newForm.Tags,
		FeaturedUsers: newForm.FeaturedUsers,
		DateUploaded:  time.Now(),
	}

	// Add clip to database
	err = s.store.CreateClip(clip)
	if err != nil {
		err = fmt.Errorf("error creating clip: %w", err)
		return err
	}

	return responseWithJSON(w, http.StatusOK, "success")
}

// Route for deleting a clip
func (s *APIServer) handleDeleteClip(w http.ResponseWriter, r *http.Request) error {
	clipID := r.PostFormValue("id")
	clip := Clip{}

	// Get the clip data from the database
	clip, err := s.store.GetClip(clipID)
	if err != nil {
		return fmt.Errorf("clip does not exist")
	}

	// Delete clip from Mux
	client := mux.NewMuxClient()
	if err := mux.DeleteAsset(client, clip.AssetID); err != nil {
		return fmt.Errorf("error deleting mux asset: %w", err)
	}

	// Delete clip_id from clips_users
	if err := s.store.DeleteClipsUsersClipID(clipID); err != nil {
		return fmt.Errorf("error deleting clip_id from clips_users: %w", err)
	}

	// Delete clip_id from clips_tags
	if err := s.store.DeleteClipsTagsClipID(clipID); err != nil {
		return fmt.Errorf("error deleting clip_id from clips_tags: %w", err)
	}

	// Delete clip
	if err := s.store.DeleteClip(clip.ID); err != nil {
		return fmt.Errorf("error deleting clip: %w", err)
	}

	return responseWithJSON(w, http.StatusOK, "success")
}

func NewDigitalOceanSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String("https://sfo3.digitaloceanspaces.com"),
		S3ForcePathStyle: aws.Bool(false),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("DO_SPACES_KEY"),
			os.Getenv("DO_SPACES_SECRET"),
			"",
		),
	})
	if err != nil {
		err = fmt.Errorf("error creating new digital ocean session: %w", err)
		return nil, err
	}

	return sess, nil
}

func NewS3Client(sess *session.Session) (*s3.S3, error) {
	svc := s3.New(sess)
	return svc, nil
}

func UploadFileToSpaces(svc *s3.S3, file multipart.File, handler *multipart.FileHeader) error {
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("lostsonstv"),
		Key:    aws.String(handler.Filename),
		Body:   file,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		err = fmt.Errorf("error uploading file to spaces: %w", err)
		return err
	}

	return nil
}
