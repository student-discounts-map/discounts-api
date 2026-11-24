# Архитектура

Схемы построены по нотации C4 (уровни Context и Container), для компонентов бэкенда использована обычная схема, для сквозного сценария диаграмма последовательности. Стек: Go, PostgreSQL + PostGIS, Redis, MinIO (S3), React (Next.js).

## 1. Контекст

```mermaid
C4Context
    title Системный контекст "Карта студенческих скидок"

    Boundary(users, "Пользователи") {
        Person(guest, "Гость", "Смотрит карту, ищет, открывает карточки")
        Person(student, "Студент", "Добавляет места, жалуется, запрашивает роль")
        Person(representative, "Представитель", "Тестовая роль: предлагает правки своих карточек")
        Person(moderator, "Модератор", "Рассматривает заявки, жалобы, запросы ролей")
    }

    Boundary(sys, "Система и внешние сервисы") {
        System_Ext(mapProvider, "Картографический сервис", "Тайлы карты и геокодирование")
        System(discountMap, "Карта скидок", "Каталог, поиск, заявки, модерация, RBAC, журнал")
        System_Ext(mailService, "Почтовый сервис", "Письма со ссылкой для входа")
    }

    Rel(guest, discountMap, "Смотрит карту")
    Rel(student, discountMap, "Заявки и жалобы")
    Rel(representative, discountMap, "Правки карточек")
    Rel(moderator, discountMap, "Решения")
    Rel(discountMap, mapProvider, "Карта")
    Rel(discountMap, mailService, "Письма")

    UpdateRelStyle(discountMap, mapProvider, $offsetY="-30", $offsetX="-35")
    UpdateRelStyle(discountMap, mailService, $offsetY="-30", $offsetX="-10")
    UpdateLayoutConfig($c4ShapeInRow="4", $c4BoundaryInRow="1")
```

## 2. Контейнеры

```mermaid
C4Container
    title Контейнеры "Карта студенческих скидок"

    Person(user, "Гость, студент, представитель", "Смотрит карту, подаёт заявки, жалобы, правки")
    Person(moderator, "Модератор", "Рассматривает заявки, жалобы, запросы ролей")
    System_Ext(mapProvider, "Картографический сервис", "Тайлы карты и геокодирование")
    System_Ext(mailService, "Почтовый сервис", "Письма для входа")

    System_Boundary(system, "Карта скидок (Docker Compose)") {
        Container(web, "Web App", "React, TypeScript, Vite", "Карта, поиск, формы, интерфейс модератора")
        Container(api, "Backend API", "Go, REST, OpenAPI", "RBAC, rate limiting, валидация, заявки, поиск, дубли, аудит")
        ContainerDb(redis, "Redis", "Redis", "Кэш, лимиты, сессии, очередь")
        Container(worker, "Worker", "Go", "Обработка фото, автоустаревание скидок")
        ContainerDb(db, "База данных", "PostgreSQL, PostGIS", "Пользователи, места, заявки, журнал")
        ContainerDb(storage, "Хранилище файлов", "MinIO (S3)", "Проверенные фото")
    }

    Rel(user, web, "Использует", "HTTPS")
    Rel(moderator, web, "Модерирует", "HTTPS")
    Rel(web, api, "API", "HTTPS")
    Rel(web, mapProvider, "Тайлы", "HTTPS")
    Rel(api, mapProvider, "Адреса", "REST")
    Rel(api, mailService, "Письма", "SMTP")
    Rel(api, redis, "Redis")
    Rel(api, db, "SQL")
    Rel(api, storage, "S3")
    Rel(worker, redis, "Redis")
    Rel(worker, db, "SQL")
    Rel(worker, storage, "S3")

    UpdateRelStyle(web, api, $offsetY="-50")
    UpdateRelStyle(api, redis, $offsetY="-50")
    UpdateRelStyle(worker, redis, $offsetY="-50")
    UpdateRelStyle(web, mapProvider, $offsetX="-40", $offsetY="-25")
    UpdateRelStyle(api, db, $offsetX="-30")
    UpdateRelStyle(api, storage, $offsetX="20")
    UpdateRelStyle(worker, db, $offsetX="-120", $offsetY="10")
    UpdateRelStyle(worker, storage, $offsetX="30", $offsetY="10")
    UpdateLayoutConfig($c4ShapeInRow="4", $c4BoundaryInRow="1")
```

