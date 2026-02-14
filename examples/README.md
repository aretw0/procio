# procio Examples

Este diretório contém exemplos práticos demonstrando os diferentes casos de uso do `procio`.

## Exemplos Disponíveis

### 📁 [basic/](basic/)

**Introdução aos primitivos fundamentais**

Demonstra o uso básico de `proc` para gerenciamento de processos e `scan` para leitura de entrada.

```bash
go run examples/basic/main.go
```

**Conceitos demonstrados:**

- Iniciar processos com garantia de limpeza (`proc.Start`)
- Scanner básico para leitura de linhas interativas
- Context-based timeout para processos

---

### 📁 [interruptible/](interruptible/)

**Scanner com suporte a Ctrl+C**

Mostra como usar `scan.WithInterruptible()` para habilitar cancelamento via context e interrupção de teclado.

```bash
go run examples/interruptible/main.go
```

**Conceitos demonstrados:**

- `scan.WithInterruptible()` para TTY cancellation
- `signal.NotifyContext` para capturar Ctrl+C
- Terminal upgrade automático (CONIN$ no Windows)
- Graceful shutdown com context

---

### 📁 [observer/](observer/)

**Telemetria customizada com Observer**

Implementa um observer customizado para capturar eventos de processo, I/O e scanning.

```bash
go run examples/observer/main.go
```

**Conceitos demonstrados:**

- Implementação da interface `procio.Observer`
- `procio.SetObserver()` para injeção de observador
- Telemetria de processo (`OnProcessStarted`, `OnProcessFailed`)
- Telemetria de I/O (`OnIOError`, `OnScanError`)
- Zero-dependency logging via observer pattern

---

### 📁 [composition/](composition/)

**Composição: Process + Scanner**

Demonstra a composição de múltiplos primitivos em uma aplicação CLI interativa com processo em background.

```bash
go run examples/composition/main.go
```

**Conceitos demonstrados:**

- Composição de `proc` + `scan` em aplicação real
- Gerenciamento de processo em background
- Scanner interativo com comandos customizados
- Context cancellation propagado para todos os componentes

---

## Executando Todos os Exemplos

Para compilar e verificar todos os exemplos:

```bash
# Compilar todos
go build ./examples/basic
go build ./examples/interruptible
go build ./examples/observer
go build ./examples/composition

# Ou executar diretamente
for dir in basic interruptible observer composition; do
  echo "Running examples/$dir..."
  go run examples/$dir/main.go
done
```

## Próximos Exemplos (Roadmap)

Exemplos planejados para versões futuras:

- **pty/** - Pseudo-terminal allocation (v0.2.0)
- **streaming/** - Streaming telemetry para métricas em tempo real (v0.2.0)
- **multiprocess/** - Coordenação de múltiplos processos filhos
- **readline/** - Integração com readline libraries

## Contribuindo com Exemplos

Exemplos devem:

1. Ser auto-contidos (não depender de arquivos externos)
2. Funcionar em Windows e Linux
3. Demonstrar um conceito específico claramente
4. Incluir comentários explicativos
5. Ter graceful shutdown (via context)

Para adicionar um novo exemplo, crie um diretório `examples/<nome>/` com um `main.go` e atualize este README.
