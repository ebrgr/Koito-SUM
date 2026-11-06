# 039 — Leitura e validação de conf.toml

**Status:** Aberta · **Depende de:** 038

## Objetivo

Acrescentar leitura tipada de `conf.toml` no diretório de dados, mantendo instalações atuais e os contratos de ambiente.

## Aceite

- Descobrir o diretório antes da leitura, sem ciclo ou mudança de banco; ausência do arquivo mantém comportamento anterior.
- Aplicar precedência e erros de `_FILE` definidos na 038; preservar distinção entre ausência, falso e vazio. Não criar arquivo automaticamente nem copiar segredos.
- Validar schema e configuração composta com erros úteis e sem dados secretos. Tratar arquivo malformado, permissões, opções desconhecidas e compatibilidade de versões conforme contrato.
- Usar a mesma descoberta e precedência na execução normal e no comando `reset-password`; documentar recuperação offline.
- Testar defaults e todas as fontes, combinações e falhas; garantir compatibilidade com inicialização, scripts, Compose e SQLite v3. Documentar eventuais mudanças de interpretação e executar a suíte relevante.
