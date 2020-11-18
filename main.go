package main

import (
	"awesomeJJ/authmethod/jwt"
	"awesomeJJ/db"
	"awesomeJJ/pkg/k8s"
	"awesomeJJ/pkg/server"
	"context"
	"k8s.io/api/authentication/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	//"bgp_agent/utils/procinfo"
	//"bgp_agent/utils/rest"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/browser"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

const DefaultLocalServerSuccessHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Authorized</title>
	<script>
		function sleep (time) {
            return new Promise((resolve) => setTimeout(resolve, time));
        }
        sleep(5000).then(() => {
            window.close()
        })
	</script>
	<style>
		body {
			background-color: #eee;
			margin: 0;
			padding: 0;
			font-family: sans-serif;
		}
		.placeholder {
			margin: 2em;
			padding: 2em;
			background-color: #fff;
			border-radius: 1em;
		}
	</style>
</head>
<body>
	<div class="placeholder">
		<h1>Authorized</h1>
		<p>You can close this window.</p>
	</div>
</body>
</html>
`

var cache map[string]map[string]string

func init() {
	cache = map[string]map[string]string{}
}

type Browser struct{}

// Open opens the default browser.
func (*Browser) Open(url string) error {
	return browser.OpenURL(url)
}

func RequestPrint(ctx *gin.Context) {
	dump, err := httputil.DumpRequest(ctx.Request, true)
	logrus.Debugf("解析请求: %s; err: %v. \n", string(dump), err)

	// proxy 透传
	//requestIP := strings.Split(ctx.Request.RemoteAddr, ":")[0]
	//fmt.Println(requestIP)
	ctx.Next()
}

func GitInfo(ctx *gin.Context) {
	buff, err := ioutil.ReadFile("")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, string(buff))
}

func Login(ctx *gin.Context) {
	logrus.Infof("+++++++login")
	//ctx.JSON(http.StatusOK, DefaultLocalServerSuccessHTML)
	ctx.HTML(http.StatusOK, "/Users/wangshuaijian/go_workspace/src/github.com/jianzi123/awesomeJJ/browser/tmp.html",
		gin.H{})
}

func Status(ctx *gin.Context) {
	logrus.Infof("+++++++status")
	ctx.JSON(http.StatusOK, "ok")
}

type AuthResponse struct {
	ClientIP string   `json:"clientIP"`
	Token    []string `json:"token"`
}

type RespErr struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (obj *App)GetTokensByIP(ctx *gin.Context) {
	ip := ctx.Query("node")
	user := ctx.Query("user")
	if len(ip) == 0 || len(user) == 0{
		errRes := RespErr{
			Code: http.StatusBadRequest,
			Msg:  "should input node info",
		}
		ctx.JSON(http.StatusBadRequest, &errRes)
		return
	}
	// get token from cache
	//if _, ok := cache[ip]; ok != true {
	//	errRes := RespErr{
	//		Code: http.StatusForbidden,
	//		Msg:  fmt.Sprintf("node %s token not find", ip),
	//	}
	//	ctx.JSON(http.StatusForbidden, &errRes)
	//	return
	//}
	tokens := cache[ip]
	tokenList := []string{}
	for _, value := range tokens {
		tokenList = append(tokenList, value)
	}
	rkey := fmt.Sprintf("%s++++++%s", ip, user)
	rTokens, err := obj.RedisC.SMembers(ctx, rkey).Result()
	if err != nil{
		logrus.Errorf("Get token from redis failed: %v", err)
		errRes := RespErr{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("node %s token not find", ip),
		}
		ctx.JSON(http.StatusInternalServerError, &errRes)
		return
	}
	logrus.Infof("obj.RedisC.SMembers %v", rTokens)
	//for _, rToken := range rTokens{
	//
	//}
	resp := AuthResponse{
		ClientIP: ip,
		Token:    rTokens,
	}
	ctx.JSON(http.StatusOK, &resp)
}

func (obj *App) GetLogin(ctx *gin.Context) {

	// 1. parse token from request paras
	token := ctx.Query("id_token")
	if len(token) == 0 {
		logrus.Errorf("idp return no id_token")
		ctx.JSON(http.StatusBadRequest, "Login")
		return
	}
	// 2. parse token
	// check whether timeout
	userInfo, err := jwt.Parse(token, "/root/wsj/public_key")
	if err != nil {
		logrus.Errorf("parse token %s failed: %v", token, err)
		ctx.JSON(http.StatusForbidden, fmt.Errorf("parse token %s failed: %v",
			token, err))
		return
	}
	logrus.Infof("parse user from token: %v", userInfo)

	// 3. get request ip
	requestIP := ctx.Request.RemoteAddr
	ipParts := strings.Split(requestIP, ":")
	if len(ipParts) > 0 {
		requestIP = ipParts[0]
	}
	//ctx.GetHeader("X-Real-IP")
	proxyIPs := ctx.GetHeader("X-Forward-For")
	logrus.Infof("+++++++GetLogin id_token %s", token)
	if len(proxyIPs) != 0 {
		requestIP = proxyIPs
	}

	if tokenList, ok := cache[requestIP]; ok {
		if _, ok := tokenList[token]; ok != true {
			tokenList[token] = token
			cache[requestIP] = tokenList
		}
	} else {
		newToken := map[string]string{token: token}
		cache[requestIP] = newToken
	}
	ipMsg := requestIP
	userNameMsg := userInfo.UserName
	key := fmt.Sprintf("%s++++++%s", ipMsg, userNameMsg)
	_, err = obj.RedisC.SAdd(ctx, key, token).Result()
	if err != nil {
		logrus.Errorf("put token to redis SAdd failed: %v", err)
	}
	timeout := userInfo.Expired - time.Now().Unix()
	logrus.Infof("++++++ timeout: %d --- %d", timeout, userInfo.Expired)
	err = obj.RedisC.SetKeyTimeout(userInfo.Token, "", time.Second*time.Duration(timeout))
	if err != nil {
		logrus.Errorf("put token to redis failed: %v", err)
	}
	logrus.Infof("+++++++GetLogin token cache %v", cache)
	logrus.Infof("+++++++Request IP %v", requestIP)
	err = k8s.CreateCRB("https://10.254.24.66:6443",
		"/etc/kubernetes/kubeconfig",
		userInfo.UserName, "defult")
	if err != nil {
		logrus.Errorf("CreateClusterRolebinding failed: %v", err)
	}
	// 4. write html to browser
	// wait 5s and close window
	ctx.Writer.Header().Add("Content-Type", "text/html")
	if _, err := fmt.Fprintf(ctx.Writer, DefaultLocalServerSuccessHTML); err != nil {
		ctx.JSON(http.StatusInternalServerError, "ByeBye")
	}
	//ctx.HTML(http.StatusOK, "/Users/wangshuaijian/go_workspace/src/github.com/jianzi123/awesomeJJ/browser/tmp.html",
	//	gin.H{})

}

func (obj *App) Review(ctx *gin.Context) {
	logrus.Infof("+++++++Review")

	resp := v1beta1.TokenReview{
		TypeMeta: v1.TypeMeta{APIVersion: server.Authv1Beta1,
			Kind: server.TokenReview},
		ObjectMeta: v1.ObjectMeta{CreationTimestamp: v1.Now()},
	}

	req := &v1beta1.TokenReview{}
	if err := ctx.ShouldBind(req); err != nil {
		logrus.Errorf("Review get request body failed: %v", err)
		resp.Status = v1beta1.TokenReviewStatus{Error: fmt.Sprintf("cannot parse token request %v",
			err)}
		ctx.JSON(http.StatusForbidden, &resp)
		return
	}

	token, err := server.CheckAndParseTokenView(req)
	if err != nil {
		logrus.Errorf("CheckAndParseTokenView failed: %v", err)
		resp.Status = v1beta1.TokenReviewStatus{Error: err.Error()}
		ctx.JSON(http.StatusForbidden, &resp)
		return
	}
	tokenInfo, err := jwt.Parse(token, "/root/wsj/public_key")
	if err != nil{
		logrus.Errorf("cannot parse token: %s, err: %v", token, err)
	}else{
		if tokenInfo.IsValid != true{
			logrus.Errorf("token is expired")
			resp.Status = v1beta1.TokenReviewStatus{Error: fmt.Sprintf("token is expired")}
			ctx.JSON(http.StatusForbidden, &resp)
			return
		}
	}

	logrus.Infof("get review request from client-go %s", token)
	// check token in webhook-server
	_, err = obj.RedisC.Get(ctx, token).Result()
	// forbidden
	if err != nil{
		logrus.Infof("obj.RedisC.Get failed: %v", err)
		resp.Status = v1beta1.TokenReviewStatus{Error: err.Error()}
		ctx.JSON(http.StatusForbidden, &resp)
		return
	}
	// suc
	resp.Status = v1beta1.TokenReviewStatus{
		Authenticated: true,
		User: v1beta1.UserInfo{
			Username: tokenInfo.UserName,
			UID:      "",
			Groups:   []string{},
		},
	}
	ctx.JSON(http.StatusOK, &resp)
	return
}

type App struct {
	RedisC *db.RedisClient
}

func main() {
	//go func() {
	//	b := Browser{}
	//	//b.Open("http://idaas-test.zhenguanyu.com/")
	//	//b.Open("http://idaas-test.zhenguanyu.com/enduser/sp/sso/yfdjwt8?enterpriseId=yfd")
	//	b.Open("http://idaas-test.zhenguanyu.com/enduser/sp/sso/yfdjwt9?enterpriseId=yfd")
	//}()
	logrus.SetLevel(logrus.DebugLevel)
	gin.SetMode(gin.DebugMode)

	client, err := db.NewRedisClient()
	if err != nil {
		logrus.Errorf("db.NewRedisClient failed: %v", err)
		return
	}

	app := App{
		RedisC: client,
	}

	msg := make(chan string)
	go client.SubscribeCustom(msg)

	go func() {
		// loop
		for redisMsg := range msg {
			// ip_username_token
			ctx := context.Background()
			msgParts := strings.Split(redisMsg, "++++++")
			if len(msgParts) != 3 {
				logrus.Errorf("get data from redis len is not 3 %s", redisMsg)
				continue
			}
			ipMsg := msgParts[0]
			userNameMsg := msgParts[1]
			tokenMsg := msgParts[2]
			key := fmt.Sprintf("%s++++++%s", ipMsg, userNameMsg)
			value := tokenMsg
			_, err := client.SRem(ctx, key, value).Result()
			if err != nil {
				logrus.Errorf("redis SRem failed: %v", err)
				continue
			}
		}
	}()
	engine := gin.New()

	engine.POST("/login", Login)

	//proc := engine.Group("/v1").Use(RequestPrint)
	proc := engine.Group("/v1").Use(RequestPrint)
	{
		proc.GET("/gitinfo", GitInfo)
		proc.POST("/login", Login)
		proc.GET("/login", app.GetLogin)
		proc.GET("/tokens", app.GetTokensByIP)
		proc.GET("/status", Status)
		proc.POST("/review", app.Review)
	}
	engine.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))

	port := os.Getenv("agent_port")
	if len(port) == 0 {
		port = "8899"
	}
	url := fmt.Sprintf("0.0.0.0:%s", port)
	err = engine.Run(url)
	logrus.Errorf("restAPI server run failed: %v. \n", err)
}
