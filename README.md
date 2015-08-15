# aproxy33

Это прокси-сервис к API FreeFeed-а, добавляющий к нему OAuth2-атворизацию. Здесь представлена только серверная часть и сам прокси, клиентская часть должна общаться с сервером по REST-протоколу.

Компиляция: `go get github.com/davidmz/aproxy33`

Запуск: `aproxy33 -c config.json`

Сервис состоит из двух функциональных частей: менеджера OAuth2-данных и прокси. Менеджер управляет списком пользователей, приложениями и авторизационными токенами, а также самим процессом OAuth2-авторизации. Прокси проксирует запросы к FreeFeed API, преобразуя OAuth2-авторизацию в авторизацию по FreeFeed-токену.

## API менеджера

Все API-методы принимают в качестве тела запроса JSON-объект и возвращают тоже JSON-объект. 

Возвращаемый объект имеет формат: `{"status": "…", "data": …}`. Значения поля "status":

* "ok" — запрос выполнен успешно, в поле "data" находится результат. HTTP-код ответа 200.
* "error" — запрос не может быть выполнен, причина в самом запросе. В поле "data" находится строка с описанием ошибки. HTTP-код ответа 4xx.
* "fail" — запрос не удалось выполнить, причина на стороне сервера. В поле "data" находится строка с описанием ошибки. HTTP-код ответа 5xx.

Большинство методов API требует авторизации. Это не OAuth2-авторизация, а локальная авторизация для данного сервиса. Авторизация осуществляется передачей HTTP-заголовка `Authorization: X-AProxy TOKEN`, где TOKEN — токен, полученный API-методом `POST /aprox/auth`.

**TODO** Добавить возможность авторизоваться по FreeFeed-токену.

**Авторизация на сервере**

Метод принимает логин и пароль от FreeFeed-а и возвращает токен для авторизации на сервисе и время его жизни.

	POST /aprox/auth
	Auth: Not required
	Request:
	{
		"username": "frf_username",
		"password": "frf_password"
	}
	Response:
	{
		"token": "…",
		"ttl": 12345
	}

**Обновление токена**

Клиент должен сам следить за обновлением токена. Следующий метод возвращает новый токен, если к нему обратиться с авторизацией действующим токеном.

	POST /aprox/auth-refresh
	Auth: Required
	Request: (empty)
	Response:
	{
		"token": "…",
		"ttl": 12345
	}

**Список приложений пользователя**

Метод возвращает список приложений пользователя (для которых пользователь является владельцем).

	GET /aprox/apps
	Auth: Required
	Request: (empty)
	Response:
	[
		{
			"key": "…",			 // идентификатор приложения
			"secret": "…",		 // секрет приложения
			"date": "…",		 // время создания
			"title": "…",		 // название
			"description": "…",  // описание
			"domains": ["…", …]  // список доменов, с которых разрешена авторизация
		},
		…
	]

**Создание приложения**

	POST /aprox/apps
	Auth: Required
	Request:
	{
		"title": "…",
		"description": "…",
		"domains": ["…", …]
	}
	Response:
	{
		"key": "…",
		"secret": "…"
	}

**Изменение приложения**

	PUT /aprox/apps/{key}
	Auth: Required
	Request:
	{
		"title": "…",
		"description": "…",
		"domains": ["…", …]
	}
	Response:
	null

**Удаление приложения**

	DELETE /aprox/apps/{key}
	Auth: Required
	Request: (empty)
	Response:
	null

**Публичная информация о приложении**

	GET /aprox/apps/{key}/info
	Auth: Not required
	Request: (empty)
	Response:
	{
		"app": 		{
			"key": "…",
			"date": "…",
			"title": "…",
			"description": "…",
			"domains": ["…", …]
		},
		"api": {
			"api_method": "title",
			…
		}
	}

Этот метод доступен без авторизации.

Метод возвращает информацию о приложении а также хэш из доступных методов API и их человекопонятных названий. "api_method" имеет вид "METHOD{SPACE}URI_MASK", например: "POST /v1/users", "GET /v1/users/*" или "POST /v1/groups/*/subscribers/*/admin". Звёздочкой заменяются участки URI, которые могут изменяться.

**Получение списка токенов**

Метод возвращает список OAuth2-токенов пользователя.

	GET /aprox/tokens
	Auth: Required
	Request: (empty)
	Response:
	[
		{
			"id": 123,			// ID токена (числовой)
			"date": "…",		// время создания
			"perms": {			// методы API, доступные токену и их названия
				"api_method": "title",
				…
			},
			"app": {				 // информация о приложении
				"title": "…",		 // название
				"description": "…",  // описание
				"owner": "…",  	     // frf-username владельца приложения
				"domains": ["…", …]  // список доменов, с которых разрешена авторизация
			}
		},
		…
	]

**Удаление токена**

	DELETE /aprox/tokens/{id}
	Auth: Required
	Request: (empty)
	Response:
	null

**Получение авторизационного OAuth2-кода**

Этот метод используется в процессе создания OAuth2-токена. Получив запрос, фронтенд должен спросить у юзера, хочет ли он авторизовать данное приложение, и если да, то обратиться за кодом авторизации:

	POST /aprox/oauth-code
	Auth: Required
	Request:
	{
		"app_key": "…",				// приложение, которое просит авторизации
		"redirect_uri": "…", 		// redirect_uri из запроса
		"perms": ["api_method", …]	// методы API, с которыми хочет работать приложение
	}
	Response:
	"auth_code"

При запросе кода приложение должно запросить список методов, с которыми оно будет работать. Это делается с помощью стандартного OAuth2-параметра "scope", который должен содержать разделённый пробелами список методов API. В этом списке пробел между HTTP-методом и URI заменяется нижним подчёркиванием. Пример: "POST_/v1/posts PUT_/v1/posts/* DELETE_/v1/posts/*" — приложение запрашивает права на создание, редактирование и удаление постов. Подчёркивание используется только для совместимости с OAuth2-протоколом, фронтенд должен разобрать "scope", заменить подчёркивания обратно на пробелы и передать список в "POST /aprox/oauth-code" в параметре "perms".

Получив код (или отказ) фронтенд должен сам сформировать правильный URL для возврата.

**Получение авторизационного OAuth2-токена по коду**

Этот метод реализует требования протокола OAuth2 (https://tools.ietf.org/html/rfc6749#section-4.1.3), поэтому у него не такой формат вызова, как у остальных:

	POST /oauth/token
	Auth: Not required
	Request: application/x-www-form-urlencoded
		grant_type
		code
		redirect_uri
		client_id
	Response: (https://tools.ietf.org/html/rfc6749#section-5.1)
	{
		"access_token": "…",		// авторизационный токен
		"token_type":   "bearer",
		"username": "…"				// frf-username пользователя, которому принадлежит токен
	}

Формат ответа в случае ошибки также отличается и соответствует https://tools.ietf.org/html/rfc6749#section-5.2

## API прокси

Для образения к API FreeFeed-а нужно использовать URI, начинающиеся с "/frf/". Например, для обращения к методу "POST /v1/posts", надо сделать запрос к прокси "POST /frf/v1/posts".

При запросе следует использовать ранее полученный авторизационный токен посредством HTTP-заголовка `Authorization: Bearer TOKEN`. Если данный API-метод разрешён для этого токена, прокси преобразует его в токен FreeFeed-а и выполнит запрос.

Отдельные методы API FreeFeed-а не требуют авторизации, в этом случае заголовок "Authorization" можно не указывать — метод будет вызван анонимно.


