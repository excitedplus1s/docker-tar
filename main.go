package main

import (
	"flag"
	"os"
	"strings"

	cli "github.com/excitedplus1s/docker-tar/pkg/cli"
	core "github.com/excitedplus1s/docker-tar/pkg/core"
)

func main() {
	var action string
	flag.StringVar(&action, "action", "", "pull: this `action` will get the tar image.\n"+
		"list: this action will list the image available architecture")
	var image string
	flag.StringVar(&image, "image", "", "The `name` of the image you want to get. It should match what you entered in the docker CLI.")
	var username string
	flag.StringVar(&username, "username", "", "Set `username` if registry need login")
	var password string
	flag.StringVar(&password, "password", "", "Set `password` if registry need login")
	var architecture string
	flag.StringVar(&architecture, "arch", "amd64", "`architecture` of the image")
	var mirror string
	flag.StringVar(&mirror, "mirror", "", "Use mirror registry to download the image\n"+
		"You can use the original image name\n"+
		"In this way, the downloaded tar does not need to be re-tagged")
	var experimental bool
	flag.BoolVar(&experimental, "lab", false, "Use the experiment feature to help you download images(AntiCensorship)")
	var network string
	flag.StringVar(&network, "network", "ip",
		"This configuration takes effect when the experiment feature is on.\n"+
			"ip4: use `network` IPv4-Only \n"+
			"ip6: use network IPv6-Only\n"+
			"ip: use network dualstack")
	var dnsServers string
	flag.StringVar(&dnsServers, "dns-servers", "", "This configuration takes effect when the experiment feature is on.\n"+
		"The default DNS configuration `ip list` is built-in.\n"+
		"The input will be split by commas.")
	var output string
	flag.StringVar(&output, "output", "", "The `filename` where the tar image is stored.")
	var dnsTimeout int
	flag.IntVar(&dnsTimeout, "dns-timeout", 2, "This configuration takes effect when the experiment feature is on.")
	flag.Parse()
	if len(os.Args) < 5 {
		flag.PrintDefaults()
	}
	config := &cli.Config{}
	if experimental {
		config.EnableExperimental()
		if len(network) > 2 {
			config.SetNetwork(network)
		}
		dnsServerList := strings.Split(dnsServers, ",")
		if len(dnsServerList) > 0 {
			config.SetDNSServerList(dnsServerList)
		}
		config.SetDNSQueryTimeout(dnsTimeout)
	}

	config.SetAction(action)
	config.SetImageInfo(image)
	config.SetArchitecture(architecture)
	config.SetMirrorRegistry(mirror)
	config.SetOutputFile(output)
	config.SetUserNamePassword(username, password)
	core.Eval(config)
}
