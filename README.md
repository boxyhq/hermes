# Hermes
Audit logs service [Audit logs in a box from BoxyHQ]

A grade 36 Bureaucrat just like Hermes Conrad. Audit logs matters that only a true bureaucrat can handle properly.

# Source code visualizer
[CodeSee codebase visualizer](https://app.codesee.io/maps/public/a248cc70-23b7-11ec-a724-79d7dd589517)

# Run
** This project is still in beta, please get in touch if you'd like to use it in production. There are more backends being supported **

- You'll need to first configure and run [Loki v1.5.0](https://github.com/grafana/loki)

- Then build and run hermes (you might need to modify the config). Docker image, Compose yaml and Kubernetes yaml will be coming soon.

- Hermes currently has 2 APIs available.
  - `POST /ingest`: This endpoint ingests an array of audit logs and stores them in Loki. It needs an Authorization header containing the Api-Key. The body is an array of [audit logs](https://github.com/boxyhq/hermes/blob/main/types/audit-log.go)
  - `POST /query`: This endpoint queries audit logs. The body contains the `query` (0 or more indexes as key-values), `start` and `end` (RFC3339).

## POST /ingest
  ```console
  curl --location --request POST 'http://localhost:8080/ingest' \
--header 'Content-Type: application/json' \
--header 'Authorization: Api-Key abcdef' \
--data-raw '[
    {
        "actor": "deepak",
        "actor_type": "user",
        "group": "boxyhq",
        "where": "127.0.0.1",
        "where_type": "ip",
        "when": "2021-05-18T20:53:39+01:00",
        "target": "user.login",
        "target_id": "target_id",
        "action": "login",
        "action_type": "U",
        "name": "user.login",
        "description": "This is a login event",
        "metadata": {
            "foo": "bar",
            "hey": "you"
        }
    }
]'
```

## POST /query
```
curl --location --request POST 'http://localhost:8080/query' \
--header 'Content-Type: application/json' \
--header 'Authorization: Api-Key abcdef' \
--data-raw '{
    "query": {
    },
    "start": "2021-05-18T20:51:39+01:00",
    "end": "2021-05-18T20:56:39+01:00"
}
'
```
