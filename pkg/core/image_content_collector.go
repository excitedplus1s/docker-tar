package core

import (
	"encoding/json"
	"time"

	moby "github.com/excitedplus1s/spec-go/moby"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/identity"
)

type ImageContentCollector struct {
	manifestJson     []byte
	repositoriesJson []byte
	blobSumV1        map[string]string
	v1Jsons          map[string][]byte
	v1IDs            []string

	imageInfo       *ImageInfoManager
	imageConfig     *ImageConfigFetcher
	imageConfigBlob *ImageConfigBlobFetcher
	outputFileInfo  *OutputFileManager
	initialized     bool
}

func (gen *ImageContentCollector) Initialize(entry *EntryPoint) {
	if entry == nil {
		panic("ImageContentCollector init failed, EntryPoint is nil")
	}
	if entry.ImageInfoManager == nil {
		panic("ImageContentCollector init failed, EntryPoint's ImageInfoManager is nil")
	}
	if entry.ImageConfigFetcher == nil {
		panic("ImageContentCollector init failed, EntryPoint's ImageConfigFetcher is nil")
	}
	if entry.ImageConfigBlobFetcher == nil {
		panic("ImageContentCollector init failed, EntryPoint's ImageConfigBlobFetcher is nil")
	}
	if entry.OutputFileManager == nil {
		panic("ImageContentCollector init failed, EntryPoint's outputFileInfo is nil")
	}
	gen.imageInfo = entry.ImageInfoManager
	gen.imageConfig = entry.ImageConfigFetcher
	gen.imageConfigBlob = entry.ImageConfigBlobFetcher
	gen.outputFileInfo = entry.OutputFileManager
	gen.v1Jsons = map[string][]byte{}
	gen.blobSumV1 = map[string]string{}
	gen.initialized = true
}

func (gen *ImageContentCollector) InitializeCheck() {
	if gen.initialized {
		return
	}
	panic("ImageContentCollector not init")
}

func (gen *ImageContentCollector) Run() error {
	imageConfigBlob := gen.imageConfigBlob
	layerIDs := identity.ChainIDs(imageConfigBlob.DiffIDs())
	var parent digest.Digest
	var lastV1Image moby.V1Image
	if err := json.Unmarshal(imageConfigBlob.Content(), &lastV1Image); err != nil {
		return err
	}
	v1IDs := make([]digest.Digest, len(layerIDs))
	imageConfig := gen.imageConfig
	blobDigests := imageConfig.BlobDigests()
	for index, layerID := range layerIDs {
		v1ImgCreated := time.Unix(0, 0).UTC()
		v1Image := moby.V1Image{
			Created: &v1ImgCreated,
		}
		if index == len(layerIDs)-1 {
			v1Image = lastV1Image
		}
		v1ID, err := moby.CreateID(v1Image, layerID, parent)
		if err != nil {
			return err
		}
		v1Image.OS = imageConfigBlob.OS()
		v1Image.ID = v1ID.Encoded()
		if parent != "" {
			v1Image.Parent = parent.Encoded()
		}
		v1JSON, err := json.Marshal(v1Image)
		if err != nil {
			return err
		}
		gen.v1Jsons[v1ID.Encoded()] = v1JSON
		parent = v1ID
		v1IDs[index] = v1ID
		gen.v1IDs = append(gen.v1IDs, v1ID.Encoded())
		gen.blobSumV1[blobDigests[index].Encoded()] = v1ID.Encoded()
	}
	type Summary struct {
		Config   string   `json:"Config"`
		RepoTags []string `json:"RepoTags"`
		Layers   []string `json:"Layers"`
	}
	imageInfo := gen.imageInfo
	repoTag := imageInfo.FullName()

	repoTags := append([]string{}, repoTag)
	layers := []string{}

	for _, v1ID := range v1IDs {
		layers = append(layers, v1ID.Encoded()+"/layer.tar")
	}

	summary := Summary{
		Config:   imageConfig.ConfigDigest().Encoded() + ".json",
		RepoTags: repoTags,
		Layers:   layers,
	}

	summarys := []Summary{}
	summarys = append(summarys, summary)

	manifestJson, err := json.Marshal(summarys)
	if err != nil {
		return err
	}
	gen.manifestJson = manifestJson

	versionMap := map[string]string{
		imageInfo.Tag(): v1IDs[len(v1IDs)-1].Encoded(),
	}
	repositories := map[string]map[string]string{
		imageInfo.FullNameWithoutTag(): versionMap,
	}
	repositoriesJson, err := json.Marshal(repositories)
	if err != nil {
		return err
	}
	gen.repositoriesJson = repositoriesJson
	return nil
}

func (gen *ImageContentCollector) WriteToFile() error {
	fmanifest, err := gen.outputFileInfo.ManifestFD()
	if err != nil {
		return err
	}
	defer fmanifest.Close()
	_, err = fmanifest.Write(gen.manifestJson)
	if err != nil {
		return err
	}
	frepositories, err := gen.outputFileInfo.RepositoriesFD()
	if err != nil {
		return err
	}
	defer frepositories.Close()
	_, err = frepositories.Write(gen.repositoriesJson)
	if err != nil {
		return err
	}
	for v1ID, data := range gen.v1Jsons {
		fjson, err := gen.outputFileInfo.LayerJsonFDByV1Id(v1ID)
		if err != nil {
			return err
		}
		defer fjson.Close()
		_, err = fjson.Write(data)
		if err != nil {
			return err
		}
		fversion, err := gen.outputFileInfo.LayerVersionFDByV1Id(v1ID)
		if err != nil {
			return err
		}
		defer fversion.Close()
		_, err = fversion.WriteString("1.0")
		if err != nil {
			return err
		}
	}
	return nil
}

func (gen *ImageContentCollector) V1ID(blobSum string) (string, bool) {
	sum, ok := gen.blobSumV1[blobSum]
	return sum, ok
}

func (gen *ImageContentCollector) V1IDs() []string {
	return gen.v1IDs
}
