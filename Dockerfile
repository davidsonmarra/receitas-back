# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Instalar dependências de build
RUN apk add --no-cache git

WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Download de dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Build otimizado para produção
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o api ./cmd/api

# Stage 2: Runtime
FROM alpine:latest

# Instalar certificados CA (necessário para HTTPS)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar binário do stage de build
COPY --from=builder /app/api .

# Expor porta (Railway/Heroku definem via env var PORT)
EXPOSE 8080

# Variáveis de ambiente padrão (podem ser sobrescritas)
ENV ENV=production
ENV LOG_LEVEL=info

# Executar aplicação
CMD ["./api"]

