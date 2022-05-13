step(`^I create a new user that does not exist$`, () => {
    const res = apiCall("POST","/users", JSON.stringify({email: "foo@bar.com", username: "xxx", password: "foobar"}), {"content-type": "application/json"})
    expect.equal(res.statusCode, 200)
    const {user_id:userId} =  JSON.parse(readToString(res.body))
    world.userId = userId
})

step(`^I the user should exist$`, () => {
    const res = apiCall("GET",`/users/${world.userId}`)
    expect.equal(res.statusCode, 200)
})
