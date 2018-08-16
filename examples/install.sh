#!/bin/bash

vault policy write helm-policy helm-policy.hcl
kubectl apply -f helm-serviceaccount.yml

VAULT_SA_NAME=$(kubectl get sa helm-vault -o jsonpath="{.secrets[*]['name']}")
SA_JWT_TOKEN=$(kubectl get secret $VAULT_SA_NAME -o jsonpath="{.data.token}" | base64 --decode; echo)
SA_CA_CRT=$(kubectl get secret $VAULT_SA_NAME -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)
K8S_HOST=$(kubectl exec consul-0 -- sh -c 'echo $KUBERNETES_SERVICE_HOST')

vault auth enable kubernetes
vault write auth/kubernetes/config \
  token_reviewer_jwt="$SA_JWT_TOKEN" \
  kubernetes_host="https://$K8S_HOST:443" \
  kubernetes_ca_cert="$SA_CA_CRT"

vault write auth/kubernetes/role/helm \
    bound_service_account_names=helm-vault \
    bound_service_account_namespaces=default \
    policies=helm-policy \
    ttl=24h