Сервис для переводов
Деплой
Для запуска сервиса, следуйте этим шагам:

Сконфигурируйте базу данных в файле docker-compose в папке database, чтобы настроить ее под ваши требования.

Запустите базу данных с помощью следующей команды:

bash
Copy code
docker-compose -f database/docker-compose.yml up -d
Выполните SQL-скрипты из папки database, чтобы настроить базу данных.

Сконфигурируйте файл docker-compose.yml для сервера, задав переменные окружения, включая:

PG_HOST=postgres_container - хост базы данных PostgreSQL.
PG_PORT=5432 - порт базы данных PostgreSQL.
PG_USER=admin - имя пользователя базы данных.
PG_PASSWORD=admin - пароль пользователя базы данных.
PG_DBNAME=bank - имя базы данных.
WORKER_NUMBER=3 - количество воркеров для обработки запросов.
Запустите сервис с помощью следующей команды:

bash
Copy code
docker-compose -f docker-compose.yml up -d
Endpoints
POST ip:port/transfer
Headers:

Authorization: Bearer token - токен, который указан в базе данных. Приложение использует этот токен для идентификации отправителя перевода.
Body form-data:

to: int - ID клиента, которому нужно отправить деньги.
amount: double - сумма, которую нужно перевести.
Пример использования:

bash
Copy code
curl -X POST -H "Authorization: Bearer your_token" -F "to=recipient_id" -F "amount=100.50" ip:port/transfer
Замените your_token на реальный токен и укажите recipient_id и сумму перевода.