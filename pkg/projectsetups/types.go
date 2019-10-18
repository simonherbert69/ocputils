package projectsetups

import (
	"fmt"
	"github.com/openshift/api/project/v1"
	"io"
	"strings"
)

type Role string
const (
	RoleBuild Role   = "build"
	RoleDeploy Role = "deploy"
	RolePromote Role = "promote"
)

func getOrRegisterProjectsetup(allProjectsetups map[string]*Projectsetup, name string) *Projectsetup {
	p, found := allProjectsetups[name]
	if !found {
		p = &Projectsetup{
			Name: name,
		}
		allProjectsetups[name] = p
	}
	return p
}

type Projectsetup struct {
	Name        string
	OwnerGroups []string
	Namespaces  []*namespace

	DetectedNamespaces []v1.Project
}

func (p *Projectsetup) addOwnerGroup(name string) {
	for _, v := range p.OwnerGroups {
		if v == name {
			return
		}
	}
	p.OwnerGroups = append(p.OwnerGroups, name)
}

func (p Projectsetup) WriteProjectsetupDefinition(w io.Writer) {
	_, _ = fmt.Fprintf(w, "apiVersion: management.telenor.no/projectv1\n")
	_, _ = fmt.Fprintf(w, "kind: Projectsetup\n")
	_, _ = fmt.Fprintf(w, "metadata:\n")
	_, _ = fmt.Fprintf(w, "  name: %s\n", p.Name)
	_, _ = fmt.Fprintf(w, "  namespace: projects-operator\n")
	_, _ = fmt.Fprintf(w, "spec:\n")
	_, _ = fmt.Fprintf(w, "  contactEmail: <fill in>\n")
	_, _ = fmt.Fprintf(w, "  description: <fill in>\n")
	_, _ = fmt.Fprintf(w, "  orderReference: <fill in>\n")
	if len(p.OwnerGroups) == 1 {
		_, _ = fmt.Fprintf(w, "  ownerGroup: %v\n", p.OwnerGroups[0])
	} else {
		_, _ = fmt.Fprintf(w, "  ownerGroup: Select one of %v\n", p.OwnerGroups)
	}
	_, _ = fmt.Fprintf(w, "  namespaces:\n")
	for _, ns := range p.Namespaces {
		_, _ = fmt.Fprintf(w, "  - ")
		ns.writeProjectsetupDefinition(w)
	}
}

type namespace struct {
	name          string
	role          string
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
	_, _ = fmt.Fprintf(w, "    - %s\n", ns.role)
}
