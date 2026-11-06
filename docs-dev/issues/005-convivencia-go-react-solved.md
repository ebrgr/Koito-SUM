# 005 — Preservar navegação e ações na convivência Go/React

Status: **fechada e resolvida em 04/11/2026**, após os testes da 004.
Issue local, sem publicação.
Plano: [roadmap 2](../ROADMAP_02_PAGINAS.md).

## Trabalho

- Registrar as páginas HTML específicas sem interceptar API, imagens e assets.
- Manter mídia, retrospectiva e administração na SPA durante a migração.
- Criar ponte de navegação completa para páginas Go quando a origem for React,
  inclusive histórico do navegador e navegação programática, sem loops.
- Preservar seleção por URL, links antigos e `next` restrito a rotas internas.
- Definir ativação opt-in, inicialmente com a implantação atual em React.
- Manter Aparência, acesso administrativo, busca e ações existentes de histórico
  funcionais; fechar os caminhos transitórios antes de ativar o modo Go.
- Preservar sessão, proteção de origem e a distinção entre admin e chave de player.

## Aceite

Jornadas Go → React → Go, duas abas/perfis, voltar/avançar, login/admin e leitura
pública validadas com dados descartáveis. Remover a SPA e publicar alterações
continuam fora do escopo desta issue.

## Encerramento

Modo opt-in `KOITO_WEB_UI=go`, React padrão, dispatcher compartilhado/testado e
ponte da SPA implementados. Mídia, rewind, admin, API e imagens preservados.
Seleção/next interno, busca, aparência e exclusão com sessão estão integrados.
Navegador confirmou React → Go, voltar/avançar, perfis em abas independentes e
login administrativo. HTTP confirmou recusa de chave de player/origem externa
e aceitação da sessão, com restauração do registro de teste.

Validação e capturas: [relatório do roadmap 2](../comparisons/roadmap-02/README.md).
`make check ui.check`, build React/Go, testes de corrida focados e validações
HTTP/multiusuário passaram. Sem commit ou publicação.
