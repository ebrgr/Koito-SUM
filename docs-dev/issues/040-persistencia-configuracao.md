# 040 — Persistência segura e concorrência da configuração

**Status:** Aberta · **Depende de:** 038 e 039

## Objetivo

Disponibilizar serviço de validação e escrita do TOML para a administração, sem expor segredos ou perder mudanças concorrentes.

## Aceite

- Aplicar contrato de segredos aprovado: manter/substituir/limpar explicitamente; nunca retornar valores secretos ou aceitar caminhos arbitrários `_FILE` da interface.
- Validar o conjunto efetivo e informar valores sobrescritos pelo ambiente; distinguir configuração salva da ativa e indicar necessidade de reinício.
- Proteger contra concorrência entre requisições e edição externa com revisão e conflito explícito; não descartar silenciosamente comentários ou campos não reconhecidos.
- Escrever de forma segura no mesmo diretório, com permissões restritas, tratamento de links, erros e substituição apropriada à plataforma; não corromper o arquivo anterior em falha.
- Testar concorrência, interrupção, falha de I/O, permissões, arquivos existentes e ausência de vazamento em logs e erros.
