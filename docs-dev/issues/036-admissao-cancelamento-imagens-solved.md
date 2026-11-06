# 036 — Admissão e cancelamento de imagens

Status: **fechada — resolvida em 06/11/2026**. Issue local, sem publicação.
Origem: [parecer Advisor das 032/033](../comparisons/transicao-imagens/ADVISOR_032_033.md).
Prioridade: corrigir antes da [034](034-build-distribuicao-sem-cgo-solved.md).

## Trabalho e aceite

- Limitar e permitir cancelamento da admissão de uploads administrativos antes
  de interpretar corpo, ler arquivos ou aguardar travas; preservar associação e
  coordenação com coleta de órfãos. A capacidade global de imagens decodificadas
  permanece separada do limite das requisições administrativas.
- Separar o contexto dos participantes do trabalho compartilhado em derivações e
  redownload: espera cancelável por cliente, limite de trabalhos distintos e prazo
  próprio; encerrar trabalho quando não houver participantes, sem perder controle
  de codecs que não aceitam interrupção.
- Padronizar erro de arquivo grande e de formato recusado em upload e download.
- Testar concorrência, cancelamento, sobrecarga, erros HTTP, imagens válidas,
  compatibilidade de cache, proteção de associação e ausência de arquivos parciais.
  Registrar resultados e limitações. O risco histórico de coleta em produtores
  automáticos do catálogo fica em demanda separada, conforme o plano acordado.

## Entrega

[Implementação, testes e limitações](../comparisons/transicao-imagens/ISSUE_036.md).
Admissão antes da leitura HTTP, cancelamento independente por participante e
padronização de erros foram implementados; suíte Go, race e Go 1.25 sem CGO
passaram no escopo documentado. Risco histórico segue na issue 037.
