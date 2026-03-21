# myfin - CLI de Gestão de Finanças Pessoais

CLI para gerenciamento de finanças pessoais com suporte a múltiplos namespaces (contextos isolados), cartões de crédito com rastreamento de faturas, categorias, tags e parsing de expressões matemáticas para valores.

## Pré-requisitos

- **Go 1.26+**
- **GCC** (necessário para compilar o driver SQLite3)

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

Ao executar pela primeira vez, o myfin cria um namespace padrão `default`. Você pode configurar padrões no arquivo `~/.myfin.yaml`:

```yaml
default.namespace: main
default.currency: BRL
default.credit-card: nubank
```

O arquivo é criado automaticamente na primeira configuração.

## Uso Rápido

### Criar uma conta bancária

```bash
myfin add account main
myfin add account savings
```

### Criar categorias

```bash
# Categoria de despesa
myfin add category --type exp --name "Transporte" --alias transp --emoji 🚗

# Categoria de receita
myfin add category --type inc --name "Salário" --alias sal --emoji 💰
```

### Criar cartão de crédito

```bash
myfin add credit-card nubank --closing-day 9 --due-day 16
```

### Adicionar despesas

```bash
# Despesa simples
myfin add expense 45.50 --category transporte --date 21-03-26 --description "Uber"

# Despesa parcelada no cartão
myfin add expense 300.00 --credit-card nubank --times 6 --category eletro --date 21-03-10

# Com tags
myfin add expense 120.00 --tags mensal,assinatura --category servicos
```

### Adicionar receitas

```bash
# Receita simples
myfin add income 5000.00 --category salario --date 26-03-01

# Com expressão matemática
myfin add income 3000+1500 --category renda-extra --description "Freelance"
```

### namespaces

Por padrão, usa o namespace `default`. Especifique outro com `-s`:

```bash
myfin -s main add expense 100.00 --category comida
```

## Referência de Comandos

### `myfin add`

| Comando | Descrição |
|---------|-----------|
| `myfin add account <nome>` | Cria uma conta bancária |
| `myfin add category [flags]` | Cria uma categoria |
| `myfin add credit-card <nome> [flags]` | Cria um cartão de crédito |
| `myfin add expense [valor]` | Adiciona uma despesa |
| `myfin add income [valor]` | Adiciona uma receita |

### Flags do `add category`

| Flag | Descrição |
|------|-----------|
| `-t, --type <inc\|exp>` | Tipo da categoria (income ou expense) |
| `-n, --name <nome>` | Nome da categoria |
| `--alias <alias>` | Apelido alternativo |
| `--emoji <emoji>` | Emoji para exibição |

### Flags do `add credit-card`

| Flag | Descrição |
|------|-----------|
| `--closing-day <n>` | Dia do fechamento da fatura (1-31) |
| `--due-day <n>` | Dia do vencimento (1-31) |

### Flags do `add expense` e `add income`

| Flag | Descrição |
|------|-----------|
| `--date <DD-MM-YY>` | Data de realização |
| `--category <nome>` | Nome ou alias da categoria |
| `--description <texto>` | Descrição da transação |
| `--tags <x,y,z>` | Tags separadas por vírgula |

### Flags específicas do `add expense`

| Flag | Descrição |
|------|-----------|
| `--credit-card <nome>` | Cartão de crédito (requer `--times`) |
| `--times <n>` | Número de parcelas |

## Expressões Matemáticas

O myfin suporta expressões matemáticas nos valores:

```bash
myfin add expense 1000/3 --category compras  # 333.33
myfin add income 5000+1000 --category bonus  # 6000
myfin add expense (100+50)*2 --category x   # 300
```

Operadores suportados: `+`, `-`, `*`, `/` com precedência padrão.

## Lógica de Fatura de Cartão de Crédito

Para um cartão com `closing-day=9` e `due-day=16`:

- Despesas realizadas nos dias **1-9** → vencimento no dia **16 do mesmo mês**
- Despesas realizadas nos dias **10-31** → vencimento no dia **16 do mês seguinte**

Exemplo com `--times 3`:

```bash
myfin add expense 300 --credit-card nubank --times 3 --date 26-03-15
```

Isso cria 3 parcelas:
1. Realização: 15/03, Vencimento: 16/04
2. Realização: 15/04, Vencimento: 16/05
3. Realização: 15/05, Vencimento: 16/06

## Arquitetura

O projeto segue **Clean Architecture** com **Ports and Adapters**:

```
myfin/
├── cmd/                    # Entry point (Cobra CLI)
├── internal/
│   ├── domain/entity/      # Entidades de domínio (pura lógica de negócio)
│   ├── core/
│   │   ├── usecase/       # Casos de uso
│   │   └── port/          # Interfaces dos repositórios
│   └── infrastructure/
│       ├── persistence/    # Implementações SQLite
│       ├── cli/            # Formatadores de saída
│       └── config/         # Loader de configuração
├── migrations/             # Migrations SQL (golang-migrate)
└── pkg/expr/              # Parser de expressões matemáticas
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
