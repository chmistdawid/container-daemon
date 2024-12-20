package cont_image

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

func DownloadImage() string {
	ctx := context.Background()
	repoName := "docker.io/library/nginx:latest" // Official NGINX Docker image
	outputDir := "./nginx-image"                 // Directory to save image layers

	// Configure an authenticated or anonymous client
	//var client auth.Client
	// Parse the repository reference
	repo, err := remote.NewRepository(repoName)
	if err != nil {
		log.Printf("Error creating repository: %v\n", err)
		return ""
	}
	//repo.Client = &client

	// Fetch the image manifest descriptor
	log.Printf("Fetching image information: %s\n", repoName)
	desc, err := repo.Resolve(ctx, "latest")
	if err != nil {
		log.Printf("Error resolving image: %v\n", err)
		return ""
	}

	// Pull the manifest
	store, err := oci.New("/tmp/oci-layout-root")

	storeOpts := credentials.StoreOptions{}
	credStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		panic(err)
	}
	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: credentials.Credential(credStore), // Use the credentials store
	}

	_, err = oras.Copy(ctx, repo, desc.Digest.String(), store, desc.Digest.String(), oras.DefaultCopyOptions)

	if err != nil {
		log.Printf("Error pulling manifest: %v\n", err)
		return ""
	}

	// Read the manifest from the store
	manifestBytes, err := store.Fetch(ctx, desc)
	if err != nil {
		log.Printf("Error fetching manifest: %v\n", err)
		return ""
	}

	// Unmarshal the manifest
	var manifest ocispec.Manifest
	manifestData, err := io.ReadAll(manifestBytes)
	fmt.Println(string(manifestData))
	if err != nil {
		log.Printf("Error reading manifest data: %v\n", err)
		return ""
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		log.Printf("Error unmarshaling manifest: %v\n", err)
		return ""
	}
	fmt.Println(manifest.Annotations)

	// Print manifest details
	log.Println("Manifest Layers:")
	for _, layer := range manifest.Layers {
		log.Printf("- MediaType: %s, Digest: %s, Size: %d\n", layer.MediaType, layer.Digest.String(), layer.Size)
	}

	log.Println("Image download completed successfully.")
	return outputDir
}
