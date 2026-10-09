# Noop relay

`noop` is the maintained deployment branch. `main` follows the original project;
upstream changes are synced into `main`, then `noop` is rebased onto `main`.

Build the relay from this checkout using the existing PostgreSQL and Redis
stack. Keep the same Compose project name and private environment file when
updating an existing deployment so its named volumes remain attached:

```sh
docker compose --env-file /path/to/private/runtime.env -p noop-sub2api \
  -f docker-compose.yml -f docker-compose.noop.yml up -d --build
```

Run from `deploy/`. The override binds the application to loopback by default
and uses `RUN_MODE=simple`: Sub2API records usage without charging a customer
balance. Noop Payment owns customer credit settlement. It also disables the
upstream per-user API key create caps (`API_KEY_CREATE_*`), because every Noop
user key is provisioned under the single relay service user.

Use the built-in OpenAI OAuth transport. No transport plugin is required.
Create an OpenAI OAuth account in the admin console, complete the authorized
account login, and create a relay API key in the same OpenAI group. The relay
key belongs on the Noop backend, never in the web client or user sandbox.

Leave the account proxy empty for direct upstream access. To use an existing
HTTP/SOCKS proxy later, add it under Proxies and select it on the account.
`localhost` inside a container refers to that container, not the host proxy.

Validate a real `/v1/responses` request with `stream=true` and `store=false`,
then a Codex CLI tool round trip using the relay key and `/v1` base URL.
OAuth HTTP requests must carry the conversation and tool results in `input`;
do not rely on server-side response storage or `previous_response_id`.

The OAuth upstream does not enforce `max_output_tokens`. Noop must use final
provider usage for settlement and preserve its billing hold when usage is
missing or exceeds the reserved credits.
