package core

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	cli "github.com/excitedplus1s/docker-tar/pkg/cli"
)

type FileDescriptor struct {
	Name           string
	CreateTime     time.Time
	LastModifyTime time.Time
}

type OutputFileManager struct {
	outputFile     string
	downloadFloder string

	imageGenerateContent *ImageContentCollector
	imageConfigBlob      *ImageConfigBlobFetcher
	initialized          bool

	manifest      FileDescriptor
	config        FileDescriptor
	repositories  FileDescriptor
	layerFloders  []FileDescriptor
	layerJsons    map[string]FileDescriptor
	layerVersions map[string]FileDescriptor
	layers        map[string]FileDescriptor
}

func (out *OutputFileManager) Initialize(entry *EntryPoint) {
	if entry == nil {
		panic("LayerDownloader init failed, EntryPoint is nil")
	}
	if entry.ImageConfigBlobFetcher == nil {
		panic("LayerDownloader init failed, EntryPoint's ImageConfigFetcher is nil")
	}
	if entry.ImageContentCollector == nil {
		panic("LayerDownloader init failed, EntryPoint's ImageContentCollector is nil")
	}
	out.imageGenerateContent = entry.ImageContentCollector
	out.imageConfigBlob = entry.ImageConfigBlobFetcher
	out.initialized = true
}

func (out *OutputFileManager) InitializeCheck() {
	if out.initialized {
		return
	}
	panic("OutputFileManager not init")
}

func (out *OutputFileManager) Run() error {
	out.manifest = FileDescriptor{
		Name:           filepath.Join(out.DownloadFloder(), "manifest.json"),
		CreateTime:     out.imageConfigBlob.UTC0Time(),
		LastModifyTime: out.imageConfigBlob.UTC0Time(),
	}
	out.repositories = FileDescriptor{
		Name:           filepath.Join(out.DownloadFloder(), "repositories"),
		CreateTime:     out.imageConfigBlob.UTC0Time(),
		LastModifyTime: out.imageConfigBlob.UTC0Time(),
	}
	out.config = FileDescriptor{
		Name:           filepath.Join(out.DownloadFloder(), out.imageConfigBlob.ConfigDigest()+".json"),
		CreateTime:     out.imageConfigBlob.CreatedTime(),
		LastModifyTime: out.imageConfigBlob.CreatedTime(),
	}
	v1IDs := out.imageGenerateContent.V1IDs()
	out.layerFloders = make([]FileDescriptor, len(v1IDs))
	out.layerJsons = map[string]FileDescriptor{}
	out.layerVersions = map[string]FileDescriptor{}
	out.layers = map[string]FileDescriptor{}
	for i, v1ID := range v1IDs {
		out.layerFloders[i] = FileDescriptor{
			Name:           filepath.Join(out.DownloadFloder(), v1ID),
			CreateTime:     out.imageConfigBlob.CreatedTime(),
			LastModifyTime: out.imageConfigBlob.CreatedTime(),
		}
		out.layerJsons[v1ID] = FileDescriptor{
			Name:           filepath.Join(out.layerFloders[i].Name, "json"),
			CreateTime:     out.imageConfigBlob.CreatedTime(),
			LastModifyTime: out.imageConfigBlob.CreatedTime(),
		}
		out.layerVersions[v1ID] = FileDescriptor{
			Name:           filepath.Join(out.layerFloders[i].Name, "VERSION"),
			CreateTime:     out.imageConfigBlob.CreatedTime(),
			LastModifyTime: out.imageConfigBlob.CreatedTime(),
		}
		out.layers[v1ID] = FileDescriptor{
			Name:           filepath.Join(out.layerFloders[i].Name, "layer.tar"),
			CreateTime:     out.imageConfigBlob.CreatedTime(),
			LastModifyTime: out.imageConfigBlob.CreatedTime(),
		}
	}
	if err := out.CreateFloder(); err != nil {
		return err
	}
	if err := out.imageConfigBlob.WriteToFile(); err != nil {
		return err
	}
	if err := out.imageGenerateContent.WriteToFile(); err != nil {
		return err
	}
	return nil
}

func (out *OutputFileManager) ChtimesAll() error {
	manifest := out.manifest
	if err := os.Chtimes(manifest.Name, manifest.CreateTime, manifest.LastModifyTime); err != nil {
		return err
	}
	repositories := out.repositories
	if err := os.Chtimes(repositories.Name, repositories.CreateTime, repositories.LastModifyTime); err != nil {
		return err
	}
	config := out.config
	if err := os.Chtimes(config.Name, config.CreateTime, config.LastModifyTime); err != nil {
		return err
	}
	for _, layerJson := range out.layerJsons {
		if err := os.Chtimes(layerJson.Name, layerJson.CreateTime, layerJson.LastModifyTime); err != nil {
			return err
		}
	}
	for _, layerVersion := range out.layerVersions {
		if err := os.Chtimes(layerVersion.Name, layerVersion.CreateTime, layerVersion.LastModifyTime); err != nil {
			return err
		}
	}
	for _, layer := range out.layers {
		if err := os.Chtimes(layer.Name, layer.CreateTime, layer.LastModifyTime); err != nil {
			return err
		}
	}
	for _, layerFloder := range out.layerFloders {
		if err := os.Chtimes(layerFloder.Name, layerFloder.CreateTime, layerFloder.LastModifyTime); err != nil {
			return err
		}
	}
	return nil
}

