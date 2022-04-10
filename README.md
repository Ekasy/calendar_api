# Api приложения НеКалендарь

Данное приложение предоставляет простое api для обработки и хранения данных приложения НеКалендарь. В нашем приложении можно создавать встречи, редактировать их, добавлять участников и не только!

## Структура хранения данных

Для хранения данных мы используем MongoDB. Основная коллекция - коллекция событий `nocalendar` имеет вид:
* набор юзеров. Ничего необычного
```
{
    "_id": "json/users",
    "users": {
        "user1": {
            "name": "<имя пользователя>",
            "surname": "<фамилия>",
            "email": "<почта>",
            "password": "<пароль>",
            "token": "<токен>"
        },
        ...
    },

}
```
* набор событий
```
{
    "_id": "json/events",
    "events": {
        "event1": {
            "meta: {
                "title": "<заголовок события>",
                "description": "<описание>",
                "timestamp": <таймстемп события>,
                "members": [
                    "<список участников события>",
                ],
                "active_members": [
                    "<список участников события, планирующих его посетить>",
                ],
                "author": "<создатель события>"
                "is_regular": true|false,
                "delta": <регулярность повторения события в днях>
            },
            "actual": {
                "title": "<заголовок события>",
                "description": "<описание>",
                "timestamp": <таймстемп события>,
                "members": [
                    "<список участников события>",
                ],
                "active_members": [
                    "<список участников события, планирующих его посетить>",
                ],
                "author": "<создатель события>"
            }
        },
        ...
    }
}
```
Для каждого события храним 2 ключа:
* `meta` - полное описание события. Обновляется если хотим обновить именно **регулярное** событие.
* `actual` - актуальная копия события из `meta`. Это, например, разовое событие. Также при разовом редактировании регулярного события меняем именно это поле.

## Ручки
Во все запросы необходимо передавать, дополнительно, заголовок `Authorization` с токеном авторизации пользователя. Конкретно такой вид: `Authorization: <token>`. Токен может меняться сервером, поэтому необходимо копировать его из ответа сервера и вставлять в новый запрос.

* `POST /api/auth` - аутентификация пользователя
    Тело запроса:
    ```
    {
        "login": "<уникальный логин>",
        "password": "<пароль>"
    }
    ```

    Ответ сервера:
    - `200 {"message": "ok"}` **устанавливается токен в заголовке!!!**
    - `400 {"message": "incorrect password"}`
    - `404 {"message": "user not found"}`
---

* `POST /api/register` - зарегестрировать пользователя

    Тело запроса:
    ```
    {
        "login": "<уникальный логин>",
        "password": "<пароль>",
        "name": "<имя пользователя>",
        "surname": "<фамилия>",
        "email": <почта>
    }
    ```

    Ответ сервера:
    - `200 {"message": "ok"}` **устанавливается токен в заголовке!!!**
    - `400 {"message": "login|email" already is used"}`
---

* `GET /api/events` - вернуть все события пользователя

    Обязательные cgi параметры:
    - `from` - timestamp с какого времени искать событие
    - `to` - timestamp до какого времени искать событие

    Ответ сервера:
    - `200`
        ```
        {
            "message": "ok",
            "events" = [
                "event1": {
                    "id": "<уникальный id ивента>",
                    "title": "<заголовок события>",
                    "description": "<описание>",
                    "timestamp": <таймстемп события>,
                    "members": [
                        "<список участников события>",
                    ],
                    "author": "<создатель события>"
                    "active_members": [
                        "<список участников события, планирующих его посетить>",
                    ]
                },
                ...
            ]
        }
        ```
    - `400 {"message": "invalid timestamps"}`
---

* `GET /api/event/<уникальный id ивента>` - вернуть информацию о событии
    Ответ сервера:
    - `200`
        ```
        {
            "message": "ok",
            "event": {
                "id": "<уникальный id ивента>"
                "title": "<заголовок события>",
                "description": "<описание>",
                "timestamp": <таймстемп события>,
                "members": [
                    "<список участников события>",
                ],
                "active_members": [
                    "<список участников события, планирующих его посетить>",
                ],
                "author": "<создатель события>"
                "is_regular": true|false,
                "delta": <регулярность повторения события в днях>
            }
        }
        ```
    - `400 {"message": "incorrect event id"}`
---

* `POST /api/event` - создать событие

    Тело запроса:
    ```
    {
        "title": "<заголовок события>",
        "description": "<описание>",
        "timestamp": <таймстемп события>,
        "members": [
            "<список участников события>",
        ],
        "is_regular": true|false,  // optional
        "delta": <регулярность повторения события в днях>  // require with is_regular field
    }
    ```

    Ответ сервера:
    - `200`
        ```
        {
            "message": "ok",
            "event_id": "<уникальный id события>"
        }
        ```
    - `400 {"message": "incorrect field"}`
---

* `POST /api/event/<уникальный id ивента>`

    Тело запроса:
    ```
    {
        "title": "<заголовок события>",
        "description": "<описание>",
        "timestamp": <таймстемп события>,
        "members": [
            "<список участников события>",
        ],
        "is_regular": true|false,
        "delta": <регулярность повторения события в днях>  // require if is_regular is true
    }
    ```
    Ответ сервера:
    - `200 {"message": 'ok"}`
    - `400 {"message": "incorrect event id"}`
    - `403 {"message": "has not permissions to edit"}`
---

* `DELETE /api/event/<уникальный id ивента>`

    Ответ сервера:
    - `200 {"message": 'ok"}`
    - `400 {"message": "incorrect event id"}`
    - `403 {"message": "only author can delete event"}`
