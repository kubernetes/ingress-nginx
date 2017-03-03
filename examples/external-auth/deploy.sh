#!/usr/bin/env bash

set -x
set -u
set -e

echo "deploying oauth2proxy for ${DOMAIN}"

if [[ -z "${DOMAIN:-}" ]]; then
	echo "You must set \$DOMAIN."
	exit -1
fi

if [[ ! -f "authenticated_emails" ]]; then
	echo "You must create './authenticated_emails'."
	exit -1
fi

if [[ ! -f "oauth2proxy.config" ]]; then
	echo "You must create './oauth2proxy.config'."
	exit -1
fi

force_cleanup="n"
answer="y"
if kubectl describe secret oauth2proxy &>/dev/null ; then
	echo "secret 'oauth2proxy' already exists."
	echo "do you want to replace it and cycle the 'oauth2proxy' container?"
	read answer
	if [[ "${answer}" == "y" ]]; then
		kubectl delete secret oauth2proxy || true
		kubectl delete deployment oauth2proxy || true
	fi
	force_cleanup="y"
fi
if [[ "${answer}" == "y" ]]; then
	kubectl create secret generic oauth2proxy \
		--from-file=oauth2proxy.config=./oauth2proxy.config \
		--from-file=authenticated_emails=./authenticated_emails
fi

sed "s|__DOMAIN__|${DOMAIN}|g" ./oauth2proxy.deployment.yaml | kubectl apply -f -
sed "s|__DOMAIN__|${DOMAIN}|g" ./oauth2proxy.ingress.yaml | kubectl apply -f -
sed "s|__DOMAIN__|${DOMAIN}|g" ./oauth2proxy.service.yaml | kubectl apply -f -
sed "s|__DOMAIN__|${DOMAIN}|g" ./dashboard.ingress.yaml | kubectl apply -f -
