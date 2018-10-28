# AuthN Service

## 启动服务

第一次启动，如果 `../../config/authn/keys` 目录里的密钥对没有，会引起 ga 启动读取公钥文件失败。
所以，第一次启动服务前，运行下面命令保证密钥对存在（不存在会创建）：

```
docker-compose run api python3 manage.py gentokenkeys
```

启动服务：

```
docker-compose up -d
```

查看服务是否启动正常

```
docker-compose ps
```


## 创建 Admin App

创建具有 admin 角色权限的 App :

```
docker-compose exec api python3 manage.py syncdb
docker-compose exec api python3 manage.py createadminapp
```
