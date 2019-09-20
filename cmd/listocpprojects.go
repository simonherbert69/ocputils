package main

import (
	"fmt"
	"github.com/kschjeld/ocputils/pkg/clienthelper"
	projectv1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

func main() {
	fmt.Println("Getting projects..")

	config, err := clienthelper.NewOCPClientWithUserconfig()
	if err != nil {
		log.Fatal(err)
	}

	projectclient, err := projectv1.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	projects, err := projectclient.Projects().List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, project := range projects.Items {
		fmt.Printf("Project: %s\n", project.Name)
		// TODO Show ownership, lease time, guests etc for the project
	}

}