## 3. Компоненты Backend API

```mermaid
flowchart LR
    web["<b>Web App</b><br/>React, TypeScript"]

    subgraph api["Backend API (Go)"]
        direction LR
        guard["<b>Защита запросов</b><br/>rate limiting, сессия и роль (RBAC), валидация"]
        auth["<b>Вход и сессии</b><br/>почта или пароль, сессии"]
        catalog["<b>Каталог мест</b><br/>поиск, дубли, срок действия"]
        moderation["<b>Заявки и модерация</b><br/>заявки, жалобы, запросы роли, история, решения"]
        files["<b>Файлы</b><br/>приём и проверка фото"]
        audit["<b>Журнал действий</b><br/>запись решений"]
        integr["<b>Внешние клиенты</b><br/>почта, геокодирование"]
    end

    subgraph data["Данные"]
        direction TB
        db[("<b>База данных</b><br/>PostgreSQL + PostGIS")]
        redis[("<b>Redis</b><br/>кэш, лимиты, сессии, очередь")]
        storage[("<b>Хранилище файлов</b><br/>MinIO (S3)")]
    end

    subgraph ext["Внешние сервисы"]
        direction TB
        mail["<b>Почтовый сервис</b>"]
        maps["<b>Картографический сервис</b>"]
    end

    web -->|запросы| guard
    guard --> auth
    guard --> catalog
    guard --> moderation
    guard --> files
    auth --> integr
    catalog --> integr
    moderation -->|публикует место, дубли| catalog
    moderation --> audit
    integr --> mail
    integr --> maps
    auth --> db
    auth -->|сессии| redis
    catalog --> db
    catalog -->|кэш| redis
    moderation --> db
    audit --> db
    files --> storage
    files -->|очередь| redis

    classDef comp fill:#dbe9fb,stroke:#2f6fed,color:#1d2330
    classDef cont fill:#2f6fed,stroke:#1d4fb0,color:#ffffff
    classDef ext fill:#e3e5e8,stroke:#8a919e,color:#1d2330
    class guard,auth,catalog,moderation,files,audit,integr comp
    class web,db,redis,storage cont
    class mail,maps ext
```

## 4. Сквозной сценарий: заявка, модерация, поиск

```mermaid
sequenceDiagram
    autonumber
    actor S as Студент
    actor Mod as Модератор
    participant W as Web App
    participant A as Backend API
    participant R as Redis
    participant K as Worker
    participant F as Хранилище файлов
    participant D as База данных

    Note over S,D: 1. Подача заявки
    S->>W: Заполняет форму нового места
    W->>A: Проверить дубли
    A->>D: Ищет похожие места (название и расстояние)
    A-->>W: Похожие места, если есть
    S->>W: Подтверждает и отправляет заявку с фото
    W->>A: Создать заявку
    A->>A: Проверяет поля, права и размер файла
    A->>F: Сохраняет фото как непроверенное
    A->>D: Сохраняет заявку (на модерации) и историю статусов
    A->>R: Ставит задачу проверить фото
    A-->>W: Заявка принята

    Note over S,D: 2. Проверка фото в фоне
    K->>R: Берёт задачу
    K->>F: Читает файл и проверяет его тип по содержимому
    alt Файл безопасен
        K->>F: Сохраняет обработанное фото
        K->>D: Фото принято
    else Файл невалиден
        K->>D: Фото отклонено, причина в заявке
    end

    Note over Mod,D: 3. Модерация
    Mod->>W: Открывает очередь заявок
    W->>A: Получить заявки
    A->>D: Читает заявки на модерации
    A-->>W: Список заявок
    alt Одобрить
        Mod->>W: Одобряет заявку
        W->>A: Одобрить
        A->>A: Проверяет, что это модератор
        A->>D: Публикует место, пишет историю и журнал действий
        A->>R: Сбрасывает кэш поиска
    else Отклонить
        Mod->>W: Указывает причину и отклоняет
        W->>A: Отклонить
        A->>D: Сохраняет причину, пишет историю и журнал действий
    end

    Note over S,D: 4. Поиск на карте
    S->>W: Открывает карту, выбирает категорию
    W->>A: Найти места
    A->>R: Есть ответ в кэше?
    alt Кэш пуст
        A->>D: Ищет опубликованные места в области карты
        A->>R: Кладёт результат в кэш
    end
    A-->>W: Места для карты
    W-->>S: Новое место видно на карте
```
