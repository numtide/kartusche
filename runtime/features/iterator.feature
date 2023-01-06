Feature: iterator

    Scenario: iterating pver an empty map
        Given an existing map
        When I iterate over the map
        Then the result should be empty array


    Scenario: iterating over map with one element
        Given an existing map
        And the map has one element
        When I iterate over the map
        Then the result should contain the element

    Scenario: iterating over map with two elements
        Given an existing map
        And the map has two elements
        When I iterate over the map
        Then the result should contain both elements

    Scenario: iterating over map with two elements with seek
        Given an existing map
        And the map has two elements
        When I iterate over the map seeking to the second element
        Then the result should contain only the second element