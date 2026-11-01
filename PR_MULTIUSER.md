# Add public listening profiles with API-key isolation and single-admin management

## Description

Allow several people to scrobble to one Koito instance while keeping separate listening histories and the existing interface. Listeners select a public profile without signing in; one password-protected administrator manages profiles and credentials.

- Resolve each incoming ListenBrainz API key to exactly one profile. Scope history, statistics, charts, search, now-playing, imports and exports to that profile while sharing catalog metadata.
- Add `/` profile selection, `/<slug>` profile routes and a sidebar selector with names and avatars.
- Add profile, avatar and API-key management inside **Settings → Administration**. Player API keys cannot authorize administrative changes.
- Add per-profile ListenBrainz forwarding with a visible external token, saved-state feedback and connection testing without saving credentials or submitting listens.
- Migrate SQLite schema 2 → 3 transactionally, preserving existing credentials and history. A single existing user's profile becomes `/default`.

**Upgrade:** back up the database first. Migration rollback requires restoring the pre-migration backup. Legacy unscoped reads return `409` when profile selection is ambiguous. Profiles and listening data are public to anyone with access to the instance; private profiles are outside this implementation.

## Related Issue(s)

None linked.

## Type of Change

- [ ] Bug fix
- [x] New feature
- [x] Breaking change — profile-scoped routes, administrative authorization and schema migration
- [x] Documentation update

## How Has This Been Tested?

- [x] Unit Tests — full Go suite, migration/isolation/authorization tests, relay tests using local mock servers, frontend type checking and six frontend tests passed. Production frontend build passed.
- [x] Manual Testing — local validation recorded in the design documents covers fresh installation, upgrading a copy of a schema-2 database, restart persistence, two independent browser tabs, profile switching, imports/exports and key revocation.

Migration and HTTP validation used disposable databases or copies. External relay delivery was simulated locally.

## Checklist

- [ ] I have performed a self-review of my own code
- [x] I have made corresponding changes to the documentation
- [x] I have ensured my changes generate no new warnings

Validation details: [PR_VALIDATION.md](PR_VALIDATION.md). The application checks and Dockerfile check are clean. Documentation builds pass with the same existing Node DEP0190 warning as the base; existing installer warnings are recorded separately.

## AI/LLM Usage

- [x] I disclose that approximately 70% of the implementation work used OpenAI Codex assistance; this is an estimate, not a measured percentage of generated lines.
- [ ] I fully understand all changes made by LLM-generated code.
- [ ] I have not and will not use an LLM to write any part of this PR description or PR comments.

This description was also drafted with OpenAI Codex. The final checkbox is intentionally unchecked. This draft does not satisfy that upstream requirement; maintainer agreement is needed before submission. Author-review declarations remain unchecked for the contributor to assess.

## Screenshots (if applicable)

To attach: profile selector and Settings → Administration showing profile and forwarding controls.
