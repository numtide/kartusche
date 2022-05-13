* ~~rename `api` to `handlers`~~
* ~~add support for static files~~
    * ~~support etags~~
* separate data from api/static by prefixing all dbpaths with 'data'
* support for websockets
    * `websocket.js` handler
        * TBD the API for callbacks etc
    * add `watch` function for database
* add access to runtime DB from the cucumber tests
* custom godog formatter to spew out JS code for missing steps
* add support for `init.js` file
* consider support for larger binary files (reading in tx instead of caching in mem)