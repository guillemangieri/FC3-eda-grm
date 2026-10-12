# EDA — Wallet Core e Balances

Projeto do desafio de Arquitetura Orientada a Eventos da Full Cycle. O repositório reúne o Wallet Core, que produz eventos de transação, e o microsserviço Balances, que consome os eventos pelo Kafka, mantém uma projeção própria dos saldos e disponibiliza uma API de consulta.

## Pré-requisitos

- Docker com o plugin Docker Compose.
- Portas `3003`, `8080`, `3307`, `9092` e `2181` disponíveis.
- Aproximadamente 4 GB de memória disponíveis para os containers.

Não é necessário instalar Go, MySQL ou Kafka na máquina.

## Executar

Na raiz do repositório, execute:

```bash
docker compose up -d
```

O comando constrói as aplicações, inicia os dois bancos, ZooKeeper e Kafka, aplica migrations, insere os seeds e cria os tópicos `transactions` e `balances`. A primeira execução pode demorar alguns minutos enquanto as imagens são baixadas e os serviços passam pelos healthchecks.

Confira o estado dos serviços:

```bash
docker compose ps
```

Quando `goapp` e `balances` estiverem em execução, as APIs estarão disponíveis em:

| Serviço | Endereço | Finalidade |
|---|---|---|
| Wallet Core | `http://localhost:8080` | Criação de clientes, contas e transferências |
| Balances | `http://localhost:3003` | Consulta da projeção de saldos |
| MySQL Wallet | `localhost:3307` | Banco próprio do Wallet Core |
| Kafka | `localhost:9092` | Broker acessível pela máquina hospedeira |
| Control Center | `http://localhost:9021` | Interface opcional do Kafka |

O Control Center não inicia no comando padrão. Para habilitá-lo:

```bash
docker compose --profile tools up -d
```

## Dados iniciais

As migrations inserem duas contas conhecidas nos bancos Wallet e Balances:

| Titular | Client ID | Account ID | Saldo inicial |
|---|---|---|---:|
| Alice | `11111111-1111-4111-8111-111111111111` | `aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa` | 1000.00000000 |
| Bob | `22222222-2222-4222-8222-222222222222` | `bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb` | 500.00000000 |

Os scripts são idempotentes: reiniciar o ambiente não duplica os seeds nem redefine saldos já movimentados.

## Testar o fluxo

O arquivo [api.http](api.http) contém o roteiro completo para clientes HTTP compatíveis com arquivos `.http`, como o REST Client do VS Code e o cliente HTTP do IntelliJ.

Primeiro, consulte os saldos iniciais:

```http
GET http://localhost:3003/balances/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa
GET http://localhost:3003/balances/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb
```

Depois, transfira 100 de Alice para Bob pelo Wallet Core:

```http
POST http://localhost:8080/transactions
Content-Type: application/json

{
  "account_id_from": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  "account_id_to": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  "amount": 100
}
```

Após o consumo assíncrono, as consultas do Balances devem retornar Alice com `900.00000000` e Bob com `600.00000000`.

## Arquitetura e consistência

```text
Cliente -> Wallet Core -> MySQL Wallet
                  |
                  v
                Kafka (tópico balances)
                  |
                  v
         Balances Consumer -> MySQL Balances -> GET /balances/{account_id}
```

O Wallet Core publica o evento `BalanceUpdated` no tópico `balances`. O consumidor usa o grupo `balances-service` e persiste os saldos absolutos das duas contas junto com o offset processado em uma transação no seu próprio banco. O offset Kafka só é confirmado depois dessa persistência.

A API do Balances consulta exclusivamente o banco do próprio serviço. Como o processamento é assíncrono, uma consulta imediatamente após a transferência pode mostrar o saldo anterior por alguns instantes. Repita a consulta até o evento ser consumido.

O tópico possui uma partição para preservar a ordem dos snapshots. O processamento SQL é idempotente por tópico, partição e offset.

## Respostas da API Balances

Consulta de conta existente:

```json
{
  "account_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  "balance": 1000.00000000
}
```

- `200 OK`: saldo encontrado.
- `404 Not Found`: a conta não possui projeção no Balances.
- `405 Method Not Allowed`: método diferente de GET.
- `500 Internal Server Error`: banco do Balances indisponível ou erro interno.

## Logs e encerramento

Acompanhar os produtores e consumidores:

```bash
docker compose logs -f goapp balances
```

Consultar todos os serviços:

```bash
docker compose logs --tail 200
```

Parar e remover os containers, preservando os dados:

```bash
docker compose down
```

Reiniciar os serviços preservando os dados:

```bash
docker compose up -d
```

## Apagar os dados e reiniciar do zero

Os comandos desta seção apagam os dados persistidos. Use-os somente quando quiser voltar aos seeds originais.

Primeiro remova containers e volumes nomeados do Balances, Kafka e ZooKeeper:

```bash
docker compose down -v
```

O banco Wallet usa o diretório local `.docker/mysql/data`, que não é removido por `down -v`. Com os containers parados, apague esse diretório manualmente.

PowerShell:

```powershell
Remove-Item -LiteralPath .docker/mysql/data -Recurse -Force
```

Bash:

```bash
rm -rf .docker/mysql/data
```

Em seguida, execute novamente `docker compose up -d` para recriar schemas e seeds automaticamente.

## Estrutura principal

- `cmd/walletcore`: inicialização do Wallet Core.
- `cmd/balances`: inicialização da API e do consumidor Balances.
- `internal/balances`: consumo Kafka, persistência e endpoint HTTP.
- `migrations/wallet`: schema e seeds do Wallet Core.
- `migrations/balances`: schema e seeds do Balances.
- `docker`: scripts automáticos de migrations e criação dos tópicos.
- `api.http`: roteiro de teste do desafio.

