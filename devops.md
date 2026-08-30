https://roadmap.sh/kubernetes

Day 1 — Images & layers (mental model)
Objective: Understand what a Docker image actually is before touching Dockerfiles.

Read how images are built from layers, how the union filesystem works, why layer caching matters for build speed
Pull and inspect a few images: docker pull golang:1.23, docker history golang:1.23, docker inspect
Understand image vs container (image = class, container = instance — this will click fast for you)
https://docs.docker.com/get-started/docker-concepts/building-images/understanding-image-layers/

### layers tutorial:

* `docker run --name=base-container -ti ubuntu`
* `apt update && apt install -y nodejs`
* `node -e 'console.log("Hello world!")'`
* `docker container commit -m "Add node" base-container node-base` - commit changes to the initial image
* `docker image history node-base` - check history

### volumes and bind mounts

* bind mounts are for development, volumes are for production
* `docker volume ls`
* `docker volume inspect mydata`

```yaml
services:
  service-1:
    image: ubuntu:latest
    command: sleep infinity
    volumes:
      - ./code:/app          # Bind mount for development
      - shared:/data         # Named volume shared with worker

  service-2:
    image: ubuntu:latest
    command: sleep infinity
    volumes:
      - shared:/data         # Same volume as service-1

volumes:
  shared:
```

* network create `docker network create demo-network -d bridge`

### security

* if u `COPY . .` then `.env` files are also copied and baked into the image - insecure. use `.dockeringore`
* test if u can access `sh` in the container

```shell
docker images | grep mcp
docker run --rm --entrypoint sh mcp -c "echo test" 2>&1
```

### k8s architecture

#### Node components

* `kube-proxy`: works on each worker node (network proxy for components like ServiceAPI). kube-proxy is a network proxy
  that runs on each node in your cluster, implementing part of the Kubernetes Service concept.
  kube-proxy maintains network rules on nodes.
* `kubelet`: An agent that runs on each node in the cluster. It makes sure that containers are running in a Pod. The
  kubelet takes
  a set of PodSpecs that are provided through various mechanisms and ensures that the containers described in those
  PodSpecs are running and healthy.

#### Control pane

* `kube-apiserver`. The API server is a component of the Kubernetes control plane that exposes the Kubernetes API. The
  API server is the
  front end for the Kubernetes control plane.
* `etcd`.Consistent and highly-available key value store used as Kubernetes' backing store for all cluster data.
* `kube-scheduler` Control plane component that watches for newly created Pods with no assigned node, and selects a node
  for them to run on.
* `kube-controller-manager`. Control plane component that runs controller processes. Logically, each controller is a
  separate process, but to reduce complexity, they are all compiled into a single binary and run in a single process.

#### First steps

* `kind version  && kubectl version --client`
* `kind create cluster --name learning`
* check `kubectl cluster-info --context kind-learning && kubectl get nodes`
* docker ps will give u container with the whole cluster in it
* `docker exec -it learning-control-plane crictl ps` shows all k8s elements inside that single container

#### Commands

* `kubectl delete pod nginx-test`
* `kubectl exec -it nginx-test -- sh` -- @COOL exec into a pod
* `kubectl logs nginx-test` -- @COOL logs
* `kubectl get pods -w` - @COOL get pods watch (watch restarts)
* `kubectl get rs` - get replica sets
* `kubectl get pods -l app=nginx-deploy` - get deployment's pods
* `kubectl describe deploy nginx-deploy` - get deployment info
* `kubectl get nginx-svc` - get service with its IP
* `kubectl describe pod mcp-task-server-8455c7b476-rlfqw` - pod info
* `kubectl get configmap mcp-task-server-config -o yaml` get config map

#### Concepts

* The ReplicaSet's job is simpler and narrower: just "keep exactly N pods matching this one specific template alive.".
  if u change pod's image, there will be a new replica set created and upscaled, the old one will be down scaled
*

