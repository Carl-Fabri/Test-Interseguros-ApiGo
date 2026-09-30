# ---- Build ----
FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---- Runtime ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/api /app/api

# ! Puerto 8080 (no 80): la imagen corre como usuario no root y Linux no permite a un usuario no root
# ! abrir puertos < 1024. En ECS Express Mode configura containerPort=8080 (ver docs/deploy/aws.md).
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
