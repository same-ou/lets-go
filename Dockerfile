FROM golang:1.23.3-bullseye AS build

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY . .

RUN go build \
    -ldflags="-linkmode external -extldflags -static" \
    -tags netgo \    
    -o ./lets-go ./cmd/web

FROM scratch
COPY --from=build /app/tls ./tls
COPY --from=build /app/ui ./ui
COPY --from=build /app/schema.sql ./schema.sql
COPY --from=build /app/lets-go lets-go

EXPOSE 4000

CMD [ "/lets-go" ]