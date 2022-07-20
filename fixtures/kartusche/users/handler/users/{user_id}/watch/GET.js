const sel = upgradeToWebsocket((msg) => println(msg))
select(
    watch(["users", vars.user_id], () => {
        const user = read(tx => tx.get(["users", vars.user_id]))
        wsSendJson(JSON.parse(user))
    })
)
