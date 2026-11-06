# 041 — Edição de configuração pelo master

**Status:** Aberta · **Depende de:** 038–040

## Objetivo

Criar na interface Go a administração das opções definidas no contrato, incluindo os fluxos específicos para itens que não pertencem ao TOML.

## Aceite

- Restringir leitura e alteração no servidor ao administrador global, com as proteções de sessão, origem e CSRF do projeto; aplicar a política de confirmação decidida na 038.
- Usar endpoints administrativos separados do endpoint público de configuração, respostas `no-store`, limite de corpo e campos permitidos; negar autenticação por chave de ingestão.
- Para Subsonic, prever teste remoto explícito com prazo, mensagens sem credenciais e recuperação offline caso a configuração impeça a inicialização.
- Mostrar descrição, origem, valor salvo quando não secreto, substituição por ambiente, erro de validação e reinício pendente; nunca exibir segredos.
- Oferecer os fluxos aprovados para conta master, bootstrap, diretório e relay por perfil, com efeitos e limitações explícitos; não prometer alterar o Compose pelo aplicativo.
- Usar templates Go e JavaScript apenas essencial; cobrir acessibilidade, navegação e testes de permissões, segredos e falhas. Registrar capturas de desktop e celular na etapa de aceite.
