# BUILD CONTAINER
FROM golang:1.15-alpine AS build_go
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /src
COPY . /src
RUN go build -o /out/b618reboot .

#############################
# RUN CONTAINER
FROM gcr.io/distroless/base
ENV ROUTER_URL=http://192.168.1.1
ENV ROUTER_USERNAME=admin
ENV ROUTER_PASSWORD=""

COPY --from=build_go /out/b618reboot .

ENTRYPOINT ["./b618reboot"]  
CMD ["help"]