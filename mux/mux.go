package mux

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	muxgo "github.com/muxinc/mux-go"
)

// Load environment variables
func init() {
	godotenv.Load()
}

// Create a new Mux client
func NewMuxClient() *muxgo.APIClient {
	client := muxgo.NewAPIClient(
		muxgo.NewConfiguration(
			muxgo.WithBasicAuth(os.Getenv("MUX_TOKEN_ID"), os.Getenv("MUX_TOKEN_SECRET")),
		))

	return client
}

// Create a Mux asset
func CreateAsset(client *muxgo.APIClient, filename string) (muxgo.AssetResponse, error) {
	asset, err := client.AssetsApi.CreateAsset(muxgo.CreateAssetRequest{
		Input: []muxgo.InputSettings{
			{
				Url: "https://lostsonstv.sfo3.digitaloceanspaces.com/" + filename,
			},
		},
		PlaybackPolicy: []muxgo.PlaybackPolicy{"PUBLIC"},
	})
	if err != nil {
		return asset, err
	}

	return asset, nil
}

// Delete a Mux asset
func DeleteAsset(client *muxgo.APIClient, assetID string) error {
	err := client.AssetsApi.DeleteAsset(assetID)
	if err != nil {
		return err
	}

	return nil
}

func ReceiveVideoStatus(w http.ResponseWriter, r *http.Request) {

	// Get the hook data
	jsonResult, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	var hook VideoAsset
	json.Unmarshal([]byte(jsonResult), &hook)

	log.Printf("Incoming webhook from: %s\n", r.RemoteAddr)

	if hook.Type == "video.asset.ready" {
		log.Println(hook.Type)
		log.Println("Upload ID: ", hook.Data.UploadID)

		dbString := os.Getenv("DB_CONN_STR")

		conn, err := pgx.Connect(context.Background(), dbString)
		if err != nil {
			fmt.Println("unable to connecto to db: ", err)
		}

		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "update clips SET playback_id = $1 WHERE unique_id = $2", hook.Data.PlaybackIds[0].ID, hook.Data.UploadID)
		if err != nil {
			fmt.Println("unable to update db: ", err)
		}
	}
}
