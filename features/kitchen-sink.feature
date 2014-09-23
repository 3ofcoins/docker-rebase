Feature: Kitchen Sink of Features
  In order to get some testing done
  As the author of kitchen-rebase
  I want all my scenarios in a single file to sort them later

  Scenario: The binary start and shows usage
    When I run `docker-rebase -help`
    Then the output should contain "Usage:"

  Scenario: Fixtures are symlinked
    When I successfully run `ls fixtures/`
    Then the output should contain "smoke.tar.gz"

  Scenario: Help text is not too wide
    When I successfully run `docker-rebase -help`
    Then no line of output exceeds 78 characters
            
  Scenario Outline: Disallowed flag combinations
    When I run `docker-rebase <flags>`
    Then the exit status should not be 0
    And the output should contain "Usage:"
    Examples:
    | flags            |
    | -build -load -   |
    | -build -zload -  |
    | -load - -zload - |
    | -save - -zsave - |
    |                  |
    | FOO              |
  
  Scenario: All-in-one simple rebase
    When I successfully run `docker-rebase -zload fixtures/smoke.tar.gz -zsave smoke.tgz e690d26ab126 a9eb17255234`
    Then file "smoke.tgz" should contain an image
    And the image's JSON should be like:
      | $.id                   | e690d26ab126791ea9ddaac6d1b512dfc5f77d653688148dfd47af5cc5f4c4dc |
      | $.parent               | a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721 |
      | $.author               | docker-rebase FTW!                                               |
      | $.config.Entrypoint[*] | /hello                                                           |
      | $.config.User          | nobody                                                           |
    
    And the image should add "/hello"
    And the image should add "/network/"
    And the image should add "/network/interfaces"
      
    And the image should delete "/etc/init.d/rcK"
    And the image should delete "/home"
    
    And the image should not add "/df"
    And the image should not delete "/df"
    And the image should not add "/network/if-up.d/"
    And the image should not delete "/network/if-up.d/"

# Scenario: remove from base, then add in child
