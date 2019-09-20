package usercache

import (
	userv1 "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Usercache struct {
	client *userv1.UserV1Client
	users map[string]string
}

func NewWithClient(client *userv1.UserV1Client) (*Usercache, error) {
	u := &Usercache{
		client: client,
	}
	return u, u.loadCache()
}

func (u *Usercache) loadCache() error {
	userlist, err := u.client.Users().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	u.users = make(map[string]string)
	for _, user := range userlist.Items {
		u.users[user.Name] = user.FullName
	}
	return nil
}

func (u *Usercache) GetFullname(uid string) string {
	return u.users[uid]
}
