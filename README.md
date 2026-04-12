# myfin - CLI de Gestão de Finanças Pessoais

CLI para gerenciamento de finanças pessoais com armazenamento em arquivos YAML, suporte a múltiplas contas, cartões de crédito com rastreamento de faturas, categorias, tags e parsing de expressões matemáticas para valores.

## Pré-requisitos

- **Go 1.26+**

## Instalação

### 1. Clone o repositório

```bash
git clone https://github.com/Sanmoo/my-finances.git
cd my-finances
```

### 2. Build

```bash
go build -o myfin ./cmd/
```

O binário `myfin` será criado no diretório atual.

### 3. (Opcional) Mova para o PATH

```bash
sudo mv myfin /usr/local/bin/
```

## Configuração

Configure padrões no arquivo `~/.myfin.yaml`:

```yaml
data.path: ~/.myfin/data
default.currency: BRL
locale: pt-BR
```

O arquivo usa os seguintes defaults se não existir:
- `data.path`: `~/.myfin/data`
- `default.currency`: `BRL`
- `locale`: `pt-BR`

## Uso Rápido

### Criar uma conta

```bash
myfin add account checking
myfin add account savings
```

### Criar tags

```bash
myfin add tag mensal
myfin add tag assinatura
```

### Criar categorias

```bash
# Categoria de despesa
myfin add category "Transporte" --account checking --type exp --alias transp --emoji 🚗

# Categoria de receita
myfin add category "Salário" --account checking --type inc --alias sal --emoji 💰
```

### Criar cartão de crédito

```bash
myfin add credit-card nubank --closing-day 9 --due-day 16
```

### Adicionar despesas

```bash
# Despesa simples
myfin add expense 45.50 --account checking --category transp --date 26-03-21 --description "Uber"

# Despesa parcelada no cartão
myfin add expense 300.00 --account checking --credit-card nubank --times 6 --category eletro --date 26-03-10 --description "TV"

# Com tags (tags devem ser previamente registradas)
myfin add expense 120.00 --account checking --tags mensal,assinatura --category servicos --date 26-03-15 --description "Spotify"
```

### Adicionar receitas

```bash
# Receita simples
myfin add income 5000.00 --account checking --category sal --date 26-03-01 --description "Salário"

# Com expressão matemática
myfin add income 3000+1500 --account checking --category renda-extra --date 26-03-15 --description "Freelance"
```

### Relatórios

```bash
# Listar lançamentos
myfin report entries --account checking --from 26-03-01 --until 26-03-31

# Listar lançamentos em formato markdown
myfin report entries --account checking --from 26-03-01 --until 26-03-31 --format md

# Filtrar por tags e categorias
myfin report entries --account checking --filter-tags mensal --filter-categories transp

# Ver saldos por conta
myfin report balances --account checking

# Ver saldos em formato markdown
myfin report balances --account checking --format md

# Agrupar por categoria
myfin report by-category --account checking --from 26-03-01 --until 26-03-31

# Listar tags registradas
myfin list tags
```

## Referência de Comandos

Use `myfin --help` e `myfin <comando> --help` para ver todas as flags disponíveis.

| Comando | Descrição |
|---------|-----------|
| `myfin add account <nome>` | Cria uma conta |
| `myfin add tag <nome>` | Registra uma tag |
| `myfin add category <nome>` | Cria uma categoria (`--account`, `--type`, `--alias` obrigatórios) |
| `myfin add credit-card <nome>` | Cria um cartão de crédito (`--closing-day`, `--due-day` obrigatórios) |
| `myfin add expense [valor]` | Adiciona uma despesa (`--account`, `--date`, `--description` obrigatórios) |
| `myfin add income [valor]` | Adiciona uma receita (`--account`, `--date`, `--description` obrigatórios) |
| `myfin list tags` | Lista todas as tags registradas |
| `myfin report entries` | Lista lançamentos com filtros |
| `myfin report balances` | Mostra saldos por conta |
| `myfin report by-category` | Agrupa lançamentos por categoria |

### Flags de data

O formato de data é flexível:
- `DD` - dia do mês atual (ex: `15`)
- `MM-DD` - mês e dia do ano atual (ex: `03-15`)
- `YY-MM-DD` - ano abreviado (ex: `26-03-15`)
- `YYYY-MM-DD` - data completa (ex: `2026-03-15`)

## Expressões Matemáticas

O myfin suporta expressões matemáticas nos valores:

```bash
myfin add expense 1000/3 --account checking --category compras  # 333.33
myfin add income 5000+1000 --account checking --category bonus  # 6000
myfin add expense (100+50)*2 --account checking --category x   # 300
```

Operadores suportados: `+`, `-`, `*`, `/` com precedência padrão e suporte a parênteses.

## Lógica de Fatura de Cartão de Crédito

Para um cartão com `closing-day=9` e `due-day=16`:

- Despesas realizadas nos dias **1-9** → vencimento no dia **16 do mesmo mês**
- Despesas realizadas nos dias **10-31** → vencimento no dia **16 do mês seguinte**

Exemplo com `--times 3`:

```bash
myfin add expense 300 --account checking --credit-card nubank --times 3 --date 26-03-15 --description "Compra"
```

Isso cria 3 parcelas:
1. Realização: 15/03, Vencimento: 16/04
2. Realização: 15/04, Vencimento: 16/05
3. Realização: 15/05, Vencimento: 16/06

## Estrutura de Dados

Os dados são armazenados em arquivos YAML no diretório configurado (`data.path`):

```
~/.myfin/data/
├── accounts.yaml              # Lista de contas
├── credit_cards.yaml          # Cartões de crédito
├── tags.yaml                  # Tags registradas
├── checking/
│   ├── categories.yaml        # Categorias da conta
│   └── 2026/
│       └── 2026-03-checking-entries.yaml  # Lançamentos de março
└── savings/
    └── ...
```

Cada arquivo de lançamentos contém entradas do tipo `income` ou `expense` com data de realização e data de pagamento (quando aplicável).

## Arquitetura

O projeto segue **Clean Architecture** com **Ports and Adapters**:

```
myfin/
├── cmd/                    # Entry point (Cobra CLI)
├── internal/
│   ├── domain/entity/      # Entidades de domínio (pura lógica de negócio)
│   ├── core/
│   │   ├── usecase/        # Casos de uso
│   │   └── port/           # Interfaces dos repositórios
│   └── infrastructure/
│       ├── persistence/    # Implementações YAML
│       ├── cli/            # Formatadores de saída
│       ├── config/         # Loader de configuração
│       └── i18n/           # Internacionalização

└── pkg/expr/               # Parser de expressões matemáticas
```

### Camadas

1. **Domain** (`internal/domain/entity/`) - Regras de negócio puras, sem dependências externas
2. **Core** (`internal/core/`) - Casos de uso e interfaces (ports)
3. **Infrastructure** (`internal/infrastructure/`) - Implementações concretas (adapters)
4. **CLI** (`cmd/`) - Interface de linha de comando

## Desenvolvimento

### Rodar testes

```bash
go test ./...
```

### Ver cobertura

```bash
go test ./... -cover
```

## Licença

MIT
