const sel = upgradeToWebsocket((msg) => println(msg))
select(
    watch(["users"], () => {
        const users = read(tx => {
            const users = []
            for (const it = tx.iterator(["users"]); !it.isDone(); it.next()) {
                users.push(it.getKey())
            }
            return users
        })
        wsSendHtml(render_template_to_s("list-of-users.mustache", { users }))
    })
)
