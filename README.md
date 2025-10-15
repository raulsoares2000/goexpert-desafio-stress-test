# Go Stress Test CLI

Esta é uma ferramenta de linha de comando (CLI) desenvolvida em Go para realizar testes de carga (stress test) em serviços web. O usuário pode especificar a URL do serviço, o número total de requisições e a quantidade de chamadas simultâneas para avaliar o comportamento da aplicação sob pressão.

## Funcionalidades

* **CLI Robusta:** Construído com a popular biblioteca [Cobra](https://github.com/spf13/cobra) para uma interface de linha de comando amigável.
* **Teste de Carga HTTP:** Realiza requisições HTTP para a URL especificada.
* **Parâmetros Configuráveis:** Permite ao usuário definir a URL do alvo, o número total de requisições e o nível de concorrência através de flags.
    * `--url / -u`: URL do serviço a ser testado.
    * `--requests / -r`: Número total de requests a serem realizados.
    * `--concurrency / -c`: Número de chamadas simultâneas (workers).
* **Relatório Detalhado:** Ao final da execução, gera um relatório com as seguintes métricas:
    * Tempo total gasto na execução.
    * Quantidade total de requests realizados.
    * Quantidade de requests com status HTTP 200 (sucesso).
    * Distribuição de outros códigos de status HTTP (4xx, 5xx, etc.).
    * Contagem de erros de conexão.

## Como o Código Funciona

A aplicação é estruturada em torno do pacote `cmd`, que utiliza a biblioteca Cobra para gerenciar os comandos da CLI. O comando principal é o `charge`, que aciona a lógica do teste de carga.

A função `loadTest` é o coração do programa:
1.  **Inicialização:** Define contadores atômicos (para segurança em ambiente concorrente), um mapa para os códigos de status e canais para gerenciar o fluxo de trabalho.
2.  **Concorrência com Goroutines:** Um número de Goroutines (workers), definido pela flag `--concurrency`, é iniciado. Cada goroutine funciona como um usuário simulado.
3.  **Distribuição de Trabalho:** Um canal (`requestCounter`) é usado como um "buffer de trabalho". A função principal envia um número de sinais igual ao total de `requests`. Cada worker pega um sinal desse canal, realiza uma requisição HTTP e continua até que os sinais acabem.
4.  **Sincronização:** Um `sync.WaitGroup` é utilizado para garantir que a função principal espere que todas as goroutines terminem seu trabalho antes de prosseguir para a geração do relatório.
5.  **Coleta de Resultados:** Os resultados, como tempos de resposta e códigos de status, são coletados de forma segura usando canais e um `sync.Mutex` para proteger o acesso ao mapa de status, evitando "race conditions".

## Como o Docker Funciona

O projeto inclui um `Dockerfile` otimizado que utiliza uma construção **multi-stage** para criar uma imagem final leve e segura.

1.  **Estágio 1: `builder`**
    * Este estágio utiliza a imagem oficial `golang:latest` como base para compilar a aplicação.
    * O código-fonte é copiado para dentro da imagem e o comando `go build` é executado, gerando um binário estático e otimizado para Linux.

2.  **Estágio 2: Final**
    * A imagem final é baseada no `scratch`, que é uma imagem Docker vazia, a menor possível.
    * Apenas o binário compilado no estágio anterior é copiado para esta nova imagem.
    * O `ENTRYPOINT` é definido para executar este binário quando um container for iniciado.

Essa abordagem resulta em uma imagem Docker final extremamente pequena (apenas alguns megabytes) e segura, pois não contém nenhum sistema operacional, shell ou bibliotecas desnecessárias.

## Como Executar

Existem duas maneiras de executar o programa: localmente na sua máquina ou via Docker.

### Executando Localmente

**Pré-requisitos:**
* Go (versão 1.18 ou superior) instalado.

1.  **Compile o executável:**
    ```bash
    go build -o stresstest .
    ```

2.  **Execute o teste:**
    ```bash
    ./stresstest charge --url=<URL_ALVO> --requests=<NUMERO> --concurrency=<NUMERO>
    ```
    **Exemplo:**
    ```bash
    ./stresstest charge --url=https://www.google.com --requests=1000 --concurrency=10
    ```

### Executando com Docker

**Pré-requisitos:**
* Docker instalado e em execução.

1.  **Construa a imagem Docker:**
    Na raiz do projeto, execute o comando:
    ```bash
    docker build -t stresstest-app .
    ```

2.  **Execute um container a partir da imagem:**
    ```bash
    docker run --rm stresstest-app charge --url=<URL_ALVO> --requests=<NUMERO> --concurrency=<NUMERO>
    ```
    **Exemplo:**
    ```bash
    docker run --rm stresstest-app charge --url=https://google.com --requests=1000 --concurrency=10
    ```
    *(O flag `--rm` é opcional, mas recomendado, pois remove o container automaticamente após a execução.)*

## Exemplo de Saída

```
Iniciando teste de carga...

--- Relatório do Teste de Carga ---
Tempo total de execução: 8.543219875s
URL Alvo: https://www.google.com
Total de requests planejados: 1000
Total de requests concluídos: 1000
Nível de concorrência: 10 workers
Tempo médio de resposta: 81.234567ms
------------------------------------
Requests com status HTTP 200 (Sucesso): 1000
Erros de conexão (ex: timeout): 0
Distribuição de outros códigos de status HTTP:
------------------------------------
```
