step('^an existing user$', () => {
    const res = apiCall("POST", "/users", JSON.stringify({ email: "foo@bar.com", username: "xxx", password: "foobar" }), { "content-type": "application/json" })
    expect.equal(res.statusCode, 200)
    const { user_id: userId } = JSON.parse(readToString(res.body))
    world.userId = userId
})

step('^I watch the user$', () => {
    world.wsClient = connectWebsocket(`/users/${world.userId}/watch`)
})

step('^I should get initial state of the user$', () => {
    const us = world.wsClient.readJson()
    expect.equal(us.email, "foo@bar.com")
})

step('^I am watching an existing user$', () => {
    const res = apiCall("POST", "/users", JSON.stringify({ email: "foo@bar.com", username: "xxx", password: "foobar" }), { "content-type": "application/json" })
    expect.equal(res.statusCode, 200)
    const { user_id: userId } = JSON.parse(readToString(res.body))
    world.userId = userId
    world.wsClient = connectWebsocket(`/users/${world.userId}/watch`)
    world.wsClient.readJson()
})

step('^I should get an update$', () => {
    const us = world.wsClient.readJson()
    expect.equal(us.email, "f@b.com")
})
