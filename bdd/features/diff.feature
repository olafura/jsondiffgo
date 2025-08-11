Feature: JSON diff
  As a developer
  I want to compute jsondiffpatch-style diffs
  So that I can track changes between JSON structures

  Scenario: Simple value change
    Given JSON A:
      """
      {"test": 1}
      """
    And JSON B:
      """
      {"test": 2}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"test": [1, 2]}
      """

  Scenario: Array deletion
    Given JSON A:
      """
      {"test": [1,2,3]}
      """
    And JSON B:
      """
      {"test": [2,3]}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"test": {"_0": [1, 0, 0], "_t": "a"}}
      """

  Scenario: Object inside array
    Given JSON A:
      """
      {"test": [{"x":1}]}
      """
    And JSON B:
      """
      {"test": [{"x":2}]}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"test": {"0": {"x": [1, 2]}, "_t": "a"}}
      """

