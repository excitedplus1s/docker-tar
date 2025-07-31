package core

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/klauspost/compress/zstd"
	"github.com/schollz/progressbar/v3"
)

type LayerDownloader struct {
	authenticator    *Authenticator
	requestInfo      *RequestInfoManager
	imageInfo        *ImageInfoManager
	imageConfig      *ImageConfigFetcher
	outputFileInfo   *OutputFileManager
	httpClientCreate HttpClientFn
	initialized      bool
}

func (layer *LayerDownloader) Initialize(entry *EntryPoint) {
	if entry == nil {
		panic("LayerDownloader init failed, EntryPoint is nil")
	}
	if entry.Authenticator == nil {
		panic("LayerDownloader init failed, EntryPoint's Authenticator is nil")
	}
	if entry.RequestInfoManager == nil {
		panic("LayerDownloader init failed, EntryPoint's RequestInfoManager is nil")
	}
	if entry.ImageInfoManager == nil {
		panic("LayerDownloader init failed, EntryPoint's ImageInfoManager is nil")
	}
	if entry.ImageConfigFetcher == nil {
		panic("LayerDownloader init failed, EntryPoint's ImageConfigFetcher is nil")
	}
	if entry.OutputFileManager == nil {
		panic("LayerDownloader init failed, EntryPoint's outputFileInfo is nil")
	}
	if entry.HttpClientFnPtr == nil {
		panic("LayerDownloader init failed, EntryPoint's httpClientFnPtr is nil")
	}
	layer.authenticator = entry.Authenticator
	layer.requestInfo = entry.RequestInfoManager
	layer.imageInfo = entry.ImageInfoManager
	layer.imageConfig = entry.ImageConfigFetcher
	layer.outputFileInfo = entry.OutputFileManager
	layer.httpClientCreate = *entry.HttpClientFnPtr
	layer.initialized = true
}

func (layer *LayerDownloader) InitializeCheck() {
	if layer.initialized {
		return
	}
	panic("LayerDownloader not init")
}

func (layer *LayerDownloader) Run() error {
	client := layer.httpClientCreate()
	imageinfo := layer.imageInfo
	outputFileInfo := layer.outputFileInfo
	fmt.Println("Pulling from ", imageinfo.FullName())
	imageConfig := layer.imageConfig
	blobDigestWithType := imageConfig.BlobDigestWithType()
	totalDownload := len(blobDigestWithType)
	requestInfo := layer.requestInfo
	authenticator := layer.authenticator
	index := 0
	for blobDigest, mediaType := range blobDigestWithType {
		// Should HEAD first,but I don't want do it (:
		layerBlobURL := fmt.Sprintf("%s/v2/%s/%s/blobs/%s",
			requestInfo.RegistryEndpoint(),
			requestInfo.Repository(),
			requestInfo.ImageName(),
			blobDigest)
		req, err := http.NewRequest(http.MethodGet, layerBlobURL, nil)
		if err != nil {
			return err
		}
		authenticator.Authorize(req)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		prefixMessage := fmt.Sprintf("[%d/%d]%s: Downloading ",
			index+1,
			totalDownload,
			blobDigest.Encoded()[:12])
		bar := progressbar.DefaultBytes(
			imageConfig.BlobDigestSize(blobDigest),
			prefixMessage,
		)
		writeFileName, err := outputFileInfo.LayerFileNameByBlobSum(blobDigest.Encoded())
		if err != nil {
			return err
		}
		downloadFileName := writeFileName + ".download"
		fw, err := os.Create(downloadFileName)
		if err != nil {
			return err
		}
		_, err = io.Copy(io.MultiWriter(fw, bar), resp.Body)
		if err != nil {
			fw.Close()
			return err
		}
		layerType := string(mediaType[len(mediaType)-4:])
		switch layerType {
		case ".tar":
			fw.Close()
			os.Rename(downloadFileName, writeFileName)
		case "gzip":
			_, err = fw.Seek(0, io.SeekStart)
			if err != nil {
				fw.Close()
				return err
			}
			gr, err := gzip.NewReader(fw)
			if err != nil {
				fw.Close()
				return err
			}
			defer gr.Close()
			tw, err := outputFileInfo.LayerFDByBlobSum(blobDigest.Encoded())
			if err != nil {
				fw.Close()
				return err
			}
			defer tw.Close()
			_, err = io.Copy(tw, gr)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fw.Close()
		case "zstd":
			_, err = fw.Seek(0, io.SeekStart)
			if err != nil {
				fw.Close()
				return err
			}
			gr, err := zstd.NewReader(fw)
			if err != nil {
				fw.Close()
				return err
			}
			defer gr.Close()
			tw, err := outputFileInfo.LayerFDByBlobSum(blobDigest.Encoded())
			if err != nil {
				fw.Close()
				return err
			}
			defer tw.Close()
			_, err = io.Copy(tw, gr)
			if err != nil {
				fw.Close()
				return err
			}
			fw.Close()
		default:
			fw.Close()
			return fmt.Errorf("layer mediaType %s not support now", mediaType)
		}
		os.Remove(downloadFileName)
		index++
	}
	return nil
}
