package main

import (
	"bufio"
	"fmt"
	"github.com/kschjeld/ocputils/pkg/clienthelper"
	projectv1 "github.com/openshift/api/project/v1"
	authv1client "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
	projectv1client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"path"
	"strings"
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

	e := extracter{
		authclient:     authclient,
		projectsclient: projectclient,
	}

	ps, unmapped := e.extractProjectsetups()
	fmt.Printf("Extracted %d projectsetups\n", len(ps))

	fmt.Println("Unmapped namespaces:")
	for _,v := range unmapped {
		fmt.Printf(" - %s\n", v.Name)
	}

	if filepath == "-" {
		fmt.Println("\nProposed projectsetups:")
		writeResourceDefinitionsStdout(ps)
	} else {
		writeResourceDefinitionsFile(filepath, ps)
	}
}

func writeResourceDefinitionsStdout( ps []*projectsetup) {
	for _,v := range ps {
		v.writeProjectsetupDefinition(os.Stdout)
		_, _ = fmt.Fprintf(os.Stdout, "-\n")
	}
}

func writeResourceDefinitionsFile(filepath string, ps []*projectsetup) {
	for _,v := range ps {

		filename := path.Join(filepath, v.name + ".yaml")
		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Failed to write resource to file: %s", err)
			return
		}
		writer := bufio.NewWriter(file)

		v.writeProjectsetupDefinition(writer)

		if err := writer.Flush(); err != nil {
			fmt.Printf("Failed to flush writer: %s", err)
			return
		}

		if err := file.Close(); err != nil {
			fmt.Printf("Failed to close file: %s", err)
			return
		}
	}
}

const (
	ROLE_BUILD = "build"
	ROLE_DEPLOY = "deploy"
	ROLE_PROMOTE = "promote"
)

type extracter struct {
	authclient *authv1client.AuthorizationV1Client
	projectsclient *projectv1client.ProjectV1Client
}

func (e extracter) extractProjectsetups() ([]*projectsetup, []projectv1.Project) {

	allProjectsetups := make(map[string] *projectsetup)
	var unmappedNamespaces []projectv1.Project

	namespaces, err := e.projectsclient.Projects().List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	_,_ = fmt.Fprintf(os.Stderr, "Extracting %d namespaces: ", len(namespaces.Items))

	for _, ns := range namespaces.Items {

		_,_ = fmt.Fprint(os.Stderr, ".")

		// Skip platform-namespaces
		if strings.HasPrefix(ns.Name, "kube") ||
			strings.HasPrefix(ns.Name, "openshift") ||
			strings.HasPrefix(ns.Name, "default") ||
			strings.HasPrefix(ns.Name, "management") {
			continue
		}

		var p *projectsetup
		var role string
		if strings.HasSuffix(ns.Name, "-ci") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-ci"))
			role = ROLE_BUILD
		} else if strings.HasSuffix(ns.Name, "-sit") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-sit"))
			role = ROLE_DEPLOY
		} else if strings.HasSuffix(ns.Name, "-brumm") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-brumm"))
			role = ROLE_DEPLOY
		} else if strings.HasSuffix(ns.Name, "-nasse") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-nasse"))
			role = ROLE_DEPLOY
		} else if strings.HasSuffix(ns.Name, "-tussi") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-tussi"))
			role = ROLE_DEPLOY
		} else if strings.HasSuffix(ns.Name, "-dev") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-dev"))
			role = ROLE_DEPLOY
		} else if strings.HasSuffix(ns.Name, "-test") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-test"))
			role = ROLE_DEPLOY
		} else if strings.HasSuffix(ns.Name, "-prod-ready") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-prod-ready"))
			role = ROLE_PROMOTE
		} else {
			unmappedNamespaces = append(unmappedNamespaces, ns)
			continue
		}

		n := namespace{
			name:          ns.Name,
			role:          role,
			editorsGroups: nil,
			viewersGroups: nil,
		}
		p.namespaces = append(p.namespaces, &n)
		p.detectedNamespaces = append(p.detectedNamespaces, ns)
	}

	var ps []*projectsetup
	for _, p := range allProjectsetups {
		ps = append(ps, p)
	}

	_,_ = fmt.Fprint(os.Stderr, "\nExtracting rolebindings: ")
	e.extractRolebindings(ps)
	_,_ = fmt.Fprint(os.Stderr, "\n")

	return ps, unmappedNamespaces
}

