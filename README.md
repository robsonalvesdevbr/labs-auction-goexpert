# Labs Auction — Go Expert

Sistema de leilões com fechamento automático via Goroutines.

## Funcionalidade implementada

Ao criar um leilão, uma Goroutine é iniciada em background. Quando o tempo configurado em `AUCTION_DURATION` expira, o status do leilão é atualizado automaticamente para `Completed` no MongoDB — sem nenhuma intervenção manual.

## Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) e [Docker Compose](https://docs.docker.com/compose/)
- Go 1.20+ (apenas para rodar testes localmente)

## Como rodar com Docker

```bash
docker compose up --build
```

A API ficará disponível em `http://localhost:8080`.

Para derrubar os containers:

```bash
docker compose down -v
```

## Variáveis de ambiente

Todas as variáveis ficam em `cmd/auction/.env`:

| Variável | Descrição | Exemplo |
|---|---|---|
| `AUCTION_DURATION` | Tempo até o leilão ser fechado automaticamente | `20s`, `5m`, `1h` |
| `AUCTION_INTERVAL` | Intervalo do batch de lances | `20s` |
| `MAX_BATCH_SIZE` | Tamanho máximo do batch de lances | `4` |
| `MONGODB_URL` | URL de conexão com o MongoDB | `mongodb://admin:admin@mongodb:27017/auctions?authSource=admin` |
| `MONGODB_DB` | Nome do banco de dados | `auctions` |

### Configurando a duração do leilão

Edite `cmd/auction/.env` e ajuste `AUCTION_DURATION`:

```env
# Fechar após 20 segundos
AUCTION_DURATION=20s

# Fechar após 5 minutos
AUCTION_DURATION=5m

# Fechar após 1 hora
AUCTION_DURATION=1h
```

O formato aceito é o padrão do Go: `s` (segundos), `m` (minutos), `h` (horas). Se o valor for inválido ou ausente, o fallback é **5 minutos**.

## Endpoints da API

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/auction` | Cria um leilão |
| `GET` | `/auction` | Lista leilões (filtros: `status`, `category`, `productName`) |
| `GET` | `/auction/:auctionId` | Busca leilão por ID |
| `GET` | `/auction/winner/:auctionId` | Retorna o lance vencedor |
| `POST` | `/bid` | Cria um lance |
| `GET` | `/bid/:auctionId` | Lista lances de um leilão |
| `GET` | `/user/:userId` | Busca usuário por ID |

### Exemplo de criação de leilão

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "Notebook Dell",
    "category": "Electronics",
    "description": "Notebook Dell Inspiron 15 polegadas, 16GB RAM, SSD 512GB",
    "condition": 1
  }'
```

Valores de `condition`: `1` = Novo, `2` = Usado, `3` = Recondicionado.

## Rodando os testes

Os testes de integração requerem MongoDB rodando:

```bash
# Subir apenas o MongoDB
docker compose up -d mongodb

# Rodar os testes
MONGODB_URL="mongodb://admin:admin@localhost:27017/auction?authSource=admin" \
MONGODB_DB="auction_test" \
AUCTION_DURATION="3s" \
go test -v ./internal/infra/database/auction/... -timeout 30s
```

O teste cria um leilão com duração de 3 segundos, aguarda 4 segundos e verifica que o status foi alterado para `Completed` no banco de dados.
