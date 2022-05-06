Feature: iterator

    Scenario: iterating over existing data
        Given a service with the "GET" handler "/list"
            """
            const allKeys = read(tx => {
                const result = []
                for (let it=tx.iterator(["data"]); !it.isDone(); it.next()) {
                    result.push(it.getKey())
                }
                return result
            })
            w.write(JSON.stringify(allKeys))
            """
        And a service with the "POST" handler "/append"
            """
            write((tx) => {
                if (!tx.exists(["data"])) {
                    tx.createMap(["data"])
                }
                tx.put(["data",tx.size(["data"])],requestBody())
            })
            w.writeHeader(204)
            """
        And I post following data to the "/append" path of the Kartusche
            """
            some data
            """
        And I should get status code 204
        When I get the path "/list" from the Kartusche
        Then I should get status code 200
        And the result body should match JSON
            """
            ["0"]
            """
