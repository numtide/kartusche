step('^there are no users created$', () => {
    
})

step('^I list users$', () => {
    const res = apiCall("GET",`/users`)
    expect.equal(res.statusCode, 200)
    world.listOfUsers = JSON.parse(readToString(res.body))
})

step('^the user list should be empty$', () => {
    expect.deepEqual([],world.listOfUsers)
})


step('^the user list should contain the user$', () => {
    expect.equal(world.listOfUsers.length,1)
})