package clienthelper

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

func NewOCPClientWithUserconfig() (*rest.Config, error) {
	home := homeDir()
	if home == "" {
		log.Fatal("Could not locate home dir")
	}

	//kubeconfig := flag.String("kubeconfig", filepath.Join(homeDir(), ".kube", "config"), "(optional) path to kubeconfig file)")
	//flag.Parse()

	configpath := filepath.Join(homeDir(), ".kube", "config")
	return clientcmd.BuildConfigFromFlags("", configpath)

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
