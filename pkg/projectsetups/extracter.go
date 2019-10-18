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

// suffixRoleMappings drives the mapping and grouping logic. Only namespaces with suffixes in this map will go into
// a Projectsetup as one of the namespaces.
var suffixRoleMappings = map[string]Role {
	"-ci": RoleBuild,

	"-sit":RoleDeploy,
	"-brumm":RoleDeploy,
	"-nasse":RoleDeploy,
	"-tussi":RoleDeploy,
	"-dev":RoleDeploy,
	"-test":RoleDeploy,
	"-atest":RoleDeploy,
	"-itest":RoleDeploy,
	"-at":RoleDeploy,
	"-it":RoleDeploy,

	"-prod-ready":RolePromote,
}

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

		if ok, detectedRole, basename := stripAndIdentifyRole(ns.Name); ok {
			p = getOrRegisterProjectsetup(allProjectsetups, basename)
			role = string(detectedRole)
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

func stripAndIdentifyRole(name string) (bool,Role,string) {
	for suffix,role := range suffixRoleMappings {
		if strings.HasSuffix(name, suffix) {
			return true, role, strings.TrimSuffix(name,suffix)
		}
	}
	return false,"",""
}

func (e Extracter) extractRolebindings(ps []*Projectsetup) {
	for _, p := range ps {
		for _, ns := range p.Namespaces {
			_, _ = fmt.Fprint(os.Stderr, "*")
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
