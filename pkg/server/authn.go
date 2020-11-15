package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"

	"k8s.io/api/authentication/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	authv1Beta1 = "authentication.k8s.io/v1beta1"
	tokenReview = "TokenReview"
)

func checkAndParseTokenView(tv *v1beta1.TokenReview) (string, error) {
	token := ""
	if tv == nil {
		return token, fmt.Errorf("tokenReview is nil")
	}
	switch {
	case tv.APIVersion != authv1Beta1:
		return token, fmt.Errorf("unsupported API version %s",
			tv.APIVersion)
	case tv.Kind != tokenReview:
		return token, fmt.Errorf("unsupported Kind %s", tv.Kind)
	case tv.Spec.Token == "":
		return token, fmt.Errorf("missing token")
	}
	return tv.Spec.Token, nil
}

func Review(ctx *gin.Context) {
	logrus.Infof("+++++++status")
	ctx.JSON(http.StatusOK, "ok")
	req := &v1beta1.TokenReview{}
	if err := ctx.ShouldBind(req); err != nil {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("cannot parse token request %v",
			err))
		return
	}

	token, err := checkAndParseTokenView(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("%v", err))
		return
	}

	logrus.Infof("get review request from client-go %s",
		token)
	// check token in webhook-server
	forbidden := false

	resp := v1beta1.TokenReview{
		TypeMeta:   v1.TypeMeta{APIVersion: authv1Beta1, Kind: tokenReview},
		ObjectMeta: v1.ObjectMeta{CreationTimestamp: v1.Now()},
	}

	// failed
	if forbidden == true {
		resp.Status = v1beta1.TokenReviewStatus{Error: err.Error()}
		ctx.JSON(http.StatusForbidden, &resp)
		return
	}

	// suc
	resp.Status = v1beta1.TokenReviewStatus{
		Authenticated: true,
		User:          v1beta1.UserInfo{
			//Username: u.Username,
			//UID:      u.UID,
			//Groups:   u.Groups,
		},
	}
	ctx.JSON(http.StatusOK, &resp)
}
