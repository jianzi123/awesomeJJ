package main

import (
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

        // 用法
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

type Browser struct{}

// Open opens the default browser.
func (*Browser) Open(url string) error {
	return browser.OpenURL(url)
}

func RequestPrint(ctx *gin.Context) {
	dump, err := httputil.DumpRequest(ctx.Request, true)
	logrus.Debugf("解析请求: %s; err: %v. \n", string(dump), err)

	// proxy 透传
	requestIP := strings.Split(ctx.Request.RemoteAddr, ":")[0]
	fmt.Println(requestIP)
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

func GetTokensByIP(ctx *gin.Context) {
	ip := ctx.Query("node")
	if len(ip) == 0 {
		ctx.JSON(http.StatusBadRequest, "...")
	}
	// get token from cache
	resp := AuthResponse{
		ClientIP: ip,
		Token:    []string{""},
	}
	ctx.JSON(http.StatusOK, &resp)
}

func GetLogin(ctx *gin.Context) {

	token := ctx.Query("id_token")
	if len(token) == 0 {
		logrus.Errorf("idp return no id_token")
		ctx.JSON(http.StatusBadRequest, "Login")
		return
	}

	requestIP := ctx.Request.RemoteAddr
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
		cache[token] = newToken
	}
	//ctx.JSON(http.StatusOK, DefaultLocalServerSuccessHTML)
	ctx.Writer.Header().Add("Content-Type", "text/html")
	if _, err := fmt.Fprintf(ctx.Writer, DefaultLocalServerSuccessHTML); err != nil {
		ctx.JSON(http.StatusInternalServerError, "ByeBye")
	}
	//ctx.HTML(http.StatusOK, "/Users/wangshuaijian/go_workspace/src/github.com/jianzi123/awesomeJJ/browser/tmp.html",
	//	gin.H{})

}

func main() {
	go func() {
		b := Browser{}
		//b.Open("http://idaas-test.zhenguanyu.com/")
		b.Open("http://idaas-test.zhenguanyu.com/enduser/sp/sso/yfdjwt8?enterpriseId=yfd")
	}()

	gin.SetMode(gin.DebugMode)
	engine := gin.New()

	engine.POST("/login", Login)

	proc := engine.Group("/v1").Use(RequestPrint)
	{
		proc.GET("/gitinfo", GitInfo)
		proc.POST("/login", Login)
		proc.GET("/login", GetLogin)
		proc.GET("/tokens", GetTokensByIP)
		proc.GET("/status", Status)
	}
	engine.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))

	port := os.Getenv("agent_port")
	if len(port) == 0 {
		port = "8899"
	}
	url := fmt.Sprintf("0.0.0.0:%s", port)
	err := engine.Run(url)
	logrus.Errorf("restAPI server run failed: %v. \n", err)
}
