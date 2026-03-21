# myfin - CLI de Gestão de Finanças Pessoais

CLI para gerenciamento de finanças pessoais com suporte a múltiplos bancos de dados SQLite, cartões de crédito com rastreamento de faturas, categorias, tags e parsing de expressões matemáticas para valores.

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

Cada banco de dados SQLite contém suas próprias contas, categorias, cartões e lançamentos. Você pode configurar padrões no arquivo `~/.myfin.yaml`:

```yaml
default.db_path: ./data/main.db
default.currency: BRL
default.credit-card: nubank
```

O arquivo é criado automaticamente na primeira configuração.

## Uso Rápido

### Criar uma conta bancária

```bash
myfin add account checking
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

### Bancos de dados

Cada banco de dados SQLite é independente. Use `-d` para especificar qual database usar:

```bash
myfin -d ./data/work.db add expense 100.00 --category comida
```

Se não especificado, usa o `default.db_path` da configuração.

## Referência de Comandos

Use `myfin add --help` e `myfin add <comando> --help` para ver todas as flags disponíveis.

| Comando | Descrição |
|---------|-----------|
| `myfin add account <nome>` | Cria uma conta bancária |
| `myfin add category` | Cria uma categoria |
| `myfin add credit-card <nome>` | Cria um cartão de crédito |
| `myfin add expense [valor]` | Adiciona uma despesa |
| `myfin add income [valor]` | Adiciona uma receita |

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
