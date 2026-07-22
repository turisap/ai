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

#### Concepts

* The ReplicaSet's job is simpler and narrower: just "keep exactly N pods matching this one specific template alive.".
  if u change pod's image, there will be a new replica set created and upscaled, the old one will be down scaled
