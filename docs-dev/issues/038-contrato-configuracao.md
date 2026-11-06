# 038 — Contrato e decisões da configuração

**Status:** Aberta · **Depende de:** estudo do item 2 · **Bloqueia:** 039–042

## Objetivo

Formalizar o contrato de migração descrito no [estudo](../ROADMAP_02_CONFIGURACAO_TOML.md), reconciliando a edição pelo master com diretório de dados, credenciais iniciais, relay por perfil, fontes externas e opções obsoletas.

## Trabalho e aceite

- Confirmar inventário e comportamentos reais, inclusive `_FILE`, valores vazios, defaults, validação e itens de Compose; distinguir segredos e configurações não funcionais.
- Decidir como atender edição de bootstrap e realocação do banco, política de segredos em arquivo, autenticação adicional, preservação de comentários e campos desconhecidos.
- Definir schema versionado, mensagens de erro redigidas, validação e compatibilidade, prioridade ambiente > arquivo secreto > TOML > padrão e efeito após reinício.
- Registrar critérios de migração e de reversão que preservem o SQLite v3 e não exijam exportar senhas para o TOML. Criar matriz de testes por opção e fonte.
- Registrar decisões e pendências no estudo. Nenhum código funcional sem issue correspondente.
