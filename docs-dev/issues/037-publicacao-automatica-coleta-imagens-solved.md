# 037 — Coordenar publicação automática e coleta de imagens

Status: **fechada — resolvida em 06/11/2026**. Issue local, sem publicação.
Origem: [revisão Advisor](../comparisons/transicao-imagens/ADVISOR_032_033.md).

## Problema

Produtores automáticos de artista/álbum podem publicar uma source antes da
associação no catálogo, sem participar da trava usada pela coleta de órfãos.
Uma coleta concorrente pode apagar o arquivo nessa janela. Esse risco antecede
a migração do codec; o redownload posterior depende da fonte remota.

## Trabalho e aceite

- Proteger publicação e associação nos caminhos automáticos sem segurar uma
  trava global durante downloads ou criar inversão de ordem de locks.
- Preservar cancelamento, `SkipCacheImage`, recuperação remota e tratamento de
  erro de banco; não remover arquivos possivelmente associados.
- Reproduzir a corrida em testes coordenados com dados fictícios, incluindo
  falha/cancelamento e concorrência entre produtores e coletor.
- Registrar o resultado em `docs-dev/comparisons/transicao-imagens/` e changelog
  se houver mudança perceptível. Pode ser tratada após a 034.

## Entrega

[Implementação e testes](../comparisons/transicao-imagens/ISSUE_037.md): reservas
por UUID protegem as cinco publicações automáticas sem manter a trava durante
rede ou banco. Testes coordenados, SQLite, cancelamento, falhas e redownload
passaram; a coordenação é limitada ao processo atual.
