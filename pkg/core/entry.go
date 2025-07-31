package core

import (
	"fmt"
	"net/http"

	cli "github.com/excitedplus1s/docker-tar/pkg/cli"
	chinadns "github.com/excitedplus1s/gfwutils/dns"
	chinahttp "github.com/excitedplus1s/gfwutils/http"
)

type Runner interface {
	Initialize(*EntryPoint)
	InitializeCheck()
	Run() error
}

func Run(r Runner) error {
	r.InitializeCheck()
	return r.Run()
}

func FRun(r Runner) func() error {
	return func() error {
		return Run(r)
	}
}

func Run00(r Runner, f func()) {
	r.InitializeCheck()
	f()
}

func FRun00(r Runner, f func()) func() error {
	return func() error {
		Run00(r, f)
		return nil
	}
}

func Run01[R1 any](r Runner, f func() R1) R1 {
	r.InitializeCheck()
	return f()
}

func FRun01(r Runner, f func() error) func() error {
	return func() error {
		return Run01(r, f)
	}
}

func Run10[A1 any](r Runner, f func(A1), arg1 A1) {
	r.InitializeCheck()
	f(arg1)
}

func FRun10[A1 any](r Runner, f func(A1), arg1 A1) func() error {
	return func() error {
		Run10(r, f, arg1)
		return nil
	}
}

func Run11[A1, R1 any](r Runner, f func(A1) R1, arg1 A1) R1 {
	r.InitializeCheck()
	return f(arg1)
}

func FRun11[A1, R1 any](r Runner, f func(A1) error, arg1 A1) func() error {
	return func() error {
		return Run11(r, f, arg1)
	}
}

func RunLoopWithPrintln(fns []func() error) {
	for _, fn := range fns {
		err := fn()
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func Run12[A1, R1, R2 any](r Runner, f func(A1) (R1, R2), arg1 A1) (R1, R2) {
	r.InitializeCheck()
	return f(arg1)
}

type HttpClientFn func() *http.Client

type EntryPoint struct {
	HttpClientFnPtr *HttpClientFn

	Authenticator          *Authenticator
	ImageInfoManager       *ImageInfoManager
	RequestInfoManager     *RequestInfoManager
	ImageIndexFetcher      *ImageIndexFetcher
	OutputFileManager      *OutputFileManager
	ImageConfigFetcher     *ImageConfigFetcher
	ImageConfigBlobFetcher *ImageConfigBlobFetcher
	ImageContentCollector  *ImageContentCollector
	LayerDownloader        *LayerDownloader
}

func (s *EntryPoint) ApplyConfig(config *cli.Config) error {
	if config == nil {
		return fmt.Errorf("imageConfigBlobFetcher: WriteToFile Failed, EntryPoint object is nil")
	}
	var dnsConfig *chinadns.Config
	var httpConfig *chinahttp.Config
	var resolverCache = map[string][]string{}
	if config.ExperimentalEnabled() {
		dnsConfig = &chinadns.Config{
			Network:    config.Network(),
			Timeout:    config.DNSQueryTimeout(),
			ServerList: config.DNSServerList(),
		}
		httpConfig = &chinahttp.Config{
			ResolverCache:     &resolverCache,
			AnticensorEnabled: true,
		}
	}
	var httpClientFn HttpClientFn = func() *http.Client {
		return chinahttp.Client(dnsConfig, httpConfig)
	}
	s.HttpClientFnPtr = &httpClientFn
	s.Authenticator = new(Authenticator)
	s.ImageInfoManager = new(ImageInfoManager)
	s.RequestInfoManager = new(RequestInfoManager)
	s.ImageIndexFetcher = new(ImageIndexFetcher)
	s.OutputFileManager = new(OutputFileManager)
	s.ImageConfigFetcher = new(ImageConfigFetcher)
	s.ImageConfigBlobFetcher = new(ImageConfigBlobFetcher)
	s.ImageContentCollector = new(ImageContentCollector)
	s.LayerDownloader = new(LayerDownloader)
	var initializes = []Runner{
		s.Authenticator,
		s.ImageInfoManager,
		s.RequestInfoManager,
		s.ImageIndexFetcher,
		s.OutputFileManager,
		s.ImageConfigFetcher,
		s.ImageConfigBlobFetcher,
		s.ImageContentCollector,
		s.LayerDownloader,
	}
	for _, init := range initializes {
		init.Initialize(s)
	}
	s.ImageInfoManager.ApplyConfig(config)
	s.RequestInfoManager.ApplyConfig(config)
	s.OutputFileManager.ApplyConfig(config)
	return nil
}

func listArchAction(config *cli.Config) {
	entry := &EntryPoint{}
	entry.ApplyConfig(config)
	listFns := []func() error{FRun(entry.Authenticator),
		FRun(entry.ImageIndexFetcher),
	}
	RunLoopWithPrintln(listFns)
	availableArch := Run01(entry.ImageIndexFetcher, entry.ImageIndexFetcher.AvailableArch)
	fmt.Println("Available Architecture:")
	for _, arch := range availableArch {
		fmt.Println(arch)
	}
}

func pullAction(config *cli.Config) {
	entry := &EntryPoint{}
	entry.ApplyConfig(config)
	pullFns := []func() error{FRun(entry.Authenticator),
		FRun(entry.ImageIndexFetcher),
		FRun(entry.ImageConfigFetcher),
		FRun(entry.ImageConfigBlobFetcher),
		FRun(entry.ImageContentCollector),
		FRun(entry.OutputFileManager),
		FRun(entry.LayerDownloader),
		FRun01(entry.OutputFileManager, entry.OutputFileManager.ChtimesAll),
		FRun01(entry.OutputFileManager, entry.OutputFileManager.TarImage),
	}
	RunLoopWithPrintln(pullFns)
}

func Eval(config *cli.Config) {
	action := config.Action()
	switch action {
	case "pull":
		pullAction(config)
	case "list":
		listArchAction(config)
	default:
		fmt.Println("Action not support:", action)
	}
}
