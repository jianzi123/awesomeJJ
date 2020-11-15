package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/browser"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/clientauthentication"
	"time"

	//_ "github.com/jianzi123/bgp_agent/utils/log"
	//"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/clientcmd"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"gopkg.in/resty.v1"
)

var client AuthnClient

type AuthnClient struct {
	Server string `json:"server"`
	Port   string `json:"port"`
}

func NewAuthnClient(ipAdrr, port string) *AuthnClient {
	return &AuthnClient{
		Server: ipAdrr,
		Port:   port,
	}
}

type AuthResponse struct {
	ClientIP string   `json:"clientIP"`
	Token    []string `json:"token"`
}

type RespErr struct {
}

func GetTokenFromAuthnServerHttp(ctx context.Context, localIP string) ([]string, error) {
	outToken := []string{}
	// GET request
	resp, err := resty.R().SetResult(&AuthResponse{}).SetError(&RespErr{}).Get("http://localhost:8899/v1/token")
	if err != nil {
		return outToken, err
	}
	if resp.IsError() {
		logrus.Errorf("GetTokenFromAuthnServer: %v", resp.Error())
		return outToken, err
	}
	//result := resp.Result()
	if result, ok := resp.Result().(*AuthResponse); ok {
		ip := result.ClientIP
		if ip != localIP {
			return outToken, fmt.Errorf("ip is not equal server: %s local IP: %s",
				ip, localIP)
		}
		outToken = result.Token
	}
	return outToken, nil
}

func checkAuthnServerStatus() (bool, error) {
	isReady := false
	// GET request
	resp, err := resty.R().SetError(&RespErr{}).Get("http://localhost:8899/v1/status")
	if err != nil {
		return isReady, err
	}
	isReady = resp.IsSuccess()
	return isReady, nil
}

func (obj *AuthnClient) GetTokenFromAuthnServer(ctx context.Context, localIP string,
	readyChan chan<- string) (string, error) {
	// 1. check authn server health for call browser
	statusReady := make(chan string)
	go func(ctx context.Context, readyChan chan<- string) {
		t := time.Tick(time.Second * 2)
		for {
			select {
			case <-ctx.Done():
				break
			case <-t:
				isExist, err := checkAuthnServerStatus()
				if err != nil {
					logrus.Errorf("checkAuthnServerStatus failed: %v", err)
					continue
				}
				if isExist == true {
					readyChan <- "ok"
				}
			}
		}
	}(ctx, statusReady)

	// after call browser get token from authn server
	tokenReady := make(chan string)

	go func(ctx context.Context, ready chan<- string) {
		t := time.Tick(time.Second * 2)
		for {
			select {
			case <-ctx.Done():
				break
			case <-t:
				token, err := GetTokenFromAuthnServerHttp(ctx, localIP)
				if err != nil {
					logrus.Errorf("GetTokenFromAuthnServerHttp failed: %v", err)
					continue
				}
				if len(token) != 0 {
					ready <- token[0]
				}
			}
		}
	}(ctx, tokenReady)
	for {
		select {
		case <-statusReady:
			readyChan <- "ready"
		case tokenStr := <-tokenReady:
			return tokenStr, nil
		case <-ctx.Done():
			logrus.Errorf("GetTokenFromAuthnServer timeout")
			return "", fmt.Errorf("GetTokenFromAuthnServer timeout")
		}
	}

}

func GetTokenCacheFromServer(localIP string) ([]string, error) {
	tokens := []string{}
	// GET request
	resp, err := resty.R().SetQueryParam("node", localIP).SetResult(&AuthResponse{}).SetError(&RespErr{}).
		Get("http://localhost:8899/v1/tokens")
	if err != nil {
		return tokens, err
	}
	if resp.IsError() {
		logrus.Errorf("GetTokenFromAuthnServer: %v", resp.Error())
		return tokens, err
	}
	//result := resp.Result()
	if result, ok := resp.Result().(*AuthResponse); ok {
		ip := result.ClientIP
		if ip != localIP {
			return tokens, fmt.Errorf("ip is not equal server: %s local IP: %s",
				ip, localIP)
		}
		tokens = result.Token
	}
	return tokens, nil

}
func main() {
	//flagSet := flag.NewFlagSet("calico-ipam", flag.ExitOnError)
	//timeoutFlag := flagSet.Int("timeout", 10, "timeout")
	//	logrus.Errorf("versionFlag upgradeFlag %d %v", timeoutFlag, upgradeFlag)
	//err := flagSet.Parse(os.Args[1:])

	timeoutFlag := flag.Int("timeout", 10, "timeout")
	localIP := flag.String("addr", "127.0.0.1", "local ip addr")
	flag.Parse()
	logrus.Debugf("%d", timeoutFlag)

	token := ""
	// get cache from server
	tokens, err := GetTokenCacheFromServer(*localIP)
	if err != nil {
		logrus.Errorf("get token cache from server failed: %v", err)
		token, err = GetToken(*localIP)
		if err != nil {
			logrus.Errorf("get token from server failed: %v", err)
			return
		}
	} else {
		if len(tokens) != 0 {
			token = tokens[0]
		} else {
			logrus.Errorf("get token cache from server failed: token num is zero")
			return
		}

	}
	authn, err := CreateClientAuthn(token)
	if err != nil {
		logrus.Errorf("CreateClientAuthn  failed: %v", err)
		return
	}
	fmt.Printf("%s", authn)
}

func CreateClientAuthn(token string) (string, error) {

	out := clientauthentication.ExecCredential{
		TypeMeta: v1.TypeMeta{},
		Spec:     clientauthentication.ExecCredentialSpec{},
		Status: &clientauthentication.ExecCredentialStatus{
			ExpirationTimestamp:   nil,
			Token:                 "token",
			ClientCertificateData: token,
			ClientKeyData:         "",
		},
	}
	buff, err := json.Marshal(&out)
	if err != nil {
		logrus.Infof("Marshal token failed: %v", err)
		return "", err
	}
	return string(buff), nil
}

func GetToken(ip string) (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	readyChan := make(chan string, 1)
	out := ""
	var eg errgroup.Group
	eg.Go(func() error {
		select {
		case url, ok := <-readyChan:
			if !ok {
				return nil
			}
			logrus.Infof("opening %s in the browser", url)
			if err := browser.OpenURL(url); err != nil {
				logrus.Infof(`error: could not open the browser: %s

Please visit the following URL in your browser manually: %s`, err, url)
				return nil
			}
			return nil
		case <-ctx.Done():
			return xerrors.Errorf("context cancelled while waiting for the local server: %w", ctx.Err())
		}
	})
	eg.Go(func() error {
		defer close(readyChan)
		tokenSet, err := client.GetTokenFromAuthnServer(ctx, ip, readyChan)
		if err != nil {
			return xerrors.Errorf("authorization code flow error: %w", err)
		}
		out = tokenSet
		logrus.Infof("got a token set by the authorization code flow")
		return nil
	})
	if err := eg.Wait(); err != nil {
		return out, xerrors.Errorf("authentication error: %w", err)
	}
	logrus.Infof("finished the authorization code flow via the browser")
	return out, nil
}
