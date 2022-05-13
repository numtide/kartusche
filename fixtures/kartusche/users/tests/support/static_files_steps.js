step('^I get the index path$', () => {
    world.response = apiCall("GET", "/")

})

step('^the Kartusche should respond with an index html file$', () => {
    expect.equal(world.response.statusCode, 200)
    expect.equal(world.response.header.get("content-type"), "text/html; charset=utf-8")
})

step('^I get the text file path$', () => {
    world.response = apiCall("GET", "/readme.txt")
})

step('^the Kartusche should respond with the text file$', () => {
    expect.equal(world.response.statusCode, 200)
    expect.equal(world.response.header.get("content-type"), "text/plain; charset=utf-8")
})

step('^the etag header should be included$', () => {
    expect.equal(world.response.header.get("etag"), '"703c445982e074e33a05c161d221217f2facbf5e"')
})

step('^I get the text file path with if-none-match header set to content etag$', () => {
    world.response = apiCall("GET", "/readme.txt", null, { "if-none-match": '"703c445982e074e33a05c161d221217f2facbf5e"' })
})

step('^the Kartusche should respond with NotModified status code$', () => {
    expect.equal(world.response.statusCode, 304)
})