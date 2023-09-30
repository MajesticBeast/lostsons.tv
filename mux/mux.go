package mux

import (
	"os"

	muxgo "github.com/muxinc/mux-go"
)

// Create a new Mux client
func NewMuxClient() *muxgo.APIClient {
	client := muxgo.NewAPIClient(
		muxgo.NewConfiguration(
			muxgo.WithBasicAuth(os.Getenv("MUX_TOKEN_ID"), os.Getenv("MUX_TOKEN_SECRET")),
		))

	return client
}

// CreateAsset creates a new Mux asset request
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
