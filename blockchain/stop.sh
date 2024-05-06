#!/bin/bash
# clean the chaincode container
function clearContainers() {
  CONTAINER_IDS=$(docker ps -a | awk '($2 ~ /dev-peer.*datashare.*/) {print $1}')
  if [ -z "$CONTAINER_IDS" -o "$CONTAINER_IDS" == " " ]; then
    echo "---- No containers available for deletion ----"
  else
    docker rm -f $CONTAINER_IDS
  fi
}

# clean the chaincode mirror
function removeUnwantedImages() {
  DOCKER_IMAGE_IDS=$(docker images | awk '($1 ~ /dev-peer.*datashare.*/) {print $3}')
  if [ -z "$DOCKER_IMAGE_IDS" -o "$DOCKER_IMAGE_IDS" == " " ]; then
    echo "---- No images available for deletion ----"
  else
    docker rmi -f $DOCKER_IMAGE_IDS
  fi
}


# clean the previous network
docker-compose -f explorer/docker-compose.yaml down -v
docker-compose -f docker-compose-byfn.yaml down -v

# clean the previous certificate
rm -rf channel-artifacts
rm -rf crypto-config

# clean all containers
clearContainers
removeUnwantedImages

