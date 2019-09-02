package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/api"
	"io/ioutil"
	"net/http"
)
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

func (secrets VaultSecrets) fetch() []Secret {
	var kubeSecrets []Secret
	for _, s := range secrets.Secrets {
		secretValue, err := vault.Logical().Read(s.Path)
		if secretValue == nil {
			continue
		}
		if err != nil {
			panic(err.Error())
		}
		kubeSecret := Secret{name: s.Name}
		for _, p := range s.Props {
			if s.Props == nil {
				continue
			}
			if secretValue.Data[p] != nil {
				se := SecretEnv{
					name:  fmt.Sprintf("%s", p),
					value: fmt.Sprintf("%s", secretValue.Data[p].(string)),
				}
				kubeSecret.entries = append(kubeSecret.entries, se)
			}
		}
		kubeSecrets = append(kubeSecrets, kubeSecret)
	}
	return kubeSecrets

}
