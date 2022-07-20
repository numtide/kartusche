step('^I update the user\'s email$', () => {
    const res = apiCall("PATCH", `/users/${world.userId}`, JSON.stringify({ email: "f@b.com" }), { "content-type": "application/json" })
    expect.equal(res.statusCode, 204)
})

step('^the user should have the new email$', () => {
    const res = apiCall("GET",`/users/${world.userId}`)
    expect.equal(res.statusCode, 200)
    const us = JSON.parse(readToString(res.body))
    expect.equal( us.email, "f@b.com")

})