func (out *OutputFileManager) ManifestFD() (*os.File, error) {
	return os.Create(out.manifest.Name)
}

func (out *OutputFileManager) RepositoriesFD() (*os.File, error) {
	return os.Create(out.repositories.Name)
}

func (out *OutputFileManager) ConfigFD() (*os.File, error) {
	return os.Create(out.config.Name)
}

func (out *OutputFileManager) LayerJsonFDByV1Id(v1id string) (*os.File, error) {
	fd, ok := out.layerJsons[v1id]
	if !ok {
		return nil, fmt.Errorf("%s json info not found", v1id)
	}
	return os.Create(fd.Name)
}

func (out *OutputFileManager) LayerVersionFDByV1Id(v1id string) (*os.File, error) {
	fd, ok := out.layerVersions[v1id]
	if !ok {
		return nil, fmt.Errorf("%s version info not found", v1id)
	}
	return os.Create(fd.Name)
}

func (out *OutputFileManager) LayerFDByV1Id(v1id string) (*os.File, error) {
	fd, ok := out.layers[v1id]
	if !ok {
		return nil, fmt.Errorf("%s layer info not found", v1id)
	}
	return os.Create(fd.Name)
}

func (out *OutputFileManager) LayerFileNameByV1Id(v1id string) (string, error) {
	fd, ok := out.layers[v1id]
	if !ok {
		return "<nil>", fmt.Errorf("%s layer info not found", v1id)
	}
	return fd.Name, nil
}

func (out *OutputFileManager) LayerFileNameByBlobSum(blobsum string) (string, error) {
	v1id, ok := out.imageGenerateContent.V1ID(blobsum)
	if !ok {
		return "<nil>", fmt.Errorf("%s v1id info not found", blobsum)
	}
	return out.LayerFileNameByV1Id(v1id)
}

func (out *OutputFileManager) LayerFDByBlobSum(blobsum string) (*os.File, error) {
	v1id, ok := out.imageGenerateContent.V1ID(blobsum)
	if !ok {
		return nil, fmt.Errorf("%s v1id info not found", blobsum)
	}
	return out.LayerFDByV1Id(v1id)
}

func (out *OutputFileManager) ApplyConfig(config *cli.Config) error {
	if config == nil {
		return fmt.Errorf("outputFileManager: ApplyConfig Failed, Config object is nil")
	}
	outputFile := config.OutputFile()
	if len(outputFile) > 0 {
		out.outputFile = outputFile
	} else {
		out.outputFile = fmt.Sprintf("%d.tar",
			time.Now().Unix())
	}
	out.downloadFloder = fmt.Sprintf("%s.%d",
		out.outputFile,
		time.Now().Unix())
	return nil
}

func (out *OutputFileManager) OutputFile() string {
	return out.outputFile
}

func (out *OutputFileManager) DownloadFloder() string {
	return out.downloadFloder
}

func (out *OutputFileManager) CreateFloder() error {
	err := os.MkdirAll(out.downloadFloder, os.ModePerm)
	if err != nil {
		return err
	}
	for _, layerFloder := range out.layerFloders {
		err := os.MkdirAll(layerFloder.Name, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (out *OutputFileManager) TarImage() error {
	baseFloder := out.DownloadFloder()
	fw, err := os.Create(out.OutputFile())
	if err != nil {
		return err
	}
	defer fw.Close()
	tw := tar.NewWriter(fw)
	defer tw.Close()
	processFn := func(fileName string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		_, hdrName, ok := strings.Cut(strings.TrimPrefix(fileName, string(filepath.Separator)), string(filepath.Separator))
		hdrName = strings.Replace(hdrName, string(filepath.Separator), "/", -1)
		if !ok {
			return nil
		}
		hdr.Name = hdrName
		if err := tw.WriteHeader(hdr); err != nil {
			fmt.Println(err)
			return err
		}
		if !fi.Mode().IsRegular() {
			return nil
		}
		fr, err := os.Open(fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer fr.Close()
		_, err = io.Copy(tw, fr)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	}
	err = filepath.Walk(baseFloder, processFn)
	os.RemoveAll(baseFloder)
	fmt.Println("Output File: ", out.OutputFile())
	return err
}
