step('^I request my IP$', () => {
    const res = apiCall("GET",`/my_ip`)
    expect.equal(res.statusCode, 200)
    world.myIP = JSON.parse(readToString(res.body)).ip

})
step('^I should see my IP$', () => {
	expect.matches(world.myIP,/^\d{1,3}(\.\d{1,3}){3}$/)
})
