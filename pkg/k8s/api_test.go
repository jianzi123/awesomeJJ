package k8s

import "testing"

func Test_CreateCRB(t *testing.T)  {
	CreateCRB("https://10.254.24.66:6443",
		"/Users/wangshuaijian/go_workspace/src/github.com/jianzi123/awesomeJJ/script/rbac/kubeconfig",
		"wangshuaijian", "china")
}
