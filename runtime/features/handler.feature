Feature: http request handlers


    Scenario: GET handler for root
        Given a kartusche with a root get handler
        When the kartusche receives GET request
        Then the kartusche should respond with 200 status code