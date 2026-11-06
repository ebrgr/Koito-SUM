# Estudo — configuração em TOML e administração pelo usuário master

Data: 06/11/2026. Este é o item 2 do roadmap atual; não se confunde com a antiga etapa 2 de migração de páginas. Trata-se de planejamento: o aplicativo ainda não lê `conf.toml`. O banco continua sendo SQLite, esquema consolidado v3; a proposta não exige migração do banco.

## Situação atual e fontes

`cmd/api/main.go` fornece a leitura de variáveis a `engine.New` e `internal/cfg/cfg.go`. Para cada opção efetivamente consultada, primeiro é usado `KOITO_NOME` quando não vazio; caso contrário, `KOITO_NOME_FILE` aponta para um arquivo cujo conteúdo é lido e tem espaços nas extremidades removidos. Um arquivo indicado que não pode ser lido encerra o processo; o valor direto prevalece quando ambos estão preenchidos. A aplicação não carrega `.env` por conta própria: ferramentas de desenvolvimento podem fornecê-lo ao processo. Não foram lidos arquivos `.env` reais nem segredos.

`KOITO_CONFIG_DIR` determina o diretório de `koito.db` (`internal/db/sqlite/sqlite.go`); seu padrão é `/etc/koito`. A configuração é carregada no início (`engine/engine.go`, `internal/cfg/cfg.go`), e clientes, servidor e templates também são inicializados com esses valores. Há um setter específico para login gate usado em testes: isso não constitui recarga geral de configuração.

## Inventário das variáveis do aplicativo

Padrões abaixo são os do código, não necessariamente os valores fornecidos pelo Compose ou `.env.example`. As variáveis marcadas como segredo não devem ter seus valores expostos em documentos, logs ou respostas da interface. `_FILE` é um mecanismo adicional para cada opção efetivamente lida; no relay os campos só são lidos quando o relay legado está habilitado.

| Variável | Padrão e finalidade atual | Tratamento proposto |
| --- | --- | --- |
| `KOITO_CONFIG_DIR` | `/etc/koito`; diretório do banco, importação e cache | Bootstrap externo; descobrir antes de ler TOML. Mudá-lo também muda o banco usado: não mover dados automaticamente. |
| `KOITO_WEB_UI` | `go`; outro valor causa erro | Compatibilidade; não oferecer React. |
| `KOITO_BIND_ADDR` | vazio; endereço de escuta | Rede; reinício. Vazio permite bind em interfaces não especificadas. |
| `KOITO_LISTEN_PORT` | `4110` se ausente/inválida | Rede; reinício; validar faixa na nova interface. |
| `KOITO_ENABLE_STRUCTURED_LOGGING` | `false`; logging estruturado | Logging; reinício. |
| `KOITO_LOG_LEVEL` | `info`; `debug`, `warn`, `error`, `fatal` aceitos | Logging; reinício. Valor desconhecido resulta em `info`. |
| `KOITO_MUSICBRAINZ_URL` | `https://musicbrainz.org` | Serviço externo; validar URL. |
| `KOITO_MUSICBRAINZ_RATE_LIMIT` | `1`; inteiro | Serviço externo; a URL padrão exige valor 1. |
| `KOITO_DEFAULT_USERNAME` | `admin`; criação inicial da conta | Apenas bootstrap; não altera usuário existente. Gerenciar conta corrente na tela própria. |
| `KOITO_DEFAULT_PASSWORD` | `changeme`; criação inicial da conta; **segredo** | Apenas bootstrap; não altera senha existente. Não copiar automaticamente para TOML; alteração da conta corrente usa o fluxo existente. |
| `KOITO_DEFAULT_THEME` | vazio na configuração; interface usa `yuu` se vazio/inválido | Aparência; validar temas disponíveis. |
| `KOITO_DISABLE_DEEZER` | `false`; desativa fonte | Serviços de imagem. |
| `KOITO_DISABLE_COVER_ART_ARCHIVE` | `false`; desativa fonte | Serviços de imagem. |
| `KOITO_DISABLE_MUSICBRAINZ` | `false`; desativa fonte | Serviços de imagem. |
| `KOITO_SUBSONIC_URL` | vazio; endereço do servidor | Integração; reinício. |
| `KOITO_SUBSONIC_PARAMS` | vazio; parâmetros de autenticação; **segredo** | Integração; tratamento de segredo. |
| `KOITO_LASTFM_API_KEY` | vazio; chave Last.fm; **segredo** | Integração; tratamento de segredo. |
| `KOITO_SKIP_IMPORT` | `false`; pula importação de inicialização | Importação; reinício. |
| `KOITO_CORS_ALLOWED_ORIGINS` | vazio; origens separadas por vírgula | Segurança HTTP; validar cada origem. |
| `KOITO_DISABLE_RATE_LIMIT` | `false`; desativa limitação | Segurança HTTP; reinício. |
| `KOITO_THROTTLE_IMPORTS_MS` | `0`; intervalo em milissegundos | Importação; validar inteiro não negativo. |
| `KOITO_IMPORT_BEFORE_UNIX` | sem limite; segundos Unix positivos | Importação; validar intervalo. |
| `KOITO_IMPORT_AFTER_UNIX` | sem limite; segundos Unix positivos | Importação; validar intervalo. |
| `KOITO_FETCH_IMAGES_DURING_IMPORT` | `false`; busca imagens durante importação | Importação. |
| `KOITO_ARTIST_SEPARATORS_REGEX` | `\s+·\s+`; padrões separados por `;;` | Importação/catálogo; compilar e validar antes de salvar. |
| `KOITO_LOGIN_GATE` | `false`; exige login nas áreas protegidas pelo gate | Segurança; validar efeito em rotas e acesso antes de aplicar. |
| `KOITO_FORCE_TZ` | vazio; fuso IANA opcional | Datas e estatísticas; validar existência do fuso. |
| `KOITO_ENABLE_LBZ_RELAY` | `false`; habilita leitura do relay legado | Migração inicial para configuração por perfil no SQLite, não substitui a gestão existente por perfil. |
| `KOITO_LBZ_RELAY_URL` | vazio; URL do relay legado | Bootstrap legado; URL só é lida quando habilitado. |
| `KOITO_LBZ_RELAY_TOKEN` | vazio; token legado; **segredo** | Bootstrap legado; não revelar nem sobrescrever automaticamente dados por perfil. |

