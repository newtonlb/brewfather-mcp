# Brewfather MCP Server

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server that exposes the full [Brewfather API v2](https://docs.brewfather.app/api) to AI agents. Built in Go, it allows any MCP-compatible client (such as [Cursor](https://www.cursor.com/)) to read and manage your brewing batches, recipes, and ingredient inventory through natural language.

---

**[Português](#servidor-mcp-para-brewfather)** | **[English](#overview)**

---

## Overview

[Brewfather](https://brewfather.app/) is a popular tool for designing and tracking homebrewed beer recipes. This MCP server wraps the entire Brewfather API v2 and presents it as 20 tools that an AI agent can invoke. All API responses are transformed into human/LLM-readable structured text rather than raw JSON, making it significantly easier for the agent to understand and reason about your brewing data.

## Features

### Batch Tools (6)

| Tool | Description |
|---|---|
| `list_batches` | List batches with filtering by status, pagination, and sorting |
| `get_batch` | Get full batch details — measured values, recipe, ingredients, notes |
| `update_batch` | Update batch status and/or measured values (gravity, volume, pH) |
| `get_batch_last_reading` | Get the most recent sensor/manual reading (gravity, temperature) |
| `get_batch_readings` | Get all readings for fermentation trend analysis |
| `get_batch_brew_tracker` | Get brew day tracker state — stages, steps, timers |

### Recipe Tools (2)

| Tool | Description |
|---|---|
| `list_recipes` | List recipes with author, style, equipment, and type info |
| `get_recipe` | Get full recipe — style guidelines, ingredients, mash/fermentation profiles, equipment |

### Inventory Tools (12)

Three tools for each of the four ingredient categories (**fermentables**, **hops**, **miscs**, **yeasts**):

| Tool Pattern | Description |
|---|---|
| `list_{category}` | List inventory items with stock levels and key properties |
| `get_{category}` | Get full ingredient details — properties, stock, cost, notes |
| `update_{category}_inventory` | Set or adjust inventory amounts |

## Requirements

- **Go 1.24+** (the project uses `go 1.26` in `go.mod`)
- **Brewfather Premium** account (API access requires a premium subscription)
- **Brewfather API Key** — generated from Brewfather Settings > API
- An MCP-compatible client (e.g., Cursor, Claude Desktop)

## Getting the API Key

1. Open [Brewfather](https://brewfather.app/) and sign in
2. Go to **Settings** (gear icon)
3. Scroll to the **API** section
4. Click **Generate API Key**
5. Copy both the **User ID** and the **API Key** — you'll need both

> **Note:** The Brewfather API has a rate limit of 500 requests per hour.

## Installation

### Clone the repository

```bash
git clone git@github.com:your-username/brewfather-mcp.git
cd brewfather-mcp
```

### Download dependencies

```bash
go mod download
```

### Build the binary

```bash
go build -o brewfather-mcp .
```

### Run the tests

```bash
go test ./... -v
```

## Configuration

### Cursor

Add the following entry to your MCP configuration file at `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "brewfather": {
      "command": "/absolute/path/to/brewfather-mcp",
      "env": {
        "BREWFATHER_USER_ID": "your-user-id",
        "BREWFATHER_API_KEY": "your-api-key"
      }
    }
  }
}
```

Replace `/absolute/path/to/brewfather-mcp` with the actual path to your compiled binary, and fill in your Brewfather credentials.

After saving, restart Cursor or reload MCP servers. The 20 Brewfather tools should appear and be available to the agent.

### Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "brewfather": {
      "command": "/absolute/path/to/brewfather-mcp",
      "env": {
        "BREWFATHER_USER_ID": "your-user-id",
        "BREWFATHER_API_KEY": "your-api-key"
      }
    }
  }
}
```

### Running Standalone (for testing)

You can also run the server directly for testing purposes:

```bash
export BREWFATHER_USER_ID="your-user-id"
export BREWFATHER_API_KEY="your-api-key"
./brewfather-mcp
```

The server communicates over **stdio** (stdin/stdout), so it won't produce visible output — it waits for MCP protocol messages.

## Architecture

```
brewfather-mcp/
├── main.go                          # Entry point — wires everything together
├── internal/
│   ├── client/
│   │   └── brewfather.go            # HTTP client for Brewfather API v2
│   ├── service/
│   │   ├── format.go                # Shared formatting utilities
│   │   ├── batch.go                 # Batch business logic + text formatting
│   │   ├── recipe.go                # Recipe business logic + text formatting
│   │   └── inventory.go             # Inventory business logic + text formatting
│   └── handler/
│       ├── result.go                # MCP result helpers
│       ├── batch.go                 # Batch tool definitions + handlers
│       ├── recipe.go                # Recipe tool definitions + handlers
│       └── inventory.go             # Inventory tool definitions + handlers
├── go.mod
├── go.sum
└── README.md
```

**Three-layer separation:**

- **Client** — handles HTTP requests, authentication (Basic Auth), error handling, and returns raw JSON
- **Service** — unmarshals JSON into maps, transforms data into LLM-friendly structured text with proper units, dates, and formatting
- **Handler** — defines MCP tool schemas (with `jsonschema` descriptions), maps inputs to service calls, and wraps results for the MCP protocol

## License

See [LICENSE](LICENSE) for details.

---

# Servidor MCP para Brewfather

Um servidor [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) que expõe toda a [API v2 do Brewfather](https://docs.brewfather.app/api) para agentes de IA. Desenvolvido em Go, ele permite que qualquer cliente compatível com MCP (como o [Cursor](https://www.cursor.com/)) leia e gerencie suas brassagens, receitas e estoque de ingredientes através de linguagem natural.

## Visão Geral

O [Brewfather](https://brewfather.app/) é uma ferramenta popular para criar e acompanhar receitas de cerveja artesanal. Este servidor MCP encapsula toda a API v2 do Brewfather e a apresenta como 20 ferramentas que um agente de IA pode invocar. Todas as respostas da API são transformadas em texto estruturado legível para humanos e LLMs, em vez de JSON bruto, facilitando significativamente a compreensão e o raciocínio do agente sobre seus dados cervejeiros.

## Funcionalidades

### Ferramentas de Brassagem (6)

| Ferramenta | Descrição |
|---|---|
| `list_batches` | Listar brassagens com filtro por status, paginação e ordenação |
| `get_batch` | Obter detalhes completos — valores medidos, receita, ingredientes, notas |
| `update_batch` | Atualizar status e/ou valores medidos (gravidade, volume, pH) |
| `get_batch_last_reading` | Obter a leitura mais recente do sensor/manual (gravidade, temperatura) |
| `get_batch_readings` | Obter todas as leituras para análise de tendência de fermentação |
| `get_batch_brew_tracker` | Obter estado do acompanhamento do dia de brassagem — etapas, passos, cronômetros |

### Ferramentas de Receita (2)

| Ferramenta | Descrição |
|---|---|
| `list_recipes` | Listar receitas com autor, estilo, equipamento e tipo |
| `get_recipe` | Obter receita completa — diretrizes de estilo, ingredientes, perfis de mostura/fermentação, equipamento |

### Ferramentas de Estoque (12)

Três ferramentas para cada uma das quatro categorias de ingredientes (**fermentáveis**, **lúpulos**, **diversos**, **leveduras**):

| Padrão da Ferramenta | Descrição |
|---|---|
| `list_{categoria}` | Listar itens do estoque com níveis e propriedades principais |
| `get_{categoria}` | Obter detalhes completos do ingrediente — propriedades, estoque, custo, notas |
| `update_{categoria}_inventory` | Definir ou ajustar quantidades no estoque |

## Requisitos

- **Go 1.24+** (o projeto usa `go 1.26` no `go.mod`)
- **Conta Premium do Brewfather** (o acesso à API requer assinatura premium)
- **Chave de API do Brewfather** — gerada em Configurações > API no Brewfather
- Um cliente compatível com MCP (ex.: Cursor, Claude Desktop)

## Obtendo a Chave de API

1. Abra o [Brewfather](https://brewfather.app/) e faça login
2. Vá em **Configurações** (ícone de engrenagem)
3. Role até a seção **API**
4. Clique em **Generate API Key**
5. Copie tanto o **User ID** quanto a **API Key** — você precisará de ambos

> **Nota:** A API do Brewfather tem um limite de 500 requisições por hora.

## Instalação

### Clonar o repositório

```bash
git clone git@github.com:seu-usuario/brewfather-mcp.git
cd brewfather-mcp
```

### Baixar dependências

```bash
go mod download
```

### Compilar o binário

```bash
go build -o brewfather-mcp .
```

### Executar os testes

```bash
go test ./... -v
```

## Configuração

### Cursor

Adicione a seguinte entrada ao seu arquivo de configuração MCP em `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "brewfather": {
      "command": "/caminho/absoluto/para/brewfather-mcp",
      "env": {
        "BREWFATHER_USER_ID": "seu-user-id",
        "BREWFATHER_API_KEY": "sua-api-key"
      }
    }
  }
}
```

Substitua `/caminho/absoluto/para/brewfather-mcp` pelo caminho real do binário compilado e preencha suas credenciais do Brewfather.

Após salvar, reinicie o Cursor ou recarregue os servidores MCP. As 20 ferramentas do Brewfather devem aparecer e ficar disponíveis para o agente.

### Claude Desktop

Adicione ao config do Claude Desktop (`~/Library/Application Support/Claude/claude_desktop_config.json` no macOS):

```json
{
  "mcpServers": {
    "brewfather": {
      "command": "/caminho/absoluto/para/brewfather-mcp",
      "env": {
        "BREWFATHER_USER_ID": "seu-user-id",
        "BREWFATHER_API_KEY": "sua-api-key"
      }
    }
  }
}
```

### Executando Diretamente (para testes)

Você também pode executar o servidor diretamente para fins de teste:

```bash
export BREWFATHER_USER_ID="seu-user-id"
export BREWFATHER_API_KEY="sua-api-key"
./brewfather-mcp
```

O servidor se comunica via **stdio** (stdin/stdout), então não produzirá saída visível — ele aguarda mensagens do protocolo MCP.

## Arquitetura

```
brewfather-mcp/
├── main.go                          # Ponto de entrada — conecta tudo
├── internal/
│   ├── client/
│   │   └── brewfather.go            # Cliente HTTP para a API v2 do Brewfather
│   ├── service/
│   │   ├── format.go                # Utilitários de formatação compartilhados
│   │   ├── batch.go                 # Lógica de brassagem + formatação de texto
│   │   ├── recipe.go                # Lógica de receita + formatação de texto
│   │   └── inventory.go             # Lógica de estoque + formatação de texto
│   └── handler/
│       ├── result.go                # Helpers de resultado MCP
│       ├── batch.go                 # Definições + handlers de ferramentas de brassagem
│       ├── recipe.go                # Definições + handlers de ferramentas de receita
│       └── inventory.go             # Definições + handlers de ferramentas de estoque
├── go.mod
├── go.sum
└── README.md
```

**Separação em três camadas:**

- **Client** — lida com requisições HTTP, autenticação (Basic Auth), tratamento de erros e retorna JSON bruto
- **Service** — deserializa JSON em mapas, transforma dados em texto estruturado otimizado para LLMs com unidades, datas e formatação adequadas
- **Handler** — define schemas das ferramentas MCP (com descrições `jsonschema`), mapeia inputs para chamadas de serviço e encapsula resultados para o protocolo MCP

## Licença

Veja [LICENSE](LICENSE) para detalhes.
