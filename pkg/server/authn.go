package server

import (
	"fmt"
	"k8s.io/api/authentication/v1beta1"
)

const (
	Authv1Beta1 = "authentication.k8s.io/v1beta1"
	TokenReview = "TokenReview"
)

func CheckAndParseTokenView(tv *v1beta1.TokenReview) (string, error) {
	token := ""
	if tv == nil {
		return token, fmt.Errorf("tokenReview is nil")
	}
	switch {
	case tv.APIVersion != Authv1Beta1:
		return token, fmt.Errorf("unsupported API version %s",
			tv.APIVersion)
	case tv.Kind != TokenReview:
		return token, fmt.Errorf("unsupported Kind %s", tv.Kind)
	case tv.Spec.Token == "":
		return token, fmt.Errorf("missing token")
	}
	return tv.Spec.Token, nil
}

