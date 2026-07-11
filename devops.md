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