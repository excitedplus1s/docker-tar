package core

import (
	"fmt"
	"net/url"

	cli "github.com/excitedplus1s/docker-tar/pkg/cli"
)

type RequestInfoManager struct {
	registryEndpoint string
	imageInfo        *ImageInfoManager

	initialized bool
}

func (req *RequestInfoManager) Initialize(entry *EntryPoint) {
	if entry == nil {
		panic("RequestInfoManager init failed, EntryPoint is nil")
	}
	if entry.ImageInfoManager == nil {
		panic("RequestInfoManager init failed, EntryPoint's ImageInfoManager is nil")
	}
	req.imageInfo = entry.ImageInfoManager
	req.initialized = true
}

func (req *RequestInfoManager) InitializeCheck() {
	if req.initialized {
		return
	}
	panic("RequestInfoManager not init")
}

func (req *RequestInfoManager) Run() error {
	return nil
}

func (req *RequestInfoManager) ApplyConfig(config *cli.Config) error {
	if config == nil {
		return fmt.Errorf("requestInfoManager: ApplyConfig Failed, Config object is nil")
	}
	registryEndpoint := config.MirrorRegistry()

	if len(registryEndpoint) == 0 {
		u := &url.URL{
			Scheme: "https",
			Host:   req.imageInfo.Registry(),
		}
		req.registryEndpoint = u.String()
	} else {
		parsedURL, err := url.Parse(registryEndpoint)
		if err != nil {
			return err
		}
		if len(parsedURL.Scheme) == 0 {
			u := &url.URL{
				Scheme: "https",
				Host:   registryEndpoint,
			}
			req.registryEndpoint = u.String()
		} else {
			req.registryEndpoint = registryEndpoint
		}
	}
	return nil
}

func (req *RequestInfoManager) RegistryEndpoint() string {
	return req.registryEndpoint
}

func (req *RequestInfoManager) Registry() (string, error) {
	if req.imageInfo == nil {
		return "<nil>", fmt.Errorf("requestInfoManager: Get Registry Failed, ImageInfoManager not init")
	}
	return req.imageInfo.Registry(), nil
}

func (req *RequestInfoManager) UserName() string {
	return req.imageInfo.UserName()
}

func (req *RequestInfoManager) Password() string {
	return req.imageInfo.Password()
}

func (req *RequestInfoManager) Repository() string {
	return req.imageInfo.Repository()
}

func (req *RequestInfoManager) ImageName() string {
	return req.imageInfo.ImageName()
}

func (req *RequestInfoManager) Tag() string {
	return req.imageInfo.Tag()
}

func (req *RequestInfoManager) Architecture() string {
	return req.imageInfo.Architecture()
}
