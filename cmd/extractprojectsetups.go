package main

import (
	"bufio"
	"fmt"
	"github.com/kschjeld/ocputils/pkg/clienthelper"
	"github.com/kschjeld/ocputils/pkg/projectsetups"
	authv1client "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
	projectv1client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	"log"
	"os"
	"path"
)

func main() {
	var filepath string
	if len(os.Args) != 2 {
		fmt.Println("\nExtract namespaces in cluster and create Projcetsetup resource definitions for those that match given patterns.")
		fmt.Println("The extracted resources MUST be modified to add extra information before using them with projects-operator.\n")
		fmt.Printf("Usage: %s <directory>\n Use value - to write to stdout.\n", path.Base(os.Args[0]))
		os.Exit(0)
	} else {
		filepath = os.Args[1]
	}

	config, err := clienthelper.NewOCPClientWithUserconfig()
	if err != nil {
		log.Fatal(err)
	}

	projectclient, err := projectv1client.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	authclient, err := authv1client.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	e := projectsetups.Extracter{
		Authclient:     authclient,
		Projectsclient: projectclient,
	}

	ps, unmapped := e.ExtractProjectsetups()

	fmt.Println("Unmapped namespaces:")
	for _, v := range unmapped {
		fmt.Printf(" - %s\n", v.Name)
	}

	if filepath == "-" {
		fmt.Println("\nGenerated projectsetups:\n")
		writeResourceDefinitionsStdout(ps)
	} else {
		writeResourceDefinitionsFile(filepath, ps)
	}

	fmt.Printf("Extracted %d projectsetups to %s\n", len(ps), filepath)
}

func writeResourceDefinitionsStdout(ps []*projectsetups.Projectsetup) {
	for _, v := range ps {
		v.WriteProjectsetupDefinition(os.Stdout)
		_, _ = fmt.Fprintf(os.Stdout, "-\n")
	}
}

func writeResourceDefinitionsFile(filepath string, ps []*projectsetups.Projectsetup) {
	for _, v := range ps {
		filename := path.Join(filepath, v.Name+".yaml")
		if err := writeProjectsetupToFile(v, filename); err != nil {
			fmt.Printf("Failed to write resource to file: %s", err)
			return
		}
	}
}

func writeProjectsetupToFile(v *projectsetups.Projectsetup, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	v.WriteProjectsetupDefinition(writer)
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}
