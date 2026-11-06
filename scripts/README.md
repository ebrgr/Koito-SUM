# Scripts de validação

Os validadores atuais usam bancos SQLite descartáveis: `validate_multiuser.py`,
`validate_admin_webui.py`, `validate_permissions_017.py`,
`validate_isolation_019.py`, `validate_migration_030.py` e
`validate_standalone_030.py`. Aceitam `--binary` e nunca usam o banco operacional.
`make multiuser.check` constrói o binário e executa a jornada principal.

Os scripts de comparação visual das issues 022–028 que iniciam React permanecem
como registro histórico dos métodos usados antes da remoção. Eles exigem uma
versão histórica do cliente e não são a suíte executável da versão atual; suas
capturas e relatórios continuam em `docs-dev/comparisons/`.
