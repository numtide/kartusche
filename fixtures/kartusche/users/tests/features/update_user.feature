Feature: updating user

    Scenario: updating email
        Given an existing user
        When I update the user's email
        Then the user should have the new email
