Feature: JSON patch
  As a developer
  I want to apply jsondiffpatch-style diffs
  So that I can transform one JSON into another

  Scenario: Simple patch value
    Given Original JSON:
      """
      {"test": 1}
      """
    And Diff:
      """
      {"test": [1, 2]}
      """
    When I apply the patch
    Then the result equals:
      """
      {"test": 2}
      """

  Scenario: Patch array replace tail
    Given Original JSON:
      """
      {"test": [1,2,3]}
      """
    And Diff:
      """
      {"test": {"2": [4], "_2": [3, 0, 0], "_t": "a"}}
      """
    When I apply the patch
    Then the result equals:
      """
      {"test": [1,2,4]}
      """

  Scenario: Patch object inside array
    Given Original JSON:
      """
      {"test": [{"x":1}]}
      """
    And Diff:
      """
      {"test": {"0": {"x": [1, 2]}, "_t": "a"}}
      """
    When I apply the patch
    Then the result equals:
      """
      {"test": [{"x":2}]}
      """

