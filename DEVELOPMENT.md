# Desenvolvimento local

Koito registra o histórico de músicas (scrobbles), recebe eventos pela API
compatível com ListenBrainz e apresenta estatísticas em uma interface web.
Original: <https://github.com/gabehf/Koito>. O remoto `origin` aponta para o fork;
`upstream`, para o original. Não é necessário Docker nem PostgreSQL para desenvolver.

## Preparação

Pré-requisitos: Go compatível com `go.mod` (Go 1.25 ou superior, incluindo os
requisitos das dependências), Node.js 24 LTS, Corepack, compilador C, pkg-config e libvips.
No macOS:

```sh
xcode-select --install # somente se as ferramentas de compilação estiverem ausentes
brew install vips pkg-config
# Se o Node instalado não incluir Corepack:
npm install -g corepack
make setup
```

Em Debian/Ubuntu, as bibliotecas nativas são `libvips-dev`, `pkg-config` e
`build-essential`. `make doctor` verifica as ferramentas disponíveis.

`make setup` preserva um `.env` existente, instala os módulos Go, usa o Yarn
4.14.1 declarado em `client/package.json` com lockfile imutável e instala o
Delve em `.bin/dlv`. Não precisa habilitar Yarn globalmente: os comandos usam
`corepack yarn`. Execute os comandos a partir da raiz do repositório.

## Executar e editar

Em dois terminais:

```sh
# Terminal 1: API
make api.dev
```

```sh
# Terminal 2: interface com atualização automática (HMR)
make client.dev
```

Abra <http://localhost:5173> e escolha um perfil. A raiz sempre mostra a seleção;
o painel fica em `/<slug>`, por exemplo `/default`. A administração é acessível
em `/admin`. O Vite encaminha `/apis` e `/image` para a API
em `127.0.0.1:4110`, inclusive cookies de autenticação. O login inicial é
`admin` / `changeme`, conforme `.env`; essas variáveis só criam o usuário
quando o banco ainda está vazio. O Go precisa ser reiniciado após alterações:
Ctrl+C e `make api.dev` novamente.

O `.env` é carregado pelo Makefile e pelo perfil de depuração da API. Executar
`go run` diretamente não carrega esse arquivo. A configuração local limita a API
ao loopback, desliga importação automática e consultas externas de metadados/imagens.
Para trabalhar nessas integrações, altere as variáveis correspondentes no `.env`.

O SQLite e os arquivos de trabalho ficam em `.local/koito`, ignorado pelo Git.
As migrações são aplicadas automaticamente na inicialização. O esquema atual
é a versão 3: a migração `003_listening_profiles.sql` reúne perfis e relay em
uma única transação para atualizar bancos v2. Não é necessário
configurar `KOITO_DATABASE_URL` nem `KOITO_SQLITE_ENABLED` nesta versão.
Para importar dados de teste, use cópias em `.local/koito/import` e defina
`KOITO_SKIP_IMPORT=false`. Veja também a documentação de importação em
`docs/src/content/docs/guides`.

Verificação rápida:

```sh
curl -f http://127.0.0.1:4110/apis/web/v1/health
curl -f http://localhost:5173/apis/web/v1/health
```

## Administrar perfis

Abra a aba **Settings → Administration** e entre com a conta administrativa
do banco atual. A gestão abre dentro do painel de Settings, mantendo a página
pública e o período selecionados; `/admin` também permite acesso direto. Em “Manage
profile”, escolha uma pessoa ou “Create a profile”. O cadastro cria uma chave
inicial; nessa tela você pode copiar o endpoint e o token, criar/revogar chaves,
alterar nome/slug e escolher uma foto ou ícone. Ouvintes não precisam de senha.

“Import history” envia um arquivo para o perfil selecionado e exige confirmação
do destino. Suporta Koito, Maloja, Spotify e Last.fm em JSON, e ListenBrainz em
ZIP de exportação. Limite de upload: 100 MiB por requisição. O processamento é
síncrono e um arquivo inválido pode deixar registros parcialmente importados.
Para arquivos maiores, use o diretório `import/<profileID>/` e o endpoint JSON
com `filename`, conforme [MULTIUSER_DESIGN.md](MULTIUSER_DESIGN.md).

“Review deletion” mostra nome e contagem de escutas e pede digitar o nome. A
exclusão remove somente o perfil e seus dados pessoais; o catálogo e a conta
administrativa continuam disponíveis, inclusive após excluir o último perfil.


### Validar o fluxo multiusuário

```sh
make multiuser.check
```

Requer Python 3, compila API/frontend e testa a jornada completa com bancos
SQLite descartáveis, instalação nova, migração v2 e reinício da API. Não usa o
banco de `.env` e não contata serviços musicais externos. Para incluir uma cópia
de um banco de exemplo já existente:

```sh
python3 scripts/validate_multiuser.py --binary ./koito --example-db /caminho/koito.db
```

O exemplo é aberto somente para leitura e as operações administrativas ocorrem
na cópia. Credenciais de teste não são impressas. Veja os resultados e os limites
no registro do passo 5 de [MULTIUSER_DESIGN.md](MULTIUSER_DESIGN.md).


