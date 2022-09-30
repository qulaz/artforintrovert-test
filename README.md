# artforintrovert-test

Тестовое задание для компании [Правое полушарие Интроверта](https://online.artforintrovert.ru/).

## Задание

Разработать API в котором 3 эндпоинта. 

Есть одна сущность. При старте сервера все записи из базы подгружаются в оперативку,


Реализовать:
- Эндпоинт удаления - удаляет в базе запись
- Эндпоинт изменения - изменяет в базе запись
- Эндпоинт показать - показывает текущий актуальный список
- Процесс, который подгружает записи с базы в оперативку, с определенным интервалом

Предпочтительная база: MongoDB

### Документация

1. proto-файлы сервиса - [api/v1](api/v1)
2. swagger-файл для REST API - [gen/api/v1/product.swagger.json](gen/api/v1/product.swagger.json).
   Посмотреть можно вставив содержимое файла в редактор [https://editor.swagger.io/](https://editor.swagger.io/)

### Предварительные требования

Для запуска должен быть установлен Docker и docker-compose.

## Запуск

1. Создать `.env` файл с конфигурацией на основе [`.env.dist`](.env.dist) файла. 
   По-умолчанию заданы оптимальные значения с которым API запустится и будет работать. 
   При желании переменные можно подкорректировать.
   ```bash
   # bash
   cp .env.dist .env
   ```
2. Собрать контейнеры `docker-compose build`
3. Заполнить базу данных тестовыми данными: 
   ```bash
   docker-compose run --rm app /usr/local/bin/populate
   # Creating network "preavor-polusharie-introverta-test_default" with the default driver
   # Creating preavor-polusharie-introverta-test_mongodb_1 ... done
   # Creating preavor-polusharie-introverta-test_jaeger_1  ... done
   # Creating preavor-polusharie-introverta-test_app_run   ... done
   # Created 50000 products
   ```
4. Запустить контейнеры: `docker-compose up`

После успешного выполнения этих действий будет запущено:
1. GRPC-сервер на порту `50051`
2. GRPC Gateway (REST API) сервер на порту `8000`
3. MongoDB на порту `27017`
4. `Jaeger UI` c трейсами запросов по адресу [http://localhost:16686/](http://localhost:16686/)

## Команды для разработки

В проекте присутствует [`Makefile`](Makefile) с полезными командами.

- `make lint` - запустит `golangci-lint` для проекта
- `make protogen` - запустит `buf generate` и сгенерит сервер из proto-файлов
