# 045 — Links e diretório do TuneTrail

**Status:** Aberta

## Objetivo

Alinhar links do repositório e referência futura da imagem ao nome TuneTrail, preservando créditos e histórico, e renomear a pasta local sem perder alterações.

## Aceite

- Atualizar links próprios e exemplos de imagem; manter referências ao Koito upstream e alertar que a imagem nova depende de publicação.
- Confirmar que o VS Code usa caminhos relativos ao workspace e preservar dados, histórico Git e alterações locais.
- Validar links locais, configuração e testes pertinentes; registrar eventuais limitações para mover o diretório ou publicar o site.

## Resultado

Links ativos do projeto, templates, metadados OCI e destino da imagem no Compose foram alinhados a `ebrgr/TuneTrail` e `ghcr.io/ebrgr/tunetrail`. O domínio do site de documentação (`koito.io`) permanece até ser definido e configurado um endereço do TuneTrail; referências históricas de imagens e ao upstream foram preservadas. O Advisor confirmou o risco de divergência de GHCR e recomendou preservar o nome operacional Compose, banco, configuração e upstream.

A pasta foi movida integralmente para `/Users/evertonberger/Projects/TuneTrail`. O VS Code usa `${workspaceFolder}` e precisa reabrir essa pasta. `make version.check`, testes de versionamento e testes dos templates passaram; os remotos foram conferidos após a mudança. A consulta remota do GitHub não retornou conteúdo no ambiente; a existência da imagem nova não foi verificada nem houve publicação.
