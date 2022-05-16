Feature: listg users

    Scenario: list when there are no users
        Given there are no users created
        When I list users
        Then the user list should be empty
    
    Scenario: list one user
        When I create a new user that does not exist
        When I list users
        Then the user list should contain the user

