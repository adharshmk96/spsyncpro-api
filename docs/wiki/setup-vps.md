## Setup VPS Connection

1. generate ssh key locally specifically for the vps

```bash
ssh-keygen -t ed25519 -C "your_email@example.com" -f ~/.ssh/_servername_
```

2. add ssh key to vps

```bash
ssh-copy-id -i ~/.ssh/_servername_ root@_vps_ip_
```

3. configure to use the ssh key for the vps

```
Host _servername_
    HostName _vip_ip_
    User username
    IdentityFile ~/.ssh/_servername_
```

## Setup Docker Context

1. setup a named context

```bash
docker context create vps1 --docker "host=ssh://_servername_"
```