package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type ImageConfigFetcher struct {
	configDigest       digest.Digest
	blobDigestWithType map[digest.Digest]string
	blobDigestWithSize map[digest.Digest]int64
	blobDigests        []digest.Digest

	authenticator    *Authenticator
	requestInfo      *RequestInfoManager
	imageIndex       *ImageIndexFetcher
	httpClientCreate HttpClientFn
	initialized      bool
}

func (config *ImageConfigFetcher) Initialize(entry *EntryPoint) {
	if entry == nil {
		panic("ImageConfigFetcher init failed, EntryPoint is nil")
	}
	if entry.RequestInfoManager == nil {
		panic("ImageConfigFetcher init failed, EntryPoint's RequestInfoManager is nil")
	}
	if entry.Authenticator == nil {
		panic("ImageConfigFetcher init failed, EntryPoint's Authenticator is nil")
	}
	if entry.ImageIndexFetcher == nil {
		panic("ImageConfigFetcher init failed, EntryPoint's ImageIndexFetcher is nil")
	}
	if entry.HttpClientFnPtr == nil {
		panic("ImageConfigFetcher init failed, EntryPoint's httpClientFnPtr is nil")
	}
	config.authenticator = entry.Authenticator
	config.requestInfo = entry.RequestInfoManager
	config.imageIndex = entry.ImageIndexFetcher
	config.httpClientCreate = *entry.HttpClientFnPtr
	config.initialized = true
}

func (config *ImageConfigFetcher) InitializeCheck() {
	if config.initialized {
		return
	}
	panic("ImageConfigFetcher not init")
}

func (config *ImageConfigFetcher) Run() error {
	client := config.httpClientCreate()
	requestInfo := config.requestInfo
	arch := requestInfo.Architecture()
	imageIndex := config.imageIndex
	archDigest, ok := imageIndex.SelectDigestByArchitecture(arch)
	if !ok {
		return fmt.Errorf("no atchitecture found")
	}
	manifestURL := fmt.Sprintf("%s/v2/%s/%s/manifests/%s",
		requestInfo.RegistryEndpoint(),
		requestInfo.Repository(),
		requestInfo.ImageName(),
		archDigest)
	req, err := http.NewRequest(http.MethodGet, manifestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set(HeaderAccept, v1.MediaTypeImageManifest)
	authenticator := config.authenticator
	authenticator.Authorize(req)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get image config failed, %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var manifest v1.Manifest
	err = json.Unmarshal(body, &manifest)
	if err != nil {
		return err
	}
	config.configDigest = manifest.Config.Digest
	config.blobDigestWithType = map[digest.Digest]string{}
	config.blobDigestWithSize = map[digest.Digest]int64{}
	config.blobDigests = make([]digest.Digest, len(manifest.Layers))
	for index, layer := range manifest.Layers {
		config.blobDigestWithType[layer.Digest] = layer.MediaType
		config.blobDigestWithSize[layer.Digest] = layer.Size
		config.blobDigests[index] = layer.Digest
	}
	return nil
}

func (config *ImageConfigFetcher) ConfigDigest() digest.Digest {
	return config.configDigest
}

func (config *ImageConfigFetcher) BlobDigestWithType() map[digest.Digest]string {
	return config.blobDigestWithType
}

func (config *ImageConfigFetcher) BlobDigestSize(d digest.Digest) int64 {
	return config.blobDigestWithSize[d]
}

func (config *ImageConfigFetcher) BlobDigests() []digest.Digest {
	return config.blobDigests
}
