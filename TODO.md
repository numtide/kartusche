* ~~rename `api` to `handlers`~~
* ~~add support for static files~~
    * ~~support etags~~
* ~~custom godog formatter to spew out JS code for missing steps~~
* ~~separate data from api/static by prefixing all dbpaths with 'data'~~
* support for websockets
    * `websocket.js` handler
        * TBD the API for callbacks etc
    * add `watch` function for database
* add access to runtime DB from the cucumber tests
* add support for `init.js` file
* consider support for larger binary files (reading in tx instead of caching in mem)
* come up with a concept of software updates (PATCH)
* create server and cli to:
    * list kartusches
    * create a new kartusche
    * delete kartusche
    * force kartusche backup
    * update kartusche
    * pause kartusche
    * resume kartusche
    * download kartusche
* come up with a concept of cronjobs
    * regular crons - maybe static
    * programmatic
* come up with a concept for performing HTTP requests from Kartusche
    * async ?
    * sync ?
* support for logging
* support for Kartusche audit log
* support on the js level for uploading content directly into the db
* support on the js level for fetching content directly from the db
* support request context when executing handlers
* support for the server to capture kartusche failures
    * current content of Kartusche
    * offending http requests

