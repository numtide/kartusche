step('^I list chats$', () => {
	world.res = apiCall("GET","/api/chats")
    expect.equal(world.res.statusCode, 200)
})

step('^the list should have only the support chat$', () => {
	expect.deepEqual(JSON.parse(readToString(world.res.body)), ["support"])
})
