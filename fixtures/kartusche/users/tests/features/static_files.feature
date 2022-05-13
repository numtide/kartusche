Feature: static files

    Scenario: index file
        When I get the index path
        Then the Kartusche should respond with an index html file


    Scenario: text file
        When I get the text file path
        Then the Kartusche should respond with the text file
        And the etag header should be included


    Scenario: responding when file has not changed
        When I get the text file path with if-none-match header set to content etag
        Then the Kartusche should respond with NotModified status code