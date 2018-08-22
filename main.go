package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/api"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
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
	if v := os.Getenv("VAULT_SECRETS"); v != "" {
		json.Unmarshal([]byte(v), &vs)
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
		namespace = v
	} else {
		panic("could not find a clusterpath")
	}

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

	for _, v := range vs.resolve() {
		v.Upsert()
	}
}

func (secrets VaultSecrets) resolve() []Secret {
	var kubeSecrets []Secret
	for _, s := range secrets.Secrets {
		secretValue, err := vault.Logical().Read(s.Path)
		//ToDo if path returns nil continue
		if err != nil {
			panic(err.Error())
		}
		kubeSecret := Secret{name: s.Name}
		for _, p := range s.Props {
			//ToDo if props returns nil continue
			if secretValue.Data[p] != nil {
				se := SecretEnv{
					name:  fmt.Sprintf("%s-%s", s.Name, p),
					value: fmt.Sprintf("%s", secretValue.Data[p]),
				}
				kubeSecret.entries = append(kubeSecret.entries, se)
			}
		}
		kubeSecrets = append(kubeSecrets, kubeSecret)
	}
	return kubeSecrets

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
			data[se.name] = []byte(se.value)
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

type Login struct {
	JWT  string `json:"jwt"`
	Role string `json:"role"`
}

type VaultAuth struct {
	Auth struct {
		ClientToken string `json:"client_token"`
	} `json:"auth"`
}

func vaultClient() *api.Client {
	dat, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		panic(err.Error())
	}
	login := Login{
		JWT:  string(dat),
		Role: "helm",
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(login)
	data, err := json.Marshal(login)
	if err != nil {
		panic(err.Error())
	}
 	req, err := http.NewRequest("POST", fmt.Sprintf("%sv1/auth/%s/login", vaultAddress, clusterPath), bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var auth VaultAuth
	err = json.NewDecoder(resp.Body).Decode(&auth)
	if err != nil {
		panic(err.Error())
	}

	vault, err = api.NewClient(&api.Config{
		Address: vaultAddress,
	})
	if err != nil {
		panic(err.Error())
	}
	vault.SetToken(auth.Auth.ClientToken)

	return vault
}
