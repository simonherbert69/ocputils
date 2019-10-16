package main

import (
	"flag"
	"fmt"
	"github.com/kschjeld/ocputils/pkg/clienthelper"
	"github.com/kschjeld/ocputils/pkg/usercache"
	"github.com/openshift/api/user/v1"
	userv1 "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"text/tabwriter"
)

func main() {

	showGroup := flag.String("group", "", "Group to show, or empty to show all")
	useSimpleOutput := flag.Bool("simple", false, "Show output using simple Ansible Tower compatible formatting")
	flag.Parse()

	config, err := clienthelper.NewOCPClientWithUserconfig()
	if err != nil {
		log.Fatal(err)
	}

	userclient, err := userv1.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	userinfo, err := usercache.NewWithClient(userclient)
	if err != nil {
		log.Fatal(err)
	}

	var groupList []v1.Group
	if *showGroup != "" {
		group, err := userclient.Groups().Get(*showGroup, metav1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}
		groupList = append(groupList, *group)
	} else {
		groups, err := userclient.Groups().List(metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		groupList = append(groupList, groups.Items...)
	}

	for _, group := range groupList {

		if *useSimpleOutput {
			printGroupSimple(group)
		} else {
			printGroupFormatted(group, userinfo)
		}
	}
}

func printGroupSimple(group v1.Group) {
	fmt.Printf("\n%s\n", group.Name)
	for _, user := range group.Users {
		fmt.Println(user)
	}
	fmt.Printf("\n")
}

func printGroupFormatted(group v1.Group, cache *usercache.Usercache) {

	w := new(tabwriter.Writer)
	if group.Users != nil {
		w.Init(os.Stdout, 8, 12, 0, '\t', 0)
		fmt.Printf("Group: %s\n", group.Name)
		fmt.Printf(" Members:\n")
		for _, user := range group.Users {
			fmt.Fprintf(w, " - %s\t%s\n", user, cache.GetFullname(user))
			w.Flush()
		}
	}
	fmt.Println("")
}
