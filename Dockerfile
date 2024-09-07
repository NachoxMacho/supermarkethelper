FROM golang:1.22.5 AS build-stage

WORKDIR /app
RUN go install github.com/a-h/templ/cmd/templ@latest

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# TODO: just copy .go files while keeping structure
COPY . ./

# Disable CGO so we can run without glibc
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-go

FROM node:18 AS tailwindcss

WORKDIR /app

COPY ./package.json ./
COPY ./package-lock.json ./
COPY ./views ./

RUN npm ci
RUN npx tailwindcss -i /app/css/styles.css -o /app/styles.css

FROM gcr.io/distroless/base-debian12:debug  AS release-stage

WORKDIR /app


COPY --from=build-stage /docker-go /app/docker-go

COPY ./public /app/public
COPY --from=tailwindcss /app/styles.css /app/public/styles.css
COPY ./migrations /app/migrations

EXPOSE 42069

# Figure out a way to re-enable this, so we aren't running as the root user
# USER nonroot:nonroot

ENTRYPOINT ["/app/docker-go"]