```
  Type	Reachable from	Typical use
  ClusterIP (default)	Inside cluster only	Internal service-to-service traffic — e.g. mcp-task-server → Postgres
  NodePort	Any node's IP at a static high port (30000-32767)	Rarely used directly in prod; mostly a building block
  LoadBalancer	External internet, via cloud provider's LB	Public-facing services — this is what Yandex Cloud provisions when you set this type on a real cluster
``` 

* docs `kubectl explain deployment.spec.strategy.rollingUpdate `

#### Sealed secrets

* `brew install kubeseal`
* `kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.27.1/controller.yaml`
* `kubectl get pods -n kube-system -l name=sealed-secrets-controller` confirm running
reseal
```
kubeseal --format yaml --namespace mcp-dev < k8s/mcp-secret.yaml > k8s/mcp-sealed-secret.yaml
kubectl apply -f k8s/mcp-sealed-secret.yaml
```

#### Namespaces

What they actually are: a way to partition objects within a single cluster into logically separate groups — think of it
as a folder structure for cluster resources, not a hard security or performance boundary.

* `kubectl create namespace mcp-dev`
* `kubectl get pods -n mcp-dev`
* `kubectl config set-context --current --namespace=mcp-dev` - @COOL sets the default namespace

#### Storage

PersistentVolume (PV): a piece of actual storage in the cluster — could be a local disk, a cloud disk (Yandex Cloud has
its own CSI driver for this), an NFS share. Cluster-scoped, not namespaced.
PersistentVolumeClaim (PVC): a namespaced request for storage — "I need 5Gi, ReadWriteOnce" — that gets matched/bound to
an available PV.

* `kubectl get pvc -n mcp-dev`
* `kubectl get pv`

### Troubleshooting

* `kubectl run test --image=busybox -it --rm -n default -- sh` - busybox image with all linux commands

#### Wrong image (old image for some reason running in the pod)

* `docker images mcp:latest --no-trunc` your image SHA

```shell
mcp          latest    sha256:b7fe87558efd21d1e36a431c9147884d1f07751f4fea1af63181a612643d9fb4   6 minutes ago   17.9MB
```

* `docker exec -it learning-control-plane crictl images | grep mcp`
  what is kind running

```
docker.io/library/import-2026-07-31@sha256:8a732554be26cdf44acfe277473c040a8c0f6c7199ea05e2b89cdeb3ec1d0c4b%
```

if there is a mismatch - `docker build --no-cache -t mcp:latest -f docker/Dockerfile .`
`kind load docker-image mcp:latest --name learning`
`kubectl rollout restart deployment/mcp-task-server -n mcp-dev
kubectl rollout status deployment/mcp-task-server -n mcp-dev`

#### Cleanup habit

* check all resources in the namespace `kubectl get all -n mcp-dev`
* `kubectl get pods -n default | grep mcp` check whether there are pods for mpc in the default namespace (conflicting)

#### Quota exceeded

* `kubectl edit resourcequota mcp-dev-quota -n mcp-dev` - edit in place or via `kubectl apply` to your quota yaml
* running a test pod from a different namespace with no quota
  `kubectl run dns-test --image=busybox -it --rm -n default -- nslookup postgres.mcp-dev.svc.cluster.local`

#### Ingress controller (nginx)

*
`kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml`
* `kubectl get all -n ingress-nginx`
* `kubectl get pods -n ingress-nginx -w`
* `kubectl get namespace ingress-nginx`
* rewrite target (like krakend backends) ```shell
  No rewrite annotation /api/healthz /api/healthz (unchanged)
  rewrite-target: / (plain)    /api/healthz / (everything after prefix discarded)
  Capture-group regex + rewrite-target: /$2 /api/healthz /healthz (prefix stripped, rest preserved)

```
annotations:
nginx.ingress.kubernetes.io/rewrite-target: /$2
...
- path: /api(/|$)(.*)
  pathType: ImplementationSpecific
```

#### Helm

* prep
    * `brew install helm && helm version`
    * `helm create mcp-task-server-chart && tree mcp-task-server-chart`
* @TODO @COOL `helm template mcp-task-server-chart` - see what yaml will be produced by rendering
* render templates with values `helm template mcp-task-server-chart -f mcp-task-server-chart/values-dev.yaml --debug 2>&1 | head -30`