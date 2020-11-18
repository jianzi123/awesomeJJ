package main

import (
	"awesomeJJ/authmethod/jwt"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/browser"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"gopkg.in/resty.v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
	"os"
	"time"
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
	Code int `json:"code"`
	Msg string `json:"msg"`
}

func GetTokenFromAuthnServerHttp(ctx context.Context,
	localIP, user, ws string) ([]string, error) {
	outToken := []string{}
	// GET request
	resp, err := resty.R().SetQueryParam("node", localIP).
		SetQueryParam("user", user).
		SetResult(&AuthResponse{}).SetError(&RespErr{}).
		Get(fmt.Sprintf("%s/v1/tokens", ws))

	logrus.Errorf("resty.R().SetQueryParam %v %v", resp, err)
	if err != nil {
		return outToken, err
	}
	if resp.IsError() {
		logrus.Errorf("call authn rest api failed: %v", resp.Error())
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

func checkAuthnServerStatus(ws string) (bool, error) {

	isReady := false
	// GET request
	resp, err := resty.R().SetError(&RespErr{}).Get(fmt.Sprintf(
		"%s/v1/status", ws))
	if err != nil {
		return isReady, err
	}
	isReady = resp.IsSuccess()
	return isReady, nil
}

func (obj *AuthnClient) GetTokenFromAuthnServer(ctx context.Context, localIP, user, ws string,
	readyChan chan<- string) ([]string, error) {
	// 1. check authn server health for call browser
	statusReady := make(chan string)
	go func(ctx context.Context, readyChan chan<- string) {
		t := time.Tick(time.Second * 2)
		for {
			select {
			case <-ctx.Done():
				break
			case <-t:
				isExist, err := checkAuthnServerStatus(ws)
				if err != nil {
					logrus.Errorf("checkAuthnServerStatus failed: %v", err)
					continue
				}
				if isExist == true {
					readyChan <- "ok"
					return
				}
			}
		}
	}(ctx, statusReady)

	// after call browser get token from authn server
	tokenReady := make(chan []string)

	go func(ctx context.Context, ready chan<- []string) {
		t := time.Tick(time.Second * 2)
		for {
			select {
			case <-ctx.Done():
				break
			case <-t:
				token, err := GetTokenFromAuthnServerHttp(ctx, localIP, user, ws)
				if err != nil {
					logrus.Errorf("GetTokenFromAuthnServerHttp failed: %v", err)
					continue
				}
				if len(token) != 0 {
					ready <- token
				}
			}
		}
	}(ctx, tokenReady)
	for {
		select {
		case <-statusReady:
			readyChan <- "http://idaas-test.zhenguanyu.com/enduser/sp/sso/yfdjwt9?enterpriseId=yfd"
		case tokenStr := <-tokenReady:
			return tokenStr, nil
		case <-ctx.Done():
			logrus.Errorf("GetTokenFromAuthnServer timeout")
			return []string{}, fmt.Errorf("GetTokenFromAuthnServer timeout")
		}
	}

}

func GetTokenCacheFromServer(localIP, user, ws string) ([]string, error) {
	tokens := []string{}
	// GET request
	resp, err := resty.R().SetQueryParam("node", localIP).
		SetQueryParam("user", user).
		SetResult(&AuthResponse{}).SetError(&RespErr{}).
		Get(fmt.Sprintf("%s/v1/tokens", ws))
	logrus.Errorf("GetTokenCacheFromServer : %v \n %v", resp, err)
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
	logrus.Infof("++++ GetTokenCacheFromServer tokens: %v", tokens)
	return tokens, nil

}
func main() {
	// args
	timeoutFlag := flag.Int("timeout", 60000, "timeout")
	localIP := flag.String("addr", "127.0.0.1", "local ip addr")
	user := flag.String("user", "wangshuaijian", "username")
	webhookServer := flag.String("ws", "http://10.254.24.66:8899",
		"webhookServer")
	pk := flag.String("publickey", "wangshuaijian", "publickey")
	logPath := flag.String("logPath", "/Users/wangshuaijian/go_workspace/src/github.com/jianzi123/awesomeJJ/wsj.log", "logPath")
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)

	// log
	file, err := os.OpenFile(*logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil{
		return
	}
	logrus.SetOutput(file)
	defer file.Close()
	logrus.Debugf("%d", *timeoutFlag)

	tokenResult := ""
	// get cache from server
	// cache没拿到或者拿到的过期了，都要重新去拿
	tokens, err := GetTokenCacheFromServer(*localIP, *user, *webhookServer)
	if err == nil && len(tokens) != 0 && len(tokens[0]) != 0{
		tokenResult, err = CheckTokensValid(tokens, *pk)
		if err != nil{
			logrus.Errorf("CheckTokensValid failed: %v", err)
		}
	}
	if len(tokenResult) == 0{
		tokenCreate, err := GetToken(*localIP, *user, *webhookServer, *timeoutFlag)
		if err != nil {
			logrus.Errorf("get token from server failed: %v", err)
			return
		}
		tokenResult, err = CheckTokensValid(tokenCreate, *pk)
		if err != nil{
			logrus.Errorf("CheckTokensValid failed: %v", err)
			return
		}
	}

	authn, err := CreateClientAuthn(tokenResult)
	if err != nil {
		logrus.Errorf("CreateClientAuthn  failed: %v", err)
		return
	}
	logrus.Infof("client end token: %s \n authn: %s", tokenResult, authn)
	// todo:
	// check timeout
	fmt.Print(authn)
}

func CheckTokensValid(tokens []string, pk string) (string, error)  {
	tokenResult := ""
	for _, tokenItem := range tokens{
		tokenInfo, err := jwt.Parse(tokenItem, pk)
		if err != nil{
			continue
		}
		if tokenInfo.IsValid != true{
			continue
		}
		tokenResult = tokenItem
		break
	}
	return tokenResult, nil
}

func CreateClientAuthn(token string) (string, error) {

	if len(token) == 0{
		return "", fmt.Errorf("token content is empty")
	}

	out := v1beta1.ExecCredential{
		TypeMeta: v1.TypeMeta{
			APIVersion: "client.authentication.k8s.io/v1beta1",
			Kind:       "ExecCredential",
		},
		Spec:     v1beta1.ExecCredentialSpec{},
		Status: &v1beta1.ExecCredentialStatus{
			ExpirationTimestamp:   nil,
			Token:                 token,
			ClientCertificateData: "",
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

func GetToken(ip, user, ws string, timeout int) ([]string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second* time.Duration(timeout))
	defer cancel()
	readyChan := make(chan string)
	out := []string{}
	var eg errgroup.Group
	eg.Go(func() error {
		select {
		case url, ok := <-readyChan:

			if !ok {
				return nil
			}
			go func() {
				logrus.Infof("opening %s in the browser", url)
				if err := browser.OpenURL(url); err != nil {
					logrus.Infof(`error: could not open the browser: %s

Please visit the following URL in your browser manually: %s`, err, url)
					return
				}
				return
			}()
			return nil
		case <-ctx.Done():
			return xerrors.Errorf("context cancelled while waiting for the local server: %w", ctx.Err())
		}
	})
	eg.Go(func() error {
		defer close(readyChan)
		tokenSet, err := client.GetTokenFromAuthnServer(ctx, ip, user, ws, readyChan)
		if err != nil {
			return xerrors.Errorf("authorization code flow error: %w %s %s", err, ip, ws)
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