## Depurar

No VS Code, abra esta pasta e instale a extensão Go recomendada. Depois de
`make setup`, selecione **Koito: Desenvolvimento (localhost:5173)** em Executar
e Depurar e pressione F5. Isso inicia o Vite como tarefa e a API com Delve,
abrindo o navegador em `http://localhost:5173` quando a API estiver pronta.
Para depurar também o TypeScript no VS Code, use **Koito: API + Frontend**,
que abre uma sessão de Chrome com depuração. Também é possível selecionar
os perfis separadamente. Pare instâncias manuais
da API/Vite antes de usar os perfis para evitar conflito de portas.

Pontos úteis para breakpoints:

- `cmd/api/main.go`: entrada e leitura das variáveis de ambiente.
- `engine/engine.go`: inicialização, banco e tarefas de fundo.
- `engine/routes.go`: associação de rotas e middlewares.
- `engine/handlers/lbz_submit_listen.go`: recebimento de scrobbles.
- `engine/handlers/auth.go`: login e sessões.
- `client/api/api.ts`: requisições feitas pela interface.
- `client/app/routes/Home.tsx`: tela inicial.

O navegador usa source maps de desenvolvimento para breakpoints TS/TSX.
Ao finalizar, encerre a tarefa **Koito: frontend dev** no terminal do VS Code
se ela continuar ativa. No terminal, `make api.debug` abre o Delve interativo;
use `break main.main`, `continue` e `exit`. Não execute dois servidores na porta 4110.

### Qual endereço abrir

- `http://localhost:5173`: interface em desenvolvimento, com HMR; requer Vite.
- `http://127.0.0.1:4110`: API e interface do último `make client.build`.
- `DAP server listening at: 127.0.0.1:XXXXX`: conexão interna do depurador,
  com porta variável a cada sessão. Esse endereço não é um site.

O perfil **Koito: API (Go)** inicia somente a API. Se usar esse perfil e tentar
abrir a porta 5173, será necessário executar `make client.dev` separadamente.
As portas HTTP são fixas: se estiverem ocupadas, os servidores falham em vez
de escolher outra porta. Pare a execução anterior antes de iniciar outra.

## Mapa do código

| Diretório | Responsabilidade |
| --- | --- |
| `cmd/api` | Executável Go e comando `reset-password` |
| `engine` | Servidor HTTP Chi, rotas, autenticação, handlers e testes de integração |
| `internal/cfg` | Configuração por ambiente |
| `internal/catalog` | Regras do catálogo: artistas, álbuns, faixas, aliases e merges |
| `internal/db/sqlite` | Persistência e consultas SQL |
| `db/migrations_sqlite` | Migrações SQLite incorporadas ao binário |
| `internal/importer` | Importadores de históricos musicais |
| `internal/mbz`, `internal/images`, `imagecache` | Metadados, fontes de imagens e cache/libvips |
| `queue` | Processamento de tarefas em fila |
| `client/app` | React 19, rotas React Router 7, componentes, temas e estado |
| `client/api` | Cliente HTTP e tipos compartilhados pela interface |
| `docs` | Site de documentação Astro/Starlight |
| `test_assets` | Dados usados pelos testes |

Fluxo principal: player → `/apis/listenbrainz/1/submit-listens` → autenticação
por chave de API → catálogo → SQLite. A interface consulta `/apis/web/v1`;
sessões de login ficam em cookies. Em desenvolvimento o Vite serve a interface;
após o build, o próprio Go serve `client/build/client`.

## Validar e compilar

```sh
make check          # testes Go + geração de tipos React Router + TypeScript
make build          # binário ./koito e interface client/build/client
make api.dev        # também serve a interface compilada em http://127.0.0.1:4110
```

`make api.build` gera um binário otimizado sem símbolos de depuração; use Delve
pelos perfis ou `make api.debug` para depurar. `make client.build` gera os assets
e suas versões gzip. Execute o servidor a partir da raiz, pois assets e interface
são localizados a partir do diretório de trabalho. Para documentação, instale as
dependências em `docs` e use `make docs.dev` (projeto separado do cliente).

Se mudar as portas, ajuste `.env`, os proxies em `client/vite.config.ts` e os
endereços de `.vscode/launch.json` e `.vscode/tasks.json` juntos. O Vite usa porta
fixa e falha se 5173 estiver ocupada, evitando abrir um endereço diferente do depurador.

Para consultar alterações do original sem modificar sua branch:

```sh
git fetch upstream
git log --oneline HEAD..upstream/main
git diff HEAD upstream/main
```


## Interface por perfil

`make check` também executa `make client.test`, com os testes de navegação e
isolamento do cliente HTTP no runner nativo do Node 24. O build continua em
`make client.build`.

Para validar com outra API sem mudar o `.env` normal, o Vite aceita um destino
explícito, por exemplo `KOITO_DEV_API_TARGET=http://127.0.0.1:14110`. A porta
padrão da interface continua 5173, com `strictPort`; uma porta alternativa só é
usada se passada explicitamente ao comando do Vite.
