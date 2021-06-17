FROM golang:1.16-alpine AS build

WORKDIR /app
RUN apk add --no-cache git
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o main ticker/src/main.go

FROM gcr.io/distroless/static

COPY --from=build /app/main /main
ENV DATA_DIR=/storage/data
ENTRYPOINT [ "/main" ]
