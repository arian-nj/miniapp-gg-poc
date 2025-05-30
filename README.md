# Godot Js Golang Telegram Miniapp Proof of concept

## Usage

### build front code:
```bash
cd frontend && npm run build
```

### serve front files and run bot:
```bash
LISTEN_ADDRESS=0.0.0.0:3000 TG_BOT_TOKEN=<BOT_TOKEN> go run ./cmd/main
```

### run caddy for https
```bash
sudo caddy run --config ./Caddyfile
```

### miniapp url
set miniapp url in bot father setting https://127.0.0.1:3000
