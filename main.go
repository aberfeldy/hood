package main

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"strings"
)

//needed for json unmarshall
type VaultSecrets struct {
	Secrets []VaultSecret `json:"secrets"`
}

type VaultSecret struct {
	Name  string   `json:"name"`
	Path  string   `json:"path"`
	Props []string `json:"props"`
}

type SecretEnv struct {
	name  string
	value string
}
type Secret struct {
	name    string
	entries []SecretEnv
}

var (
	vaultAddress string
	clusterPath  string
	vault        *api.Client
	clientset    *kubernetes.Clientset
	namespace    string
)

func main() {
	vs := VaultSecrets{}
	parseEnv(vs)

	vault = vaultClient()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	} else {
		clientset = cs
	}

	for _, v := range vs.fetch() {
		v.Upsert()
	}
}

func parseEnv(vs VaultSecrets) {
	if v := os.Getenv("VAULT_SECRETS"); v != "" {
		err := json.Unmarshal([]byte(v), &vs)
		if err != nil {
			panic("could not unmarshall config from JSON")
		}
	} else {
		panic("could not parse config from Tiller")
	}
	if v := os.Getenv("VAULT_ADDRESS"); v != "" {
		vaultAddress = v
	} else {
		panic("could not find a valid vault address")
	}
	if v := os.Getenv("NAMESPACE"); v != "" {
		namespace = v
	} else {
		panic("could not find a valid namespace name")
	}
	if v := os.Getenv("CLUSTER_PATH"); v != "" {
		clusterPath = v
	} else {
		panic("could not find a clusterpath")
	}
}

func (s *Secret) Render() v1.Secret {
	var ks v1.Secret
	if s.name == "regsecret" {
		ks = v1.Secret{
			Type: v1.SecretTypeDockercfg,
			Data: map[string][]byte{
				".dockercfg": []byte(s.entries[0].value),
			},
		}

	} else {
		data := make(map[string][]byte)
		for _, se := range s.entries {
			data[se.name] = []byte(strings.Trim(se.value, " "))
		}
		ks = v1.Secret{
			Type: "opaque",
			Data: data,
		}

	}
	ks.Name = s.name
	ks.Namespace = namespace
	return ks
}

func (s *Secret) Upsert() {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(s.name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		sec := s.Render()
		_, err := clientset.CoreV1().Secrets(namespace).Create(&sec)
		if err != nil {
			panic(err.Error())
		}
	} else if err != nil {
		panic(err.Error())
	} else {
		err := clientset.CoreV1().Secrets(namespace).Delete(secret.Name, nil)
		if err != nil {
			panic(err.Error())
		}
		sec := s.Render()
		_, err = clientset.CoreV1().Secrets(namespace).Create(&sec)
		if err != nil {
			panic(err.Error())
		}
	}
}

