Feature: handler response

    Scenario: responding with a 200 status code
        Given a service with the "GET" handler "/foo"
            """
            w.write("hello world!")
            """
        When I get the path "/foo" from the Kartusche
        Then I should get status code 200

