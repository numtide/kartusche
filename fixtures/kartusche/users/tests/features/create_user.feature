Feature: create user

Scenario: creating a new user
    When I create a new user that does not exist
    Then I the user should exist
    And the user name should be "ha!"
