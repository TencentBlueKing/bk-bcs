## 编译后产出物详情
```text
.
|-- bcs-services
|   |-- bcs-bkcmdb-synchronizer
|   |   |-- Dockerfile
|   |   |-- bcs-bkcmdb-synchronizer
|   |   |-- bcs-bkcmdb-synchronizer.json.template
|   |   `-- container-start.sh
|   |-- bcs-cluster-manager
|   |   |-- Dockerfile
|   |   |-- bcs-cluster-manager
|   |   |-- bcs-cluster-manager.json.template
|   |   |-- cloud.json.template
|   |   |-- container-start.sh
|   |   `-- swagger
|   |       `-- swagger-ui
|   |           |-- clustermanager.swagger.json
|   |           |-- favicon-16x16.png
|   |           |-- favicon-32x32.png
|   |           |-- index.html
|   |           |-- oauth2-redirect.html
|   |           |-- swagger-ui-bundle.js
|   |           |-- swagger-ui-bundle.js.map
|   |           |-- swagger-ui-standalone-preset.js
|   |           |-- swagger-ui-standalone-preset.js.map
|   |           |-- swagger-ui.css
|   |           |-- swagger-ui.css.map
|   |           |-- swagger-ui.js
|   |           `-- swagger-ui.js.map
|   |-- bcs-cluster-reporter
|   |   |-- Dockerfile
|   |   `-- bcs-cluster-reporter
|   |-- bcs-data-manager
|   |   |-- Dockerfile
|   |   |-- bcs-data-manager
|   |   |-- bcs-data-manager.json.template
|   |   |-- container-start.sh
|   |   `-- swagger
|   |       |-- bcs-data-manager.swagger.json
|   |       |-- favicon-16x16.png
|   |       |-- favicon-32x32.png
|   |       |-- index.html
|   |       |-- oauth2-redirect.html
|   |       |-- swagger-ui-bundle.js
|   |       |-- swagger-ui-bundle.js.map
|   |       |-- swagger-ui-standalone-preset.js
|   |       |-- swagger-ui-standalone-preset.js.map
|   |       |-- swagger-ui.css
|   |       |-- swagger-ui.css.map
|   |       |-- swagger-ui.js
|   |       `-- swagger-ui.js.map
|   |-- bcs-gateway-discovery
|   |   |-- Dockerfile.apisix
|   |   |-- Dockerfile.gateway
|   |   |-- Dockerfile.micro-gateway-apisix
|   |   |-- README.md
|   |   |-- apisix
|   |   |   |-- bcs-auth
|   |   |   |   |-- authentication.lua
|   |   |   |   |-- bklogin.lua
|   |   |   |   |-- jwt.lua
|   |   |   |   `-- mock-bklogin.lua
|   |   |   |-- bcs-auth.lua
|   |   |   |-- bcs-common
|   |   |   |   `-- upstreams.lua
|   |   |   |-- bcs-dynamic-route.lua
|   |   |        |-- favicon-32x32.png
|   |           |-- index.html
|   |           |-- oauth2-redirect.html
|   |           |-- swagger-ui-bundle.js
|   |           |-- swagger-ui-bundle.js.map
|   |           |-- swagger-ui-standalone-preset.js
|   |           |-- swagger-ui-standalone-preset.js.map
|   |           |-- swagger-ui.css
|   |           |-- swagger-ui.css.map
|   |           |-- swagger-ui.js
|   |           `-- swagger-ui.js.map
|   |-- bcs-storage
|   |   |-- Dockerfile
|   |   |-- bcs-storage
|   |   |-- bcs-storage.json.template
|   |   |-- container-start.sh
|   |   |-- queue.conf.template
|   |   `-- storage-database.conf.template
|   |-- bcs-user-manager
|   |   |-- Dockerfile
|   |   |-- bcs-user-manager
|   |   |-- bcs-user-manager.json.template
|   |   `-- container-start.sh
|   `-- cryptools   |-- bkbcs-auth
|   |   |   |   `-- bkbcs.lua
|   |   |   `-- bkbcs-auth.lua
|   |   |-- apisix-start.sh
|   |   |-- bcs-gateway-discovery
|   |   |-- bcs-gateway-discovery.json.template
|   |   |-- config.yaml.template
|   |   `-- container-start.sh
|   |-- bcs-helm-manager
|   |   |-- Dockerfile
|   |   |-- bcs-helm-manager
|   |   |-- bcs-helm-manager-migrator
|   |   |-- container-start.sh
|   |   |-- lc_msgs.yaml
|   |   `-- swagger
|   |       `-- swagger-ui
|   |           |-- bcs-helm-manager.swagger.json
|   |           |-- favicon-16x16.png
|   |           |-- favicon-32x32.png
|   |           |-- index.html
|   |           |-- oauth2-redirect.html
|   |           |-- swagger-ui-bundle.js
|   |           |-- swagger-ui-bundle.js.map
|   |           |-- swagger-ui-standalone-preset.js
|   |           |-- swagger-ui-standalone-preset.js.map
|   |           |-- swagger-ui.css
|   |           |-- swagger-ui.css.map
|   |           |-- swagger-ui.js
|   |           `-- swagger-ui.js.map
|   |-- bcs-k8s-watch
|   |   |-- Dockerfile
|   |   |-- bcs-k8s-watch
|   |   |-- bcs-k8s-watch.json.template
|   |   |-- bcs-k8s-watch.yaml.template
|   |   |-- container-start.sh
|   |   `-- filter.json
|   |-- bcs-kube-agent
|   |   |-- Dockerfile
|   |   |-- bcs-kube-agent
|   |   |-- kube-agent-secret.yml
|   |   `-- kube-agent.yaml
|   |-- bcs-nodegroup-manager
|   |   |-- Dockerfile
|   |   |-- bcs-nodegroup-manager
|   |   |-- bcs-nodegroup-manager.json.template
|   |   `-- container-start.sh
|   |-- bcs-project-manager
|   |   |-- Dockerfile
|   |   |-- bcs-project-manager
|   |   |-- bcs-project-migration
|   |   |-- bcs-variable-migration
|   |   `-- swagger
|   |       |-- bcsproject.swagger.json
|   |       `-- swagger-ui
|   |           |-- favicon-16x16.png
|-- bcs-runtime
|   `-- bcs-k8s
|       |-- bcs-component
|       |   |-- bcs-apiserver-proxy
|       |   |   |-- Dockerfile
|       |   |   |-- bcs-apiserver-proxy
|       |   |   |-- bcs-apiserver-proxy-tools
|       |   |   |-- bcs-apiserver-proxy.json.template
|       |   |   |-- bcs-apiserver-proxy.yaml
|       |   |   `-- container-start.sh
|       |   |-- bcs-cluster-autoscaler
|       |   |   |-- Dockerfile
|       |   |   |-- bcs-cluster-autoscaler
|       |   |   `-- hyper
|       |   |       |-- bcs-cluster-autoscaler-1.16
|       |   |       `-- bcs-cluster-autoscaler-1.22
|       |   |-- bcs-external-privilege
|       |   |   |-- Dockerfile
|       |   |   `-- bcs-external-privilege
|       |   |-- bcs-general-pod-autoscaler
|       |   |   |-- Dockerfile
|       |   |   `-- bcs-general-pod-autoscaler
|       |   |-- bcs-image-loader
|       |   |   |-- Dockerfile
|       |   |   `-- bcs-image-loader
|       |   |-- bcs-k8s-custom-scheduler
|       |   |   |-- Dockerfile
|       |   |   |-- bcs-k8s-custom-scheduler
|       |   |   |-- bcs-k8s-custom-scheduler-kubeconfig.yaml
|       |   |   |-- bcs-k8s-custom-scheduler.manifest.template
|       |   |   `-- policy-config.json
|       |   |-- bcs-netservice-controller
|       |   |   |-- Dockerfile
|       |   |   |-- bcs-netservice-controller
|       |   |   |-- bcs-netservice-ipam
|       |   |   `-- bcs-underlay-cni
|       |   `-- bcs-webhook-server
|       |       |-- Dockerfile
|       |       |-- bcs-webhook-server
|       |       `-- container-start.sh
|       `-- bcs-network
|           `-- bcs-ingress-controller
|               |-- Dockerfile
|               |-- bcs-ingress-controller
|               `-- container-start.sh   
|-- bcs-scenarios
|   |-- bcs-gitops-manager
|   |   |-- Dockerfile
|   |   |-- bcs-gitops-manager
|   |   |-- bcs-gitops-manager.json.template
|   |   `-- container-start.sh
|   |-- bcs-gitops-proxy
|   |   |-- Dockerfile
|   |   |-- bcs-gitops-proxy
|   |   |-- bcs-gitops-proxy.json.template
|   |   `-- container-start.sh
|   `-- bcs-powertrading
|       |-- Dockerfile
|       |-- bcs-powertrading.json.template
|       `-- container-start.sh
```