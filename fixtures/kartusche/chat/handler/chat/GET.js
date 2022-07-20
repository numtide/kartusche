const username = r.url.query().get("username")

write(tx => {
    tx.put(["chat", uuidv7()], `${username} has joined`)
})

select(
    upgradeToWebsocket(({ message }) => {
        write(tx => {
            tx.put(["chat", uuidv7()], `${username}: ${message}`)
        })
    }),
    watch(["chat"], () => {
        const lines = read(tx => {
            const lines = []
            for (const it = tx.iterator(["chat"]); !it.isDone(); it.next()) {
                const line = it.getValue()
                lines.push(line)
            }
            return lines
        })
        wsSendHtml(render_template_to_s("chat-latest.mustache", { lines }))
        return false
    })
)

write(tx => {
    while(tx.size(["chat"]) > 10) {
        const it= tx.iterator(["chat"])
        if (it.isDone()) {
            break
        }
        const k = it.getKey()
        tx.delete(["chat",k])
    }
    tx.put(["chat", uuidv7()], `${username} has left`)
})
