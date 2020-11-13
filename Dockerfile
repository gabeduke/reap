FROM golang:1.14-alpine AS build

WORKDIR /src/
COPY main.go go.* /src/
RUN CGO_ENABLED=0 go build -o /bin/reap

FROM scratch
COPY --from=build /bin/reap /bin/reap
ENTRYPOINT ["/bin/reap"]
