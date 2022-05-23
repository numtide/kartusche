Feature: list chats

    Scenario: getting the list of the initial chat
        When I list chats
        Then the list should have only the support chat