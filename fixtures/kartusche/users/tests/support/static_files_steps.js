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