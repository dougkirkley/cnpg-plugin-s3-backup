#!/bin/bash

kubectl apply --server-side -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.2.yaml
sleep 10
kubectl patch deployment -n cnpg-system cnpg-controller-manager --patch-file kubernetes/deployment-patch.json
sleep 30
kubectl apply -f kubernetes/cluster-example.yaml
sleep 60
kubectl apply -f kubernetes/backup-example.yaml