
# Use Cases do My-Finances

```bash
# Adição de Contas
myfin add account main
myfin add account deh
myfin add account edu

# Configuração de Valores Default
myfin config default.currency BRL
myfin config default.namespace main
myfin config default.credit-card main

# Criação de Categorias
# Cria uma categoria de tipo income
myfin add category --type exp --name "Transporte & Derivados" --alias t_and_d --emoji 🚘️
myfin add category --type inc --name "Salário Líquido BJD" --alias bjd --emoji 🚜

# Configuração de Cartões de Crédito
myfin add credit-card main --closing-day 9 --due-day 16

# Inclusão de despesas
myfin add expense --tags debito --date 26-05-30 --category doacoes_e_emprestimos --description "Doação para o Thales" 45.3

# Se passar --credit-card, --times é obrigatório
# Esse comando "--times" gera N despesas, uma para cada mês da fatura do cartão.
myfin add expense --tags credito,div --date 26-05-30 --credit-card main --times 5
myfin add expense --tags credit --category sa --date 26-05-31 --category reembolso --description "Reembolso do Thales" 45.3/2
myfin add expense --tags credit --credit-card main --category rest --date 26-03-01 --description "Clube Ifood" 12.9

# Inclusão de receitas
# Os comandos podem aceitar expressões matemáticas em vez de valores fechados.
myfin add income --namespace main --date 26-05-31 --category sl --description "Salário John Deere" 45.3*3

# Geração de Relatórios
# Formato Markdown simples elencando todas as despesas
# Flags --from e --until são opcionais, e definem margens inclusivas para listagem das entries.
myfin report --format md --namespace main --from 26-05-01 --until 26-05-10 --filter-tags credito --filter-categories sa,transport

# Mostra o saldo de todas as contas, considerando todas as entradas de receitas e despesas
myfin report --format md balances

# Remove um record por id em um determinado namespace. Obviamente, assume-se o namespace default se a flag não for informada
myfin remove --namespace main record 123

# Observações
# Depois de cada comando, vamos imprimir na tela o(s) registros que foram adicionados, para dar uma confirmação ao usuário. Faz parte imprimir o
# id único gerado para aquela receita. Esse id também deverá aparecer discretamente no relatório, ele é importante caso o usuário queira referenciá-lo para
# exclusão.
```

## Data de Pagamento da Despesa versus Data de Realização da Despesa

Para despesas de cartão de crédito, existe uma diferença entre data de pagamento da despesa das parcelas versus data de realização da despesa, e isso é importante para a exibição do relatório depois e para a contabilização do saldo de uma determinada conta.

Para fins de cálculo de balanço da conta, o importante é a data de pagamento da despesa. Contudo, para fins de relatório, o usuário também precisará saber a data de realização da despesa. Esses valores são determinados pelo due-day e closing-day configurados no cartão. A "--date" passada no comando é sempre a data de realização da despesa, não do pagamento. A data de pagamento é inferida a partir da data da despesa, due-day e closing-day (do cartão).

Exemplos:

* Assumindo que o cartão está configurado com closing day 9 e due day 16:
  * data de realização da despesa é 9, ou 10, ou 15. A data de pagamento da primeira parcela é o dia 16 do mês seguinte à da realização da despesa.
  * data de realização da despesa é 8, ou 1, ou 2. A data de pagamento da primeira parcela é o dia 16 do mesmo mês de realização da despesa.

## Considerações sobre arquitetura

Vamos seguir Clean Architecture à risca, com ports and adapters, e também DDD tático para isolamento completo da lógica de negócio na camada de domínio.
