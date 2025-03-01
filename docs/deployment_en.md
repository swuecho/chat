## How to Deploy

Refer to `docker-compose.yaml`

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/tk7jWU?referralCode=5DMfQv)

Then configure the environment variables.

```
PORT=8080
OPENAI_RATELIMIT=0
```

Fill in the other two keys if you have them.

<img width="750" alt="image" src="https://user-images.githubusercontent.com/666683/232234418-941c9336-783c-4430-857c-9e7b703bb1c1.png">

After deployment, registering users, the first user is an administrator, then go to
<https://$hostname/#/admin/user> to set rate limiting. Public deployment,
only adds rate limiting to trusted emails, so even if someone registers, it will not be available.

<img width="750" alt="image" src="https://user-images.githubusercontent.com/666683/232227529-284289a8-1336-49dd-b5c6-8e8226b9e862.png">

This helps ensure only authorized users can access the deployed system by limiting registration to trusted emails and enabling rate limiting controls.
