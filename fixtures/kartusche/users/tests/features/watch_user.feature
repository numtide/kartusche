Feature: create user

    Scenario: getting the initial state
        Given an existing user
        When I watch the user
        Then I should get initial state of the user

    Scenario: watching for updates
        Given I am watching an existing user
        When I update the user's email
        Then I should get an update
