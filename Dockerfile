FROM golang:alpine AS base

WORKDIR /app
COPY . .
RUN go mod download && CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build

FROM alpine as runner

ENV ENV=dev
WORKDIR /app
RUN apk update && apk add --no-cache git openssh && mkdir /root/.ssh/ && ssh-keyscan github.com >> ~/.ssh/known_hosts
COPY --from=base /app/github_pull /app/github_pull
EXPOSE 80
CMD ./github_pull --config_file=config/${ENV}.yaml