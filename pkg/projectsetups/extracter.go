package projectsetups

import (
	"fmt"
	projectv1 "github.com/openshift/api/project/v1"
	authclientv1 "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
	projectclientv1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"strings"
)

type Extracter struct {
	Authclient     *authclientv1.AuthorizationV1Client
	Projectsclient *projectclientv1.ProjectV1Client
}

func (e Extracter) ExtractProjectsetups() ([]*Projectsetup, []projectv1.Project) {

	allProjectsetups := make(map[string]*Projectsetup)
	var unmappedNamespaces []projectv1.Project

	namespaces, err := e.Projectsclient.Projects().List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Extracting %d namespaces: ", len(namespaces.Items))

	for _, ns := range namespaces.Items {

		_, _ = fmt.Fprint(os.Stderr, ".")

		// Skip platform-namespaces
		if strings.HasPrefix(ns.Name, "kube") ||
			strings.HasPrefix(ns.Name, "openshift") ||
			strings.HasPrefix(ns.Name, "default") ||
			strings.HasPrefix(ns.Name, "management") {
			continue
		}

		var p *Projectsetup
		var role string
		if strings.HasSuffix(ns.Name, "-ci") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-ci"))
			role = RoleBuild
		} else if strings.HasSuffix(ns.Name, "-sit") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-sit"))
			role = RoleDeploy
		} else if strings.HasSuffix(ns.Name, "-brumm") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-brumm"))
			role = RoleDeploy
		} else if strings.HasSuffix(ns.Name, "-nasse") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-nasse"))
			role = RoleDeploy
		} else if strings.HasSuffix(ns.Name, "-tussi") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-tussi"))
			role = RoleDeploy
		} else if strings.HasSuffix(ns.Name, "-dev") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-dev"))
			role = RoleDeploy
		} else if strings.HasSuffix(ns.Name, "-test") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-test"))
			role = RoleDeploy
		} else if strings.HasSuffix(ns.Name, "-prod-ready") {
			p = getOrRegisterProjectsetup(allProjectsetups, strings.TrimSuffix(ns.Name, "-prod-ready"))
			role = RolePromote
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
		p.Namespaces = append(p.Namespaces, &n)
		p.DetectedNamespaces = append(p.DetectedNamespaces, ns)
	}

	var ps []*Projectsetup
	for _, p := range allProjectsetups {
		ps = append(ps, p)
	}

	_, _ = fmt.Fprint(os.Stderr, "\nExtracting rolebindings: ")
	e.extractRolebindings(ps)
	_, _ = fmt.Fprint(os.Stderr, "\n")

	return ps, unmappedNamespaces
}

func (e Extracter) extractRolebindings(ps []*Projectsetup) {
	for _, p := range ps {
		for _, ns := range p.Namespaces {
			if rblist, err := e.Authclient.RoleBindings(ns.name).List(metav1.ListOptions{}); err == nil {
				for _, rb := range rblist.Items {
					_, _ = fmt.Fprint(os.Stderr, ".")
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
