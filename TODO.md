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
* ~~create server and cli to:~~
    * ~~update kartusche's code~~
    * ~~upload kartusche~~
    * ~~list kartusches~~
    * ~~delete kartusche~~
* ~~manifest for Kartusche~~
    * ~~redefine where static files are~~
    * ~~fix update code to use the same logic~~
    * define how to build static files (wonder if worth the trouble!)
* ~~development server cli~~
    * ~~watches for changed files and does automatic code update~~
    * ~~run tests after the code or test update~~
* ~~add CLI to initialize a new Kartusche~~
* add SPA option into the manifest?
* ~~add support for HTTP requests from Kartusche~~
* get kartusche backup
* initialize kartusche config
* pause kartusche
* resume kartusche
* get kartusche info
    * list of handlers
    * list of static files
    * request counters per handler and static file
    * db stats?
* support for .kartuscheignore
* add executing of `update.js` after updating code
* add closing of http requests from tests
* add code to close db watches from handlers
* wrap handlers into functions - support for easy return
* add access to runtime DB from the cucumber tests
* consider support for larger binary files (reading in tx instead of caching in mem)
* come up with a concept of cronjobs
    * regular crons - maybe static
    * programmatic
* support for logging
* support for Kartusche audit log
* support on the js level for uploading content directly into the db
* support on the js level for fetching content directly from the db
* add support for prometheus, exporting stats of each Kartusche
* support for the server to capture kartusche failures
    * current content of Kartusche
    * offending http requests
* add special handler for websockets - GET is misleading

