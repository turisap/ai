https://roadmap.sh/kubernetes

Day 1 — Images & layers (mental model)
Objective: Understand what a Docker image actually is before touching Dockerfiles.

Read how images are built from layers, how the union filesystem works, why layer caching matters for build speed
Pull and inspect a few images: docker pull golang:1.23, docker history golang:1.23, docker inspect
Understand image vs container (image = class, container = instance — this will click fast for you)
https://docs.docker.com/get-started/docker-concepts/building-images/understanding-image-layers/