#!/bin/bash

kubectl delete -f kubernetes/backup-example.yaml --wait
kubectl delete -f kubernetes/cluster-example.yaml --wait
kubectl delete -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.1.yaml
