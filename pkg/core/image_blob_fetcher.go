package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type ImageConfigBlobFetcher struct {
	blobContent []byte
	blobImage   v1.Image

	authenticator    *Authenticator
	requestInfo      *RequestInfoManager
	imageInfo        *ImageInfoManager
	imageConfig      *ImageConfigFetcher
	outputFileInfo   *OutputFileManager
	httpClientCreate HttpClientFn
	initialized      bool
}

func (blob *ImageConfigBlobFetcher) Initialize(entry *EntryPoint) {
	if entry == nil {
		panic("ImageConfigBlobFetcher init failed, EntryPoint is nil")
	}
	if entry.RequestInfoManager == nil {
		panic("ImageConfigBlobFetcher init failed, EntryPoint's RequestInfoManager is nil")
	}
	if entry.Authenticator == nil {
		panic("ImageConfigBlobFetcher init failed, EntryPoint's Authenticator is nil")
	}
	if entry.ImageInfoManager == nil {
		panic("ImageConfigBlobFetcher init failed, EntryPoint's ImageInfoManager is nil")
	}
	if entry.ImageConfigFetcher == nil {
		panic("ImageConfigBlobFetcher init failed, EntryPoint's ImageConfigFetcher is nil")
	}
	if entry.OutputFileManager == nil {
		panic("ImageConfigBlobFetcher init failed, EntryPoint's outputFileInfo is nil")
	}
	if entry.HttpClientFnPtr == nil {
		panic("ImageConfigBlobFetcher init failed, EntryPoint's httpClientFnPtr is nil")
	}
	blob.authenticator = entry.Authenticator
	blob.imageInfo = entry.ImageInfoManager
	blob.imageConfig = entry.ImageConfigFetcher
	blob.requestInfo = entry.RequestInfoManager
	blob.outputFileInfo = entry.OutputFileManager
	blob.httpClientCreate = *entry.HttpClientFnPtr
	blob.initialized = true
}

func (blob *ImageConfigBlobFetcher) InitializeCheck() {
	if blob.initialized {
		return
	}
	panic("ImageConfigBlobFetcher not init")
}

func (blob *ImageConfigBlobFetcher) Run() error {
	client := blob.httpClientCreate()
	requestInfo := blob.requestInfo
	imageConfig := blob.imageConfig
	configDigest := imageConfig.ConfigDigest()
	blobURL := fmt.Sprintf("%s/v2/%s/%s/blobs/%s",
		requestInfo.RegistryEndpoint(),
		requestInfo.Repository(),
		requestInfo.ImageName(),
		configDigest)
	req, err := http.NewRequest(http.MethodGet, blobURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set(HeaderAccept, v1.MediaTypeImageConfig)
	authenticator := blob.authenticator
	authenticator.Authorize(req)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get image config blobs failed, %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &blob.blobImage)
	if err != nil {
		return err
	}
	blob.blobContent = make([]byte, len(body))
	copy(blob.blobContent, body)
	return nil
}

func (blob *ImageConfigBlobFetcher) Content() []byte {
	return blob.blobContent[:]
}

func (blob *ImageConfigBlobFetcher) DiffIDs() []digest.Digest {
	return blob.blobImage.RootFS.DiffIDs[:]
}

func (blob *ImageConfigBlobFetcher) WriteTo(dst io.Writer) (int64, error) {
	reader := bytes.NewReader(blob.blobContent)
	return io.Copy(dst, reader)
}

func (blob *ImageConfigBlobFetcher) WriteToFile() error {
	fconfig, err := blob.outputFileInfo.ConfigFD()
	if err != nil {
		return err
	}
	defer fconfig.Close()
	_, err = fconfig.Write(blob.blobContent)
	if err != nil {
		return err
	}
	return nil
}

func (blob *ImageConfigBlobFetcher) UTC0Time() time.Time {
	ts := time.Unix(0, 0).UTC()
	return ts
}

func (blob *ImageConfigBlobFetcher) CreatedTime() time.Time {
	ts := blob.UTC0Time()
	if blob.blobImage.Created != nil {
		ts = *blob.blobImage.Created
	}
	return ts
}

func (blob *ImageConfigBlobFetcher) OS() string {
	return blob.blobImage.OS
}

func (blob *ImageConfigBlobFetcher) ConfigDigest() string {
	imageConfig := blob.imageConfig
	configDigest := imageConfig.ConfigDigest()
	return configDigest.Encoded()
}
