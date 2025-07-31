package cli

const defaultArchitecture = "amd64"

type ExperimentalFeature struct {
	// IPv-Only,IPv6-Only or Dual
	Network         string
	DNSServerList   []string
	DNSQueryTimeout int
}

type Config struct {
	action         string
	outputFile     string
	imageInfo      string
	username       string
	password       string
	architecture   string
	mirrorRegistry string
	experimental   *ExperimentalFeature
}

func (c *Config) SetAction(action string) {
	c.action = action
}

func (c *Config) Action() string {
	return c.action
}

func (c *Config) SetOutputFile(outputFile string) {
	c.outputFile = outputFile
}

func (c *Config) OutputFile() string {
	return c.outputFile
}

func (c *Config) SetImageInfo(imageInfo string) {
	c.imageInfo = imageInfo
}

func (c *Config) ImageInfo() string {
	return c.imageInfo
}

func (c *Config) SetArchitecture(architecture string) {
	c.architecture = architecture
}

func (c *Config) Architecture() string {
	if len(c.architecture) == 0 {
		return defaultArchitecture
	}
	return c.architecture
}

func (c *Config) SetMirrorRegistry(mirrorRegistry string) {
	c.mirrorRegistry = mirrorRegistry
}

func (c *Config) MirrorRegistry() string {
	return c.mirrorRegistry
}

func (c *Config) ExperimentalEnabled() bool {
	return c.experimental != nil
}

func (c *Config) EnableExperimental() {
	if c.experimental == nil {
		c.experimental = new(ExperimentalFeature)
	}
}

func (c *Config) SetNetwork(network string) {
	if c.experimental != nil {
		c.experimental.Network = network
	}
}

func (c *Config) Network() string {
	var result string
	if c.experimental != nil {
		result = c.experimental.Network
	}
	return result
}

func (c *Config) SetDNSServerList(dnsServerList []string) {
	if c.experimental != nil {
		c.experimental.DNSServerList = dnsServerList
	}
}

func (c *Config) DNSServerList() []string {
	var result []string
	if c.experimental != nil {
		result = c.experimental.DNSServerList
	}
	return result
}

func (c *Config) SetDNSQueryTimeout(dnsQueryTimeout int) {
	if c.experimental != nil {
		c.experimental.DNSQueryTimeout = dnsQueryTimeout
	}
}

func (c *Config) SetUserNamePassword(username string, password string) {
	c.username = username
	c.password = password
}

func (c *Config) UserName() string {
	return c.username
}

func (c *Config) Password() string {
	return c.password
}

func (c *Config) DNSQueryTimeout() int {
	var result int
	if c.experimental != nil {
		result = c.experimental.DNSQueryTimeout
	}
	return result
}
