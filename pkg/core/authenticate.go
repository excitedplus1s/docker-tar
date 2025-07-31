package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

type Authenticator struct {
	token string

	requestInfo      *RequestInfoManager
	httpClientCreate HttpClientFn
	initialized      bool
}

func (auth *Authenticator) Initialize(entry *EntryPoint) {
	if entry == nil {
		panic("Authenticator init failed, EntryPoint is nil")
	}
	if entry.RequestInfoManager == nil {
		panic("Authenticator init failed, EntryPoint's RequestInfoManager is nil")
	}
	if entry.HttpClientFnPtr == nil {
		panic("Authenticator init failed, EntryPoint's httpClientFnPtr is nil")
	}
	auth.requestInfo = entry.RequestInfoManager
	auth.httpClientCreate = *entry.HttpClientFnPtr
	auth.initialized = true
}

func (auth *Authenticator) InitializeCheck() {
	if auth.initialized {
		return
	}
	panic("Authenticator not init")
}

func (auth *Authenticator) Run() error {
	return auth.Challenge()
}

func (auth *Authenticator) Challenge() error {
	client := auth.httpClientCreate()
	challengeURL := fmt.Sprintf("%s/v2/", auth.requestInfo.RegistryEndpoint())
	req, err := http.NewRequest(http.MethodGet, challengeURL, nil)
	if err != nil {
		return err
	}
	requestInfo := auth.requestInfo
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var realm string
	var service string
	if resp.StatusCode == http.StatusUnauthorized {
		// Copy from azure
		// matches challenges having quoted parameters, capturing scheme and parameters
		challenge := regexp.MustCompile(`(?:(\w+) ((?:\w+="[^"]*",?\s*)+))`)
		// captures parameter names and values in a match of the above expression
		challengeParams := regexp.MustCompile(`(\w+)="([^"]*)"`)
		// WWW-Authenticate can have multiple values, each containing multiple challenges
		for _, h := range resp.Header.Values(HeaderWWWAuthenticate) {
			for _, sm := range challenge.FindAllStringSubmatch(h, -1) {
				// sm is [challenge, scheme, params] (see regexp documentation on submatches)
				for _, sm := range challengeParams.FindAllStringSubmatch(sm[2], -1) {
					// sm is [key="value", key, value] (see regexp documentation on submatches)
					if sm[1] == "realm" {
						realm = sm[2]
					}
					if sm[1] == "service" {
						service = sm[2]
					}
				}
			}
		}
	} else {
		return fmt.Errorf("challage failed, %s", resp.Status)
	}
	if len(service) == 0 || len(realm) == 0 {
		return fmt.Errorf("auth endpoint not found")
	}
	baseURL := realm
	params := url.Values{}
	params.Add("scope", fmt.Sprintf(
		"repository:%s/%s:pull",
		requestInfo.Repository(),
		requestInfo.ImageName()))
	params.Add("service", service)
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()
	req, err = http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	username := requestInfo.UserName()
	password := requestInfo.Password()
	if len(username) > 0 && len(password) > 0 {
		loginToken := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		req.Header.Set(HeaderAuthorization, BasicTokenPrefix+loginToken)
	}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		type Token struct {
			Token        string `json:"token"`
			AccessToken  string `json:"access_token,omitempty"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		}
		var token Token
		if err := json.Unmarshal(body, &token); err != nil {
			return err
		}

		auth.token = token.Token
	} else {
		return fmt.Errorf("get token failed, %s", resp.Status)
	}
	return nil
}

func (auth *Authenticator) Authorize(req *http.Request) error {
	if req == nil {
		return fmt.Errorf("request object is nil")
	}

	if len(auth.token) > 0 {
		req.Header.Set(HeaderAuthorization, BearerTokenPrefix+auth.token)
	}
	return nil
}
