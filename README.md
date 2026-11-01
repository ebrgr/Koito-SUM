<div align="center">
  
![Koito logo](https://github.com/user-attachments/assets/bd69a050-b40f-4da7-8ff1-4607554bfd6d)

_Koito (小糸) is a Japanese surname. It is also homophonous with the words 恋と (koi to), meaning "and/with love"._

</div>

> **S**imple **U**ser **M**anager
>
> [Listening profiles and upgrade guide](docs/src/content/docs/guides/profiles.md) · [Player keys and per-profile forwarding](docs/src/content/docs/guides/scrobbler.md)
>
> Current source version: **Koito 0.3.2-SUM-RC2**. Run `make version.bump` once before each subsequent commit; `make version.check` checks synchronized versions.
>
> This version adds multiuser scrobble submissions while keeping the experience simple. Approximately 70% of the work on these changes was done with assistance from OpenAI Codex.
>
> **Back up your database before testing. This version can upgrade an existing installation that has already migrated to SQLite; use its existing data directory so the database is migrated automatically on startup.**


<div align="center">
  
  [![Ko-Fi](https://img.shields.io/badge/Ko--fi-F16061?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ko-fi.com/gabehf)
  
</div>

Koito is a modern, themeable ListenBrainz-compatible scrobbler for self-hosters who want control over their data and insights into their listening habits.
It supports relaying to other compatible scrobblers, so you can try it safely without replacing your current setup.

> This project is under active development and still considered "unstable", which means the main branch can be unusable at times, and some breaking changes may occur. If you don't want to replace your current scrobbler
> with Koito quite yet, you can [set up a relay](https://koito.io/guides/scrobbler/#set-up-a-relay) from Koito to another ListenBrainz-compatible
> scrobbler. This is what I've been doing for the entire development of this app and it hasn't failed me once. Or, you can always use something
> like [multi-scrobbler](https://github.com/FoxxMD/multi-scrobbler).

## Features

- ⚡ More performant than similar software
- 🖌️ Sleek UI
- 🔁 Compatible with anything that scrobbles to ListenBrainz
- 🔌 Easy relay to your existing setup
- 📂 Import support for Maloja, ListenBrainz, LastFM, and Spotify

## Demo

You can view my public instance with my listening data at https://koito.mnrva.dev

## Screenshots

![screenshot one](assets/screenshot1_2.png)
![screenshot two](assets/screenshot2.png)
![screenshot three](assets/screenshot3.png)

## Installation

For this multiuser fork, see [DEPLOYMENT.md](DEPLOYMENT.md). The release image is
`ghcr.io/ebrgr/koito-sum:0.3.2-SUM-RC2`, targeting Linux amd64 and arm64. This source version requires a matching published image; until publication, follow the local build instructions in [DEPLOYMENT.md](DEPLOYMENT.md).

```sh
cp deploy/.env.example deploy/.env
# Set the administrator password and your data directory in deploy/.env first.
docker compose --env-file deploy/.env -f deploy/compose.yaml pull
docker compose --env-file deploy/.env -f deploy/compose.yaml up -d
```

Back up an existing SQLite v2 installation before pointing the volume at its
existing data directory. Migration to schema v3 runs automatically at startup.
The [English multiuser design](MULTIUSER_DESIGN.en.md) describes the model,
permissions, migration, and validation results.

Original Koito documentation: [installation guide](https://koito.io/guides/installation/).

## Importing Data

See the [data importing guide](https://koito.io/guides/importing/) in the docs.

Before importing, it is recommended you set up any image sources you want to have enabled. You can view the configuration options for image sources in the full list of configuration options.

## Full list of configuration options

See the [configuration reference](https://koito.io/reference/configuration/) in the docs.

## Contributing

For local editing, running, and debugging (Go API + React frontend), see
[DEVELOPMENT.md](DEVELOPMENT.md).

There are currently some known issues that I am actively working on, in addition to adding new features. If you want to contribute (especially more and more robust testing) feel free to fork the repository and make a pull request with your changes.

If you have any feature ideas, open a GitHub issue to let me know. I'm sorting through ideas to decide which data visualizations and customization options to add next.

You can also donate money to me, @gabehf, the primary developer, via the "sponsor" or "ko-fi" buttons, and I will spend that money on luxurious trips and fine cuisine in foreign countries, or rent.

## Albums that fueled development + notes

More relevant here than any of my other projects...

Not just during development, you can see my complete listening data on my [live demo instance](https://koito.mnrva.dev).

#### Random notes

- I find it a little annoying when READMEs use emoji but everyone else is doing it so I felt like I had to...
- About 50% of the reason I built this was minor/not-so-minor greivances with Maloja. Could I have just contributed to Maloja? Maybe, but I like building stuff and I like Koito's UI a lot more anyways.
