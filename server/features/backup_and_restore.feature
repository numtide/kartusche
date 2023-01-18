Feature: backup and restore

    Scenario: backup and restore of a single kartusche
        Given a server with a single kartusche
        When I backup the server
        When I delete the kartusche
        And I restore the server
        Then the kartusche should existd