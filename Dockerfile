# Copyright 2021 Kirill Scherba <kirill@scherba.ru>.  All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
#
# Teonet docker file
#
# Build:
#
#  docker build -t teomon .
#
# Publish to github:
#
#  docker login docker.pkg.github.com -u USERNAME -p TOKEN
#  docker tag teomon docker.pkg.github.com/kirill-scherba/teonet-go/teomon:0.0.3
#  docker push docker.pkg.github.com/kirill-scherba/teonet-go/teomon:0.0.3
#
# Publish to local repository:
#
#  docker tag teonet 192.168.106.5:5000/teonet
#  docker push 192.168.106.5:5000/teonet
#
# Run docker container:
#
#  docker run --rm -it teomon
#
# Run in swarm claster:
#
#  docker volume create teonet-config
#  docker service create --constraint 'node.hostname == teonet'   --network teo-overlay --hostname=teo-go-01 --name teo-go-01 --mount type=volume,source=teonet-config,target=/root/.config/teonet 192.168.106.5:5000/teonet-go teonet -a 5.63.158.100 -r 9010 -n teonet teo-go-01
#  docker service create --constraint 'node.hostname == dev-ks-2' --network teo-overlay --hostname=teo-go-02 --name teo-go-02 --mount type=volume,source=teonet-config,target=/root/.config/teonet 192.168.106.5:5000/teonet-go teonet -a 5.63.158.100 -r 9010 -n teonet teo-go-02
#
# Or update existing service in swarm claster:
#
#  docker service update --image 192.168.106.5:5000/teonet-go teonet-go
#


#
# temporary wail this repos is private use next command to build image:
#
#   docker build -t teomon -f ./Dockerfile ../.
#
# it recomendet use host network when run teomon  
#
#   docker run --restart always -it --name teomon --network host docker.pkg.github.com/kirill-scherba/teonet-go/teomon:0.2.0 teomon -u
#
#

# Docker builder
# 
FROM golang:1.17.7 AS builder

WORKDIR /go/src/github.com/kirill-scherba/
RUN apt update 
# && apt install -y libssl-dev

COPY ./teonet ./teonet
COPY ./teomon ./teomon
COPY ./tru ./tru

RUN ls /go/src/github.com/kirill-scherba/

WORKDIR /go/src/github.com/kirill-scherba/teomon

RUN go install ./cmd/teomon

RUN ls /go/bin

CMD ["teomon"]

# #############################################################
# Compose production image
#
FROM ubuntu:latest AS production
WORKDIR /app

# runtime dependencies
RUN apt update 

# install ssl cretificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# install previously built application
COPY --from=builder /go/bin/* /usr/local/bin/

CMD ["/usr/local/bin/teomon"]  
