# 001 — Prévia visual da interface em Go (roadmap 1)

**Status: fechada e resolvida em 4 de novembro de 2026.**

Primeira prévia implementada e validada localmente. A revisão ampliada de paridade
continua nas etapas seguintes do roadmap e não mantém esta issue aberta.
Registro local autorizado pelo usuário; não publicar issue, código ou imagens
no GitHub nesta etapa.

## Objetivo

Reproduzir layout, fontes, 12 temas, barra lateral, avatares e famílias de cards
com templates Go, CSS e JavaScript mínimo. Preservar a SPA e o SQLite de esquema
consolidado v3. Plano: [roadmap 1](../ROADMAP_01_INTERFACE.md).

## Escopo

- Prévia isolada em loopback, sem banco, credenciais ou serviços musicais reais.
- Dados fictícios compartilhados com uma referência da SPA para comparação.
- Capturas desktop/celular, registro de diferenças e comandos reproduzíveis.
- Preferência de aparência, tema personalizado e diálogos acessíveis.
- Nenhuma migração de rotas públicas, consultas ou administração real.

## Aceite

- Executar a prévia Go sem Node em runtime e sem alterar a aplicação principal.
- Preservar as medidas e paletas descritas na análise; revisar capturas reais.
- Verificar navegação local, teclado, responsividade e fallback de aparência.
- Registrar os resultados e limitações em `docs-dev/`, com imagens locais.

## Decisões

- Issues ficam em Markdown por orientação explícita do usuário, substituindo
  nesta etapa a exigência de registro no GitHub de `agents.md`.
- A numeração 001 é local; não corresponde a uma issue publicada no GitHub.

## Entrega e validação

- Implementados `cmd/ui-preview` e `internal/webui`, com fixtures, temas, diálogos,
  avatares e componentes de apresentação. Execução por `make ui.preview`.
- Testes da prévia/JavaScript, detector de corridas, build sem CGO, suíte Go completa
  e build React passaram. Nenhuma rota ou migração da aplicação foi substituída.
- [Relatório e screenshots](../comparisons/roadmap-01/README.md) registram as
  comparações e os limites, incluindo a matriz de breakpoints ainda incompleta.
- Próxima etapa: revisar diferenças visuais restantes e integrar as páginas reais
  conforme o passo 2, preservando perfil explícito e permissões.
