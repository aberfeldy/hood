package main

import "testing"
import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
)

var config = []byte(`
{
  "secrets": [
    {
      "name": "rds",
      "path": "secret/prd/rds/project",
      "props": [
        "host",
        "port"
      ]
    },
    {
      "name": "stuff",
      "path": "secret/prd/stuff",
      "props": [
        "archive-aws-s3-bucket",
        "archive-aws-s3-key",
        "archive-aws-s3-secret"
      ]
    }
  ]
}`)

func renderHelper(j []byte) (VaultSecrets, error) {
	vs := VaultSecrets{}
	err := json.Unmarshal([]byte(config), &vs)
	return vs, err
}

func TestParserToReturnValidConfig(t *testing.T) {

	should := VaultSecrets{Secrets: []VaultSecret{
		{
			Name: "rds",
			Path: "secret/prd/rds/project",
			Props: []string{
				"host",
				"port",
			},
		},
		{
			Name: "stuff",
			Path: "secret/prd/stuff",
			Props: []string{
				"archive-aws-s3-bucket",
				"archive-aws-s3-key",
				"archive-aws-s3-secret",
			},
		},
	}}
	vs, err := renderHelper(config)
	if err != nil {
		t.Errorf("%s\n", err.Error())
	}
	a, _ := json.Marshal(vs)
	b, _ := json.Marshal(should)

	assert.Equal(t, string(b), string(a))

}

func TestRenderToReturnKubeSecret(t *testing.T) {
	namespace = "test"
	should := `{"metadata":{"name":"ChangeMe","namespace":"test","creationTimestamp":null},"data":{"test-db-super-secret-host":"aG9yc3RuYS5tZQ==","test-db-super-secret-port":"MzMwNg==","test-db-super-secret-user":"YmVybmQ="},"type":"opaque"}`
	sec := Secret{
		name: "ChangeMe",
		entries: []SecretEnv{
			{"test-db-super-secret-host", "horstna.me"},
			{"test-db-super-secret-port", "3306"},
			{"test-db-super-secret-user", "bernd"},
		},
	}
	secTrim := Secret{
		name: "ChangeMe",
		entries: []SecretEnv{
			{"test-db-super-secret-host", "horstna.me "},
			{"test-db-super-secret-port", "3306"},
			{"test-db-super-secret-user", "bernd"},
		},
	}

	kubeSecret := sec.Render()
	j, err := json.Marshal(kubeSecret)
	if err != nil {
		t.Errorf("could not marshall Secret: %s", sec)
	}
	assert.Equal(t, should, string(j))
	kubeSecretTrim := secTrim.Render()
	k, err := json.Marshal(kubeSecretTrim)
	if err != nil {
		t.Errorf("could not marshall Secret: %s", sec)
	}
	assert.Equal(t, should, string(k))

}

func TestRenderToReturnRegsecret(t *testing.T) {
	namespace = "test"
	should := `{"metadata":{"name":"regsecret","namespace":"test","creationTimestamp":null},"data":{".dockercfg":"YSByZWFsIHNlY3JldCBkb2NrZXIgbG9naW4ganNvbg=="},"type":"kubernetes.io/dockercfg"}`
	sec := Secret{
		name: "regsecret",
		entries: []SecretEnv{
			{name: "dockercfg", value: "a real secret docker login json"},
		},
	}

	kubeSecret := sec.Render()
	j, err := json.Marshal(kubeSecret)
	if err != nil {
		t.Errorf("could not marshall Secret: %s", sec)
	}
	assert.Equal(t, should, string(j))

}
