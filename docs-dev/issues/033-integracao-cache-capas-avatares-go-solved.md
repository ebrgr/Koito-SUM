# 033 — Integrar pipeline Go ao cache, capas e avatares

Status: **fechada — resolvida em 06/11/2026**. Issue local, sem publicação.
Depende da [032](032-pipeline-imaging-webp-go-pure-solved.md).
Plano: [transição de imagens](../PLANO_TRANSICAO_IMAGENS_GO.md).

## Trabalho

- Migrar DownloadImageContext, SaveImage, compressImage e a validação de artwork
  para o mesmo componente; conferir todos os caminhos de avatar e download.
- Limitar também chamadas diretas com reader ilimitado, sem depender somente
  dos limites do formulário/HTTP. Manter política por operação da 031.
- Preservar UUIDs, caminhos `.webp`, MIME, tamanhos, temporário/rename,
  singleflight e coordenação com associação/coleta de órfãos.
- Testar cache antigo, derivado ausente, source ausente/redownload, falhas de
  decode/write/associação e requisições simultâneas com o mesmo e outros IDs.
- Verificar capas e avatares por UI e HTTP em cópias fictícias, com comparações
  visuais dos arquivos e amostra desktop/celular, além de integridade do banco.
- Registrar mudança visível no changelog e resultados no `ISSUE_033.md`.

## Aceite

Nenhum caminho operacional de imagem usa bimg; erros não publicam arquivos
parciais ou associam imagem inválida. Contratos HTTP/cache e regras por perfil
continuam verificados. Não limpar/recodificar o acervo nem alterar SQLite v3.
Dependências nativas serão retiradas formalmente na 034.

## Entrega

[Resultados, testes e comparação visual](../comparisons/transicao-imagens/ISSUE_033.md):
cache, capas e avatares integrados ao pipeline Go, com formatos e limites
registrados para o usuário. Testes de concorrência, HTTP, compatibilidade WebP,
rollback e capturas em bases descartáveis passaram. A remoção formal da bimg,
libvips e ajustes de distribuição continuam na 034.
