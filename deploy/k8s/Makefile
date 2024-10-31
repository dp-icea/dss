.PHONY: start-k8s-cluster
start-k8s-cluster:
	kind create cluster --config ./cluster/brutm-cluster.yml --verbosity=3
	kubectl cluster-info --context kind-dp-icea-brutm-cluster

.PHONY: setup-k8s-namespace
setup-k8s-namespace:
	kubectl apply -f ./namespace/namespace.yml

.PHONY: create-k8s-service-account
create-k8s-service-account:
	kubectl apply -f ./cluster-auth/service-account.yml
	kubectl apply -f ./cluster-auth/cluster-role.yml
	kubectl apply -f ./cluster-auth/secrets.yml
	kubectl get secret admin-user -n development -o jsonpath={".data.token"} | base64 -d

.PHONY: start-k8s-dashboard
start-k8s-dashboard:
	helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
	helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard
	kubectl -n kubernetes-dashboard port-forward svc/kubernetes-dashboard-kong-proxy 8443:443

.PHONY: start-k8s-helm-dashboard
start-k8s-helm-dashboard:
	helm plugin install https://github.com/komodorio/helm-dashboard.git
	helm plugin update dashboard
	helm dashboard


.PHONY: stop-k8s-cluster
stop-k8s-cluster:
	kubectl delete serviceaccount brutm -n development
	kubectl delete clusterrolebinding cluster-admin  -n development
	kind delete cluster --name dp-icea-brutm-cluster