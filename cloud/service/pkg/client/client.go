package client

import (
	"encoding/base64"

	helmclient "github.com/mittwald/go-helm-client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func ParseKubeConfigBytes(cfg string) ([]byte, error) {
	kubeConfigBytes, err := base64.StdEncoding.DecodeString(cfg)
	if err != nil {
		return nil, err
	}

	return kubeConfigBytes, err
}

func NewClientSetFromBytes(data []byte) (*kubernetes.Clientset, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func NewClientSetFromString(cfg string) (*kubernetes.Clientset, error) {
	kubeConfigBytes, err := ParseKubeConfigBytes(cfg)
	if err != nil {
		return nil, err
	}

	return NewClientSetFromBytes(kubeConfigBytes)
}

func NewClusterSet(cfg string) (*ClusterSet, error) {
	kubeConfigBytes, err := ParseKubeConfigBytes(cfg)
	if err != nil {
		return nil, err
	}

	cs := &ClusterSet{}
	if err = cs.Complete(kubeConfigBytes); err != nil {
		return nil, err
	}

	return cs, nil
}

func NewHelmClient(namespace string, kubeConfig *rest.Config) (helmclient.Client, error) {
	opt := &helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace: namespace,
			Debug:     true,
			Linting:   false,
			DebugLog: func(format string, v ...interface{}) {
				klog.Infof(format, v)
			},
		},
		RestConfig: kubeConfig,
	}

	return helmclient.NewClientFromRestConf(opt)
}
