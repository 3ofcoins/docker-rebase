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

  Scenario Outline: Rebase from a (gzipped) tarball to a (gzipped) tarball
    When I run: <precmd> docker-rebase <flags> $FIXTURE_SMOKE_SHORT_ID $FIXTURE_SMOKE_BASE_SHORT_ID
    Then file "<outfile>" should contain an image
    And the image's JSON should be like:
      | $.id          | $FIXTURE_SMOKE_ID      |
      | $.parent      | $FIXTURE_SMOKE_BASE_ID |
      | $.author      | docker-rebase FTW!     |
      | $.config.User | nobody                 |
    
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

    And the image should add "/etc/init.d/S20urandom"
    And the image should not delete "/etc/init.d/S20urandom"

    Examples:
    | precmd                                         | flags                                       | outfile |
    | gunzip -c fixtures/smoke.tar.gz > smoke.tar && | -load smoke.tar -save out.tar               | out.tar |
    | gunzip -c fixtures/smoke.tar.gz > smoke.tar && | -load smoke.tar -zsave out.tgz              | out.tgz |
    |                                                | -zload fixtures/smoke.tar.gz -save out.tar  | out.tar |
    |                                                | -zload fixtures/smoke.tar.gz -zsave out.tgz | out.tgz |