A leitura booleana atual considera verdadeiro apenas `true`, sem distinguir maiúsculas. Valores numéricos inválidos frequentemente retornam ao padrão ou a zero. O código atual só habilita Subsonic se URL e parâmetros estiverem ambos preenchidos; a verificação seguinte de configuração parcial é inalcançável, portanto um valor isolado hoje desabilita a integração silenciosamente. Uma validação mais rigorosa precisa ser planejada como alteração de comportamento, sem impedir inesperadamente a inicialização de instalações existentes.

### Legado e escopo externo

| Nome | Situação comprovada |
| --- | --- |
| `KOITO_DATABASE_URL` | Legado PostgreSQL: se preenchido, a inicialização falha. Não reativar PostgreSQL. |
| `KOITO_SQLITE_ENABLED` | Opção obsoleta; SQLite já é o banco, sem escolha de backend. |
| `KOITO_ALLOWED_HOSTS` | Lida como legado; o engine avisa que está obsoleta, sem torná-la controle de segurança novo. |
| `KOITO_ENABLE_FULL_IMAGE_CACHE` | Constante declarada, mas não consultada pelo carregador atual; não anunciar como funcional. |
| `KOITO_BASE_URL` | Apenas declaração comentada; não é opção ativa. |
| `TZ` | Variável de ambiente do runtime Go para fuso local; é diferente de `KOITO_FORCE_TZ` e não passa pelo carregador Koito. Não prometer que salvar TOML altera o ambiente do processo. |
| `KOITO_IMAGE_TAG`, `KOITO_HOST_BIND`, `KOITO_HOST_PORT`, `KOITO_DATA_DIR`, `KOITO_ADMIN_USERNAME`, `KOITO_ADMIN_PASSWORD`, `KOITO_TIMEZONE` | Interpolação do Compose, não opções lidas diretamente por `cfg`; podem definir imagem, portas, volume e variáveis encaminhadas ao contêiner. Não são editáveis pelo processo como substitutos do Compose. |
| `KOITO_VERSION` e variáveis de teste/build | Usadas na construção ou em testes; não são configurações comuns do servidor em execução. |

Outras variáveis do sistema ou de bibliotecas podem influenciar a execução, por exemplo proxy e certificados; não se deve interpretar esta lista como inventário exaustivo do runtime Go. Confirmar nomes e uso no manifesto de distribuição antes de modificar documentação de implantação.

## Proposta de funcionamento

