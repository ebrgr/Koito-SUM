# 034 — Distribuir e validar a aplicação sem CGO/libvips

Status: **fechada — resolvida em 06/11/2026**. Issue local, sem publicação.
Depende da [033](033-integracao-cache-capas-avatares-go-solved.md) e da
[036](036-admissao-cancelamento-imagens-solved.md).
Encerra a fase 1 do [plano](../PLANO_TRANSICAO_IMAGENS_GO.md).

## Trabalho

- Confirmar zero consumidores de bimg e retirar a dependência com revisão de
  go.mod/go.sum. Manter x/image quando usado pelas fontes ou imaging.
- Atualizar build, execução, VS Code, Docker, CI e documentação para o caminho
  normal sem CGO/libvips; remover dependências nativas dispensáveis.
- Preservar um job -race com CGO/compilador C quando necessário, sem libvips.
- Compilar no Go declarado, validar Linux amd64/arm64 conforme distribuição
  atual e testar o ambiente local, certificados, fusos, sinais e permissões.
- Repetir testes Go, administração/ingestão/migração e validação do binário em
  ambiente sem libvips; testar contêiner descartável e healthcheck.
- Consolidar benchmarks e comparações 031–033, leitura de cache legado, saída
  nova pelo codec anterior e procedimento de upgrade/rollback em cópias.
- Registrar conclusão, limites, revisão Advisor e changelog; não publicar imagem
  ou push sem autorização específica.

## Aceite

Binário e imagem funcionais sem libvips, build normal com CGO_ENABLED=0 e testes
da matriz concluídos com alcance explícito. Nenhum downgrade de esquema, limpeza
ou recodificação automática de dados. Relatório `ISSUE_034.md` permite avaliar
qualidade e custos reais; fase 2 segue na 035 após essa conclusão técnica.

## Entrega

[Resultados, Advisor e reprodução](../comparisons/transicao-imagens/ISSUE_034.md):
dependência bimg removida, build/debug/CI/Docker atualizados, binários Go 1.25
sem CGO e contêineres arm64 e amd64 validados. Testes Go, migração, UI, race,
TLS, fuso, persistência, healthcheck, sinais e rollback documentados. Nenhuma
imagem foi publicada; 035 e 037 estavam abertas naquela conclusão.
