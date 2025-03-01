## Local Development Guide

1. Clone the repository
2. Golang development

```bash
cd chat; cd api
go install github.com/cosmtrek/air@latest
go mod tidy
# Set environment variables based on your environment
export DATABASE_URL= postgres://user:pass@192.168.0.1:5432/db?sslmode=disable

# Not required if using `debug` model
# export OPENAI_API_KEY=sk-xxx
# export OPENAI_RATELIMIT=100

make serve
```

3. Node.js development

```bash
cd ..; cd web
npm install
npm run dev
```

4. End-to-end testing

```bash
cd ..; cd e2e
# Set environment variables based on your environment
export DATABASE_URL= postgres://user:pass@192.168.0.1:5432/db?sslmode=disable

npm install
npx playwright test # --ui 
```

Ask in issue or discussion if unclear.
