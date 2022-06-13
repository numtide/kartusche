Feature: my ip API

    Scenario: getting my ip
        When I request my IP
        Then I should see my IP