func (e extracter) extractRolebindings(ps []*projectsetup) {
	for _, p := range ps {
		for _, ns := range p.namespaces {
			_,_ = fmt.Fprint(os.Stderr, ".")
			if rblist, err := e.authclient.RoleBindings(ns.name).List( metav1.ListOptions{}); err == nil {
				for _, rb := range rblist.Items {
					if rb.RoleRef.Name == "admin" && len(rb.Subjects) > 0 && rb.Subjects[0].Kind == "Group" {
						p.addOwnerGroup(rb.Subjects[0].Name)
					}
					if rb.RoleRef.Name == "edit" && len(rb.Subjects) > 0 && rb.Subjects[0].Kind == "Group" {
						ns.addEditorGroup(rb.Subjects[0].Name)
					}

					if rb.RoleRef.Name == "view" && len(rb.Subjects) > 0 && rb.Subjects[0].Kind == "Group" {
						ns.addViewerGroup(rb.Subjects[0].Name)
					}
				}
			}
		}
	}
}

func getOrRegisterProjectsetup(allProjectsetups map[string] *projectsetup, name string) *projectsetup {
	p, found := allProjectsetups[name]
	if !found {
		p = &projectsetup{
			name:               name,
		}
		allProjectsetups[name] = p
	}
	return p
}

//
// Projectsetup

type projectsetup struct {
	name        string
	ownerGroups []string
	namespaces  []*namespace

	detectedNamespaces []projectv1.Project
}

func (p *projectsetup) addOwnerGroup(name string) {
	for _, v := range p.ownerGroups {
		if v == name {
			return
		}
	}
	p.ownerGroups = append(p.ownerGroups, name)
}


func (p projectsetup) writeProjectsetupDefinition(w io.Writer) {
	_, _ = fmt.Fprintf(w, "apiVersion: management.telenor.no/projectv1\n")
	_, _ = fmt.Fprintf(w, "kind: Projectsetup\n")
	_, _ = fmt.Fprintf(w, "metadata:\n")
	_, _ = fmt.Fprintf(w, "  name: %s\n", p.name)
	_, _ = fmt.Fprintf(w, "  namespace: projects-operator\n")
	_, _ = fmt.Fprintf(w, "spec:\n")
	_, _ = fmt.Fprintf(w, "  contactEmail: <fill in>\n")
	_, _ = fmt.Fprintf(w, "  description: <fill in>\n")
	_, _ = fmt.Fprintf(w, "  orderReference: <fill in>\n")
	if len(p.ownerGroups) == 1 {
		_, _ = fmt.Fprintf(w, "  ownerGroup: %v\n", p.ownerGroups[0])
	} else {
		_, _ = fmt.Fprintf(w, "  ownerGroup: Select one of %v\n", p.ownerGroups)
	}
	_, _ = fmt.Fprintf(w, "  namespaces:\n")
	for _, ns := range p.namespaces {
		_, _ = fmt.Fprintf(w, "  - ")
		ns.writeProjectsetupDefinition(w)
	}
}

//
// Namespace

type namespace struct {
	name string
	role string
	editorsGroups []string
	viewersGroups []string
}

func (n *namespace) addEditorGroup(name string) {
	for _, v := range n.editorsGroups {
		if v == name {
			return
		}
	}
	n.editorsGroups = append(n.editorsGroups, name)
}

func (n *namespace) addViewerGroup(name string) {
	for _, v := range n.viewersGroups {
		if v == name {
			return
		}
	}
	n.viewersGroups = append(n.viewersGroups, name)
}

func (ns *namespace) writeProjectsetupDefinition(w io.Writer) {
	_, _ = fmt.Fprintf(w, "name: %s\n", ns.name)
	_, _ = fmt.Fprintf(w, "    editorGroups: %s\n", strings.Join(ns.editorsGroups, ","))
	_, _ = fmt.Fprintf(w, "    viewerGroups: %s\n", strings.Join(ns.viewersGroups, ","))
	_, _ = fmt.Fprintf(w, "    roles:\n")
	_, _ = fmt.Fprintf(w, "    - %s\n",ns.role)
}
