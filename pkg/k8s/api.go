package k8s

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	rv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

func CreateCRB(clusterUrl, filePath, username, cluster string) error {
	clientset, err := CreateClient(clusterUrl, filePath)
	if err != nil{
		logrus.Errorf("CreateClient failed: %v", err)
		return err
	}
	err = CheckClusterRolebindingExist(clientset, username, cluster)
	if err == nil{
		return nil
	}
	if err != nil{
		if strings.Contains(err.Error(), "not found") != true{
			return err
		}
	}

	err = CreateClusterRolebinding(clientset, username, cluster)
	if err != nil{
		logrus.Errorf("CheckClusterRolebindingExist failed: %v", err)
		return err
	}
	return nil
}

func CreateClient(clusterUrl, filePath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags(clusterUrl, filePath)
	if err != nil{
		logrus.Errorf("clientcmd.BuildConfigFromFlags failed: %v", err)
		return nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil{
		logrus.Errorf("kubernetes.NewForConfig failed: %v", err)
		return nil, err
	}
	return clientset, nil
}

func CheckClusterRolebindingExist(clientset *kubernetes.Clientset, username, cluster string) error {
	crbName := fmt.Sprintf("%s-%s", username, cluster)
	// access the API to list pods
	ctx := context.Background()
	_, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, crbName, v1.GetOptions{})
	if err != nil{
		return err
	}
	return nil
}


func CreateClusterRolebinding(clientset *kubernetes.Clientset, username, cluster string) error {
	crbName := fmt.Sprintf("%s-%s", username, cluster)
	// access the API to list pods
	ctx := context.Background()
	
	usr := rv1.Subject{
		Kind:      "User",
		APIGroup:  "rbac.authorization.k8s.io",
		Name:      username,
		Namespace: "",
	}
	crb := &rv1.ClusterRoleBinding{
		TypeMeta:   v1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:   crbName,
		},
		Subjects:   []rv1.Subject{usr},
		RoleRef:    rv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "kubectl-cr-all",
		},
	}

	_, err := clientset.RbacV1().ClusterRoleBindings().Create(ctx, crb, v1.CreateOptions{})
	if err != nil{
		if strings.Contains(err.Error(), "exist") == true{
			logrus.Infof("clusterrolebinding %s exist", crbName)
			return nil
		}
		return err
	}
	return nil
}
