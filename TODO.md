* ~~rename `api` to `handlers`~~
* ~~add support for static files~~
    * ~~support etags~~
* ~~custom godog formatter to spew out JS code for missing steps~~
* ~~separate data from api/static by prefixing all dbpaths with 'data'~~
* ~~support for context.Context~~
* ~~support for websockets~~
* ~~add `watch` function for database~~
* ~~add support for `init.js` file~~
* ~~fix bug when requiring same library from multiple places~~
* ~~accelerate tests by creating one 'master' kartusche file and just copy it for each test~~
* create server and cli to:
    * ~~upload kartusche~~
    * list kartusches
    * delete kartusche
    * get kartusche backup
    * update kartusche's code
    * pause kartusche
    * resume kartusche
* add closing of http requests from tests
* add code to close db watches from handlers
* add special handler for websockets - GET is misleading
* wrap handlers into functions - support for easy return
* add access to runtime DB from the cucumber tests
* consider support for larger binary files (reading in tx instead of caching in mem)
* come up with a concept of software updates (PATCH)
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
* support for the server to capture kartusche failures
    * current content of Kartusche
    * offending http requests

