# Copyright 2024 Daniel Moch, all rights reserved
FROM golang:1.21-alpine AS build

RUN mkdir -p djmo.ch/dgit

COPY ./ /djmo.ch/dgit/

RUN cd /djmo.ch/dgit/cmd/dgit && CGO_ENABLED=0 go build ./

FROM scratch

COPY --from=build /djmo.ch/dgit/cmd/dgit/dgit /usr/local/bin/

CMD ["/usr/local/bin/dgit"]

USER 1001
