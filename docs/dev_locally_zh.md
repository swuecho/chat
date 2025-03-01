## 本地开发指南

1. 克隆仓库
2. Golang 开发环境

```bash
cd chat; cd api
go install github.com/cosmtrek/air@latest
go mod tidy

# 根据你的环境设置环境变量
export DATABASE_URL= postgres://user:pass@192.168.0.1:5432/db?sslmode=disable

# 如果使用 `debug` 模型则不需要设置
# export OPENAI_API_KEY=sk-xxx
# export OPENAI_RATELIMIT=100

make serve
```

3. Node.js 开发环境

```bash
cd ..; cd web
npm install
npm run dev
```

4. 端到端测试

```bash
cd ..; cd e2e
# 根据你的环境设置环境变量
export DATABASE_URL= postgres://user:pass@192.168.0.1:5432/db?sslmode=disable
npm install
npx playwright test # --ui 
```

如有疑问，请在 issue 或 discussion 中提问。
