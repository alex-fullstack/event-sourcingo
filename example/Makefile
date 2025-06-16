SHELL = /bin/sh

start:
	minikube start --driver=docker --force
	kubectl delete secret user-secret -n=default --ignore-not-found \
 	&& kubectl create secret generic user-secret -n=default \
 	--from-literal=postgres-url=${USER_POSTGRES_URL} \
 	--from-literal=mongo-url=${MONGO_URL} \
 	--from-literal=access-token-secret=${ACCESS_TOKEN_SECRET} \
 	--from-literal=refresh-token-secret=${REFRESH_TOKEN_SECRET} \
    && kubectl apply -f ms-user/charts/configmap.yaml \
    && kubectl apply -f ms-user/charts/deployment.yaml \
    && kubectl apply -f ms-user/charts/service.yaml \
    && kubectl delete secret policy-secret -n=default --ignore-not-found \
    && kubectl create secret generic policy-secret -n=default \
    --from-literal=postgres-url=${POLICY_POSTGRES_URL} \
    --from-literal=mongo-url=${MONGO_URL} \
    && kubectl apply -f ms-policy/charts/configmap.yaml \
    && kubectl apply -f ms-policy/charts/deployment.yaml \
    && kubectl apply -f ms-policy/charts/service.yaml \
    && kubectl delete secret api-secret -n=default --ignore-not-found \
    && kubectl create secret generic api-secret -n=default \
    --from-literal=elasticsearch-password=${ELASTICSEARCH_PASSWORD} \
    --from-literal=elasticsearch-certificate-fingerprint=${ELASTICSEARCH_CERTIFICATE_FINGERPRINT} \
    --from-literal=access-token-secret=${ACCESS_TOKEN_SECRET} \
    && kubectl apply -f ms-api/charts/configmap.yaml \
    && kubectl apply -f ms-api/charts/deployment.yaml \
    && kubectl apply -f ms-api/charts/service.yaml