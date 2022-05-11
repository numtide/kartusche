Feature: deploy kartusche

    Scenario: deploying Kartusche
        Given the server is running
        When I deploy Kartusche "users"
        Then the Kartusche "users" should be running

    