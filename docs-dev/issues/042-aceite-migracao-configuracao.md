# 042 — Integração e aceite da configuração

**Status:** Aberta · **Depende de:** 038–041

## Objetivo

Validar a experiência completa de configuração e documentar implantação, atualização e reversão.

## Aceite

- Exercitar instalação existente sem TOML e instalação nova; salvar, reiniciar e confirmar valores efetivos e precedência, incluindo `_FILE` e substituições pelo ambiente.
- Validar conta master, negação de acesso, CSRF, segredos, relay por perfil, fusos, importação, porta, erros de arquivo e alterações concorrentes.
- Verificar Compose, Docker, VS Code e execução local; preservar dados existentes e SQLite v3. Executar testes pertinentes e registrar limitações.
- Atualizar documentação do usuário, exemplo seguro, changelog e roadmap; registrar capturas desktop/celular e resultados, sem credenciais. Descrever backup e reversão sem divulgar segredos.
