Feature: iterator

    Scenario: iterating an empty map
        Given an existing map
        When I iterate over the map
        Then the result should be empty
