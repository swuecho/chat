## 如何部署

参考 `docker-compose.yaml`

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/tk7jWU?referralCode=5DMfQv)

然后配置环境变量就可以了.

```
PORT=8080
OPENAI_RATELIMIT=0
```

别的两个 api key 有就填.

<img width="750" alt="image" src="https://user-images.githubusercontent.com/666683/232234418-941c9336-783c-4430-857c-9e7b703bb1c1.png">

部署之后,  注册用户, 第一个用户是管理员, 然后到  <https://$hostname/#/admin/user>,
设置 ratelimit, 公网部署, 只对信任的email 增加 ratelimit, 这样即使有人注册, 也是不能用的.

<img width="750" alt="image" src="https://user-images.githubusercontent.com/666683/232227529-284289a8-1336-49dd-b5c6-8e8226b9e862.png">
