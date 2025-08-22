package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type ImageIndexFetcher struct {
	architectureIndex map[string]digest.Digest

	authenticator    *Authenticator
	requestInfo      *RequestInfoManager
	httpClientCreate HttpClientFn
	initialized      bool
}

func (index *ImageIndexFetcher) Initialize(entry *EntryPoint) {
	if entry == nil {
		panic("ImageIndexFetcher init failed, EntryPoint is nil")
	}
	if entry.RequestInfoManager == nil {
		panic("ImageIndexFetcher init failed, EntryPoint's RequestInfoManager is nil")
	}
	if entry.Authenticator == nil {
		panic("ImageIndexFetcher init failed, EntryPoint's Authenticator is nil")
	}
	if entry.HttpClientFnPtr == nil {
		panic("ImageIndexFetcher init failed, EntryPoint's httpClientFnPtr is nil")
	}
	if index.architectureIndex == nil {
		index.architectureIndex = map[string]digest.Digest{}
	}
	index.authenticator = entry.Authenticator
	index.requestInfo = entry.RequestInfoManager
	index.httpClientCreate = *entry.HttpClientFnPtr
	index.initialized = true
}

func (index *ImageIndexFetcher) InitializeCheck() {
	if index.initialized {
		return
	}
	panic("ImageIndexFetcher not init")
}

func (index *ImageIndexFetcher) isSupportedIndexType(mt string) bool {
	const MediaTypeDockerSchema2ManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"
	switch mt {
	case v1.MediaTypeImageIndex, MediaTypeDockerSchema2ManifestList:
		return true
	default:
		return false
	}
}

func (index *ImageIndexFetcher) Run() error {
	client := index.httpClientCreate()
	requestInfo := index.requestInfo
	indexURL := fmt.Sprintf("%s/v2/%s/%s/manifests/%s",
		requestInfo.RegistryEndpoint(),
		requestInfo.Repository(),
		requestInfo.ImageName(),
		requestInfo.Tag())
	req, err := http.NewRequest(http.MethodGet, indexURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set(HeaderAccept, v1.MediaTypeImageIndex)
	authenticator := index.authenticator
	authenticator.Authorize(req)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get index info failed, %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var v1index v1.Index
	err = json.Unmarshal(body, &v1index)
	if err != nil {
		return err
	}
	respContentTpye := resp.Header.Get(HeaderContentType)
	if !index.isSupportedIndexType(v1index.MediaType) {
		return fmt.Errorf("%s is not support now,please let me know", string(respContentTpye))
	}
	for _, manifest := range v1index.Manifests {
		if manifest.Platform.OS == "linux" {
			arch := manifest.Platform.Architecture
			variant := manifest.Platform.Variant
			digest := manifest.Digest
			index.architectureIndex[arch+variant] = digest
		}
	}
	if len(index.architectureIndex) == 0 {
		return fmt.Errorf("no platform found in %s", string(body))
	}
	return nil
}

func (index *ImageIndexFetcher) AvailableArch() []string {
	result := make([]string, len(index.architectureIndex))
	i := 0
	for key, _ := range index.architectureIndex {
		result[i] = key
		i++
	}
	return result
}

func (index *ImageIndexFetcher) SelectDigestByArchitecture(arch string) (digest.Digest, bool) {
	dig, ok := index.architectureIndex[arch]
	return dig, ok
}
