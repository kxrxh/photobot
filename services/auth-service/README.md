# AuthService

Auth for photobox: messenger initData, web login/password, service tokens.

## Local dev

```bash
make dev
```

First run: creates `compose.env`, RSA keys, starts Postgres + Redis.

`DEV_MODE=true` seeds user `dev` / `dev` and relaxes signature checks.
