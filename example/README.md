## 1 Инициализация

### 1.1 Создать файл .env и заполнить его значениями из .env.example.
```shell script
cp ./.env.example ./.env
```

### 1.2 Выполнить
```shell script
hostname -I
```

### 1.3 Заменить в .env hostip полученным значением IP хоста.
Например, в случае вывода 192.168.0.5 192.168.122.1 172.17.0.1 ... меняем hostip на 192.168.0.5

Также заменить hostip в файлах
```
ms-user/charts/configmap.yaml
ms-policy/charts/configmap.yaml
ms-api/charts/configmap.yaml
```

### 1.4 собрать образ с minikube
```shell script
docker build -t minikube .
```

## 2 Запуск
### 2.1 После инициализации согласно 1 запустить контейнеры
```shell script
docker compose up -d
```

### 2.2 Запустить сервисы
```shell script
docker exec -it test-kuber make start
```

### 2.3 Проверить, что сервисы стартовали без ошибок
```shell script
docker exec -it test-kuber kubectl get pods
```

При успешном запуске должно подняться без ошибок 6 подов с тремя сервисами
```
NAME                         READY   STATUS    RESTARTS   AGE
mse-api-85c4447bbf-mjl6v     1/1     Running   0          30s
mse-api-85c4447bbf-zcxm6     1/1     Running   0          30s
mse-policy-d4df9c899-97r95   1/1     Running   0          31s
mse-policy-d4df9c899-wfx7b   1/1     Running   0          31s
mse-user-5d587b584-fbts7     1/1     Running   0          31s
mse-user-5d587b584-m7n6n     1/1     Running   0          31s
```

### 2.4 Заполнить БД сервиса ms-policy. Для этого выполнить команду инициализации для любого из запущенных подов сервиса (см. выше) - выполняется однократно при первом запуске
```shell script
docker exec -it test-kuber kubectl exec mse-policy-d4df9c899-97r95 -- ./policy admin init
```

## 3 Отключение
### 3.1 Для запущенного проекта выполнить
```shell script
docker exec -it test-kuber minikube delete
```

### 3.2 Отключить контейнеры
```shell script
docker compose down
```

## 4 Проверка работоспособности
### 4.1 Для запущенного проекта выполнить добавление пользователей
```shell script
curl -d '{"email":"test_user@mail.com", "password": "123456789i"}' -H "Content-Type:application/json" -X POST http://192.168.49.2:32300/api/v1/sign
```
```shell script
curl -d '{"email":"test_user2@mail.com", "password": "123456789i"}' -H "Content-Type:application/json" -X POST http://192.168.49.2:32300/api/v1/sign
```

### 4.2 Добавить одному из пользователей роль администратора
```shell script
docker exec -it test-kuber kubectl exec mse-policy-d4df9c899-97r95 -- ./policy admin assign-user test_user@mail.com admin
```

### 4.3 Авторизовать пользователя с ролью администратора
```shell script
curl -d '{"email":"test_user@mail.com", "password": "123456789i"}' -H "Content-Type:application/json" -X POST http://192.168.49.2:32300/api/v1/login
```

### 4.4 Выполнить запрос списка пользователей с токеном авторизации, который был получен в п.4.3 (должен вернуться список из двух пользователей)
```shell script
curl -H "Content-Type:application/json" -H "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDUxNzI4NjIsInN1YiI6IjJkN2UxNmM4LTAwM2QtNDJiNy1iMDBhLWRhMjVhZTUyNWZiZiJ9.TxqMFAkkN5w82gge41VVlW2yMam6p4ixrc14WYuMo_g" -X POST http://192.168.49.2:32500/api/v1/users
```

### 4.5 Авторизовать пользователя без роли администратора
```shell script
curl -d '{"email":"test_user2@mail.com", "password": "123456789i"}' -H "Content-Type:application/json" -X POST http://192.168.49.2:32300/api/v1/login
```

### 4.6 Выполнить запрос списка пользователей с токеном авторизации, который был получен в п.4.5 (должна вернуться ошибка доступа)
```shell script
curl -H "Content-Type:application/json" -H "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDUxNzI5MzAsInN1YiI6Ijc0YTdkYWIzLTAyY2QtNGE5Yi04YjI1LWUzMjNkZjlmZmUwNCJ9.7DfM1F5iPOkNV7Ise9EvxJDUHgdO52L41kqeTLhG3nE" -X POST http://192.168.49.2:32500/api/v1/users
```