1. Resolver primeiro o diretório por `KOITO_CONFIG_DIR`/`KOITO_CONFIG_DIR_FILE` ou padrão, e localizar `<diretório>/conf.toml`. A ausência do arquivo mantém o comportamento atual; erro de leitura ou sintaxe de arquivo existente deve produzir erro claro, sem mostrar segredos. Criar diretório no momento correto sem abrir outro banco acidentalmente.
2. Aplicar, por opção, prioridade: variável direta não vazia > conteúdo de `_FILE` > TOML > padrão atual. Preservar a semântica atual para valores vazios e falha de leitura de `_FILE`; definir explicitamente e testar a distinção entre opção TOML ausente, string vazia e booleano falso. Nenhuma migração automática de segredo ou escrita de arquivo no primeiro boot.
3. Schema tipado e versionado, organizado por rede, serviços, importação, segurança e aparência; validar configuração resultante como um conjunto. Documentar alteração intencional de validação quando for incompatível com a interpretação atual. Uma biblioteca TOML deve ser escolhida na implementação, considerando versão Go, licença e manutenção; ainda não foi adicionada dependência.
4. A tela administrativa mostrará valor salvo, origem efetiva e aviso quando uma variável externa o sobrescrever. Segredos terão apenas estado configurado/não configurado e operações explícitas manter, substituir ou limpar, sem revelar o valor. Os caminhos `_FILE` continuam controlados pelo ambiente, sem entrada de caminho arbitrário via formulário. Senha e usuário da conta ativa usam a gestão de conta existente.
5. Salvar primeiro como configuração pendente e exigir reinício externo para aplicar; a interface não deve dizer que uma opção já está efetiva. Validar antes de gravar. Nenhuma reinicialização automática. A implantação poderá, depois, especificar quais opções admitem recarga segura.
6. Usar endpoints administrativos próprios, sem ampliar o endpoint público `/apis/web/v1/config`; limitar corpo e campos, aplicar `Cache-Control: no-store` e negar autenticação por chave de ingestão. Autorizar no servidor somente a sessão da conta administrativa global existente; reaproveitar proteção de origem/CSRF e cookies do projeto. Proteger leitura e escrita, não apenas esconder menus. Registrar eventos sem valores confidenciais. A política de nova confirmação de senha para ações sensíveis deve ser definida e testada antes de implementar.
7. A integração Subsonic hoje executa verificação remota na inicialização e pode impedir o servidor de iniciar; sua URL pode conter credenciais. Antes de expor sua edição, prever validação estrutural, teste remoto explícito com prazo, redação de erros e procedimento offline para corrigir configuração inválida.
8. Escrita concorrente segura: serializar alterações, comparar revisão para detectar modificações externas, preservar arquivo anterior em erro e evitar perda silenciosa de comentários ou campos desconhecidos. Arquivo temporário no mesmo diretório, permissões restritas (como `0600`), sincronização/renomeação apropriadas e tratamento explícito de links simbólicos e falhas; testar portabilidade e permissões dos volumes.

### Decisões que precisam de definição antes da implementação

- O pedido inclui editar **todos** os dados pelo master, mas diretório de dados, variáveis do Compose e bootstrap de credenciais têm efeitos diferentes das demais opções. A proposta é mostrar essas origens e usar a tela de conta para credenciais atuais; decidir se haverá fluxo específico de realocação de dados e edição de bootstrap, com instruções e confirmação adequadas. Não anunciar essas opções como editáveis até existir tal fluxo.
- Segredos no TOML exigem armazenamento em texto legível pelo processo e proteção do arquivo/backup; alternativa é manter `_FILE` para operação externa. Definir se a interface poderá salvar novos segredos em TOML ou apenas orientar o administrador para um mecanismo de segredos. Nunca substituir silenciosamente uma referência externa.
- Garantir a mesma descoberta e precedência no comando `reset-password` e na execução normal; prever recuperação offline.
- Definir formato de preservação de comentários e campos futuros, tratamento de TOML desconhecido, política de compatibilidade e eventual aviso de depreciação das variáveis. A manutenção de compatibilidade com Compose, Docker, VS Code, scripts e `_FILE` é o ponto de partida; não remover variáveis nesta etapa.

## Sequência de execução

As issues locais [038](issues/038-contrato-configuracao.md), [039](issues/039-leitura-configuracao-toml.md), [040](issues/040-persistencia-configuracao.md), [041](issues/041-administracao-configuracao.md) e [042](issues/042-aceite-migracao-configuracao.md) descrevem contrato, leitura, escrita segura, interface e aceite. São propostas abertas, sem publicação e sem implementação. A issue 038 resolve as decisões pendentes antes das seguintes.

O aceite deve cobrir instalação sem arquivo, instalações existentes e novas, precedência e `_FILE`, reinício, segredo não revelado, acesso não autorizado e CSRF, validação, escrita concorrente/interrompida, permissões e arquivo manual, Compose/VS Code, fuso, importação, relay por perfil e persistência do mesmo SQLite v3. Capturas desktop e celular serão necessárias quando a tela for criada.

## Revisão do Advisor

Parecer: **reprovado com ressalvas para implementação imediata**. O Advisor considerou a proposta viável, mas exigiu definir os limites de bootstrap, o armazenamento de segredos e a proteção das operações antes de programar. Seus apontamentos sobre Subsonic, endpoint público, recuperação offline, concorrência e reinício foram incorporados ao plano e às issues. A issue 038 deve resolver as decisões pendentes antes da implementação.
