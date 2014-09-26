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
    When I successfully run `docker-rebase -zload fixtures/smoke.tar.gz -zsave smoke.tgz 183fcd3414f6 a9eb17255234`
    Then file "smoke.tgz" should contain an image
    And the image's JSON should be like:
      | $.id                   | 183fcd3414f6604ba10ecf0bf8af0170fa679946cd87530ab827b2f32e21e147 |
      | $.parent               | a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721 |
      | $.author               | docker-rebase FTW!                                               |
      | $.config.User          | nobody                                                           |
    
    And the image should add "/baz"
    And the image should add "/network/"
    And the image should add "/network/interfaces"
    And the image should add "/foobar"
    And the image should not add "/foo/"
    And the image should not delete "/foo/"
      
    And the image should delete "/etc/init.d/rcK"
    And the image should delete "/home"
    And the image should add "/home.is.not.here"
    
    And the image should not add "/df"
    And the image should not delete "/df"
    And the image should not add "/network/if-up.d/"
    And the image should not delete "/network/if-up.d/"

# Scenario: remove from base, then add in child
