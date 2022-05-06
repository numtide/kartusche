Feature: persisting data

    Scenario: storing and retrieving single string
        Given a service with the "GET" handler "/data"
            """
            w.write(read((tx) => tx.get(["data"])))
            """
        And a service with the "POST" handler "/data"
            """
            write((tx) => {
                tx.put(["data"],requestBody())
            })
            w.writeHeader(204)
            """
        When I post following data to the "/data" path of the Kartusche
            """
            {"foo": "bar"}
            """
        When I get the path "/data" from the Kartusche
        Then I should get status code 200
        And the result body should match JSON
            """
            {"foo": "bar"}
            """
