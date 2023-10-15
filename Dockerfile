FROM golang:1.20-alpine AS build

WORKDIR /src/
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/reap

FROM scratch
COPY --from=build /bin/reap /bin/reap
ENTRYPOINT ["/bin/reap"]
