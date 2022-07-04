Feature: authentication

    Scenario: authentication using browser
        Given the server is running
        When I authenticate the user using browser
        Then the user config should contain token for the server
