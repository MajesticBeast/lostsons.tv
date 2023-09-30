package main

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

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

	// Create a new Mux client
	client := mux.NewMuxClient()

	// Create a Mux asset
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
	}

	// Add clip to database
	err = s.store.CreateClip(clip)
	if err != nil {
		err = fmt.Errorf("error creating clip: %w", err)
		return err
	}

	responseWithJSON(w, http.StatusOK, "success")

	return nil
}

// Create a function to set up a new digital ocean session
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

// Create a new s3 client
func NewS3Client(sess *session.Session) (*s3.S3, error) {
	svc := s3.New(sess)
	return svc, nil
}

// Upload file to DigitalOcean Spaces
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
