Feature: static files

    Scenario: index file
        When I get the index path
        Then the Kartusche should respond with an index html file


    Scenario: text file
        When I get the text file path
        Then the Kartusche should respond with the text file
