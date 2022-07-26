# Handlers

Handlers are JS scripts that are executed when a HTTP request reaches a Kartusche.

## Matching HTTP Paths
Each handler is located in the `handler` directory or a sub-directory thereof.
The directory name in which the handler code is located matches the path of the HTTP request.

For example, handler located in `handlers` directory will match the request made to the root (`/`) path of the HTTP request.
Handler contained in `handlers/api` would match any request with the path `/api`

## Named Paths
If any part of the path consists of a name enclosed in curly braces (e.g. `{name}`) that path will match any sub-path and the name of the matched path will be contained in the `vars` object passed to the handler.

## Matching Verbs
Handler for each HTTP verb (`GET`, `PUT`, `POST`, `DELETE`, ...) is located in the directory matching the path with the name of the verb and extension `.js`.
For example, handler for `GET` HTTP request to `/api/users` would be the file `handler/api/users/GET.js`.

