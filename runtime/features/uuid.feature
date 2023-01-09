Feature: generate and parse UUIDs


    Scenario: Generating v6 uuid
        When when I generate a v6 UUID
        Then generated UUID should not be an empty string


    Scenario: Parsing v6 uuid
        When when I generate and parse a v6 UUID
        Then the result should be a Date
