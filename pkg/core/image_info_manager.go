package core

import (
	"fmt"
	"strings"

	cli "github.com/excitedplus1s/docker-tar/pkg/cli"
)

var defaultRegistry = "registry-1.docker.io"

type ImageInfoManager struct {
	registry     string
	repository   string
	imageName    string
	tag          string
	architecture string
	username     string
	password     string
}

func (info *ImageInfoManager) Initialize(*EntryPoint) {
}

func (info *ImageInfoManager) InitializeCheck() {
}

func (info *ImageInfoManager) Run() error {
	return nil
}

func (info *ImageInfoManager) ApplyConfig(config *cli.Config) error {
	if config == nil {
		return fmt.Errorf("imageInfoManager: ApplyConfig Failed, Config object is nil")
	}
	info.username = config.UserName()
	info.password = config.Password()
	info.architecture = config.Architecture()
	imageInfo := config.ImageInfo()
	parts := strings.Count(imageInfo, "/")
	cutImageTag := func(data string) (string, string) {
		image, tag, ok := strings.Cut(data, ":")
		if !ok {
			tag = "latest"
		}
		return image, tag
	}
	if parts < 2 {
		info.registry = defaultRegistry
		if parts == 0 {
			info.repository = "library"
			imageName, tag := cutImageTag(imageInfo)
			info.imageName = imageName
			info.tag = tag
		} else {
			repository, imageWithTag, _ := strings.Cut(imageInfo, "/")
			info.repository = repository
			imageName, tag := cutImageTag(imageWithTag)
			info.imageName = imageName
			info.tag = tag
		}
	} else {
		registry, suffix, _ := strings.Cut(imageInfo, "/")
		info.registry = registry
		repository, imageWithTag, _ := strings.Cut(suffix, "/")
		info.repository = repository
		imageName, tag := cutImageTag(imageWithTag)
		info.imageName = imageName
		info.tag = tag
	}
	return nil
}
func (info *ImageInfoManager) UserName() string {
	return info.username
}

func (info *ImageInfoManager) Password() string {
	return info.password
}

func (info *ImageInfoManager) Registry() string {
	return info.registry
}

func (info *ImageInfoManager) Repository() string {
	return info.repository
}

func (info *ImageInfoManager) ImageName() string {
	return info.imageName
}

func (info *ImageInfoManager) Tag() string {
	return info.tag
}

func (info *ImageInfoManager) Architecture() string {
	return info.architecture
}

func (info *ImageInfoManager) FullNameWithoutTag() string {
	if info.Registry() == defaultRegistry {
		if info.Repository() == "library" {
			return info.ImageName()
		} else {
			return fmt.Sprintf("%s/%s",
				info.Repository(),
				info.ImageName())
		}
	}
	return fmt.Sprintf("%s/%s/%s",
		info.Registry(),
		info.Repository(),
		info.ImageName())
}

func (info *ImageInfoManager) FullName() string {
	return fmt.Sprintf("%s:%s",
		info.FullNameWithoutTag(),
		info.Tag())
}
