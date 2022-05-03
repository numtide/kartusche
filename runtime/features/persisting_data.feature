Feature: persisting data

    Scenario: storing and retrieving single string
        Given a service with the "GET" handler "/data"
            """
            w.Write(read((tx) => tx.Get(["data"])))
            """
        Given a service with the "POST" handler "/data"
            """
            write((tx) => {
                tx.Put(["data"],requestBody())
            })
            w.WriteHeader(204)
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
