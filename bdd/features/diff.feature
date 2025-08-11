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
  Scenario: Array all changed
    Given JSON A:
      """
      {"1": [1,2,3]}
      """
    And JSON B:
      """
      {"1": [4,5,6]}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"1": {"0": [4], "1": [5], "2": [6], "_0": [1,0,0], "_1": [2,0,0], "_2": [3,0,0], "_t": "a"}}
      """

  Scenario: Same object yields empty diff
    Given JSON A:
      """
      {"1": [1,2,3], "2": 1}
      """
    And JSON B:
      """
      {"1": [1,2,3], "2": 1}
      """
    When I compute the diff
    Then the diff equals:
      """
      {}
      """

  Scenario: Array shift one at start
    Given JSON A:
      """
      {"1": [1,2,3]}
      """
    And JSON B:
      """
      {"1": [0,1,2,3]}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"1": {"0": [0], "_t": "a"}}
      """

  Scenario: Array with duplicate values
    Given JSON A:
      """
      {"1": [1,2,1,3,3,2]}
      """
    And JSON B:
      """
      {"1": [3,1,2,1,2,3,3,2,1]}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"1": {"_t": "a", "0": [3], "4": [2], "8": [1]}}
      """

  Scenario: Object with multiple values in array
    Given JSON A:
      """
      {"1": [{"1":1,"2":2}]}
      """
    And JSON B:
      """
      {"1": [{"1":2,"2":2}]}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"1": {"0": {"1": [1, 2]}, "_t": "a"}}
      """

  Scenario: Object with multiple values plus in array
    Given JSON A:
      """
      {"1": [{"1":1,"2":2},{"3":3,"4":4}]}
      """
    And JSON B:
      """
      {"1": [{"1":2,"2":2},{"3":5,"4":6}]}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"1": {"0": {"1": [1, 2]}, "1": {"3": [3, 5], "4": [4, 6]}, "_t": "a"}}
      """

  Scenario: One object in array
    Given JSON A:
      """
      {"1": [1]}
      """
    And JSON B:
      """
      {"1": [{"1":2}]}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"1": {"0": [{"1":2}], "_0": [1,0,0], "_t": "a"}}
      """

  Scenario: Deleted value with object change in array
    Given JSON A:
      """
      {"1": [1,{"1":1}]}
      """
    And JSON B:
      """
      {"1": [{"1":2}]}
      """
    When I compute the diff
    Then the diff equals:
      """
      {"1": {"0": [{"1":2}], "_0": [1,0,0], "_1": [{"1":1},0,0], "_t": "a"}}
      """

  Scenario: Big diff from fixtures
    Given JSON A from file "testdata/big_json1.json"
    And JSON B from file "testdata/big_json2.json"
    When I compute the diff
    Then the diff equals:
      """
      {"_id": ["56353d1bca16dd7354045f7f", "56353d1bec3821c78ad14479"],
       "about": [
        "Laborum cupidatat proident deserunt fugiat aliquip deserunt. Mollit deserunt amet ut tempor veniam qui. Nulla ipsum non nostrud ut magna excepteur nulla non cupidatat magna ipsum.\r\n",
        "Consequat ullamco proident anim sunt ipsum esse Lorem tempor pariatur. Nostrud officia mollit aliqua sit consectetur sint minim veniam proident labore anim incididunt ex. Est amet laboris pariatur ut id qui et.\r\n"
       ],
       "address": [
        "265 Sutton Street, Tioga, Hawaii, 9975",
        "919 Lefferts Avenue, Winchester, Colorado, 2905"
       ],
       "age": [21, 29],
       "balance": ["$1,343.75", "$3,273.15"],
       "company": ["RAMJOB", "ANDRYX"],
       "email": ["eleanorbaxter@ramjob.com", "talleyreyes@andryx.com"],
       "eyeColor": ["brown", "blue"],
       "favoriteFruit": ["apple", "banana"],
       "gender": ["female", "male"],
       "friends": {"0": {"name": ["Larsen Sawyer", "Shelby Barrett"]}, "1": {"name": ["Frost Carey", "Gloria Mccray"]}, "2": {"name": ["Irene Lee", "Hopper Luna"]}, "_t": "a"},
       "greeting": ["Hello, Eleanor Baxter! You have 8 unread messages.", "Hello, Talley Reyes! You have 2 unread messages."],
       "guid": ["809e01c1-b8c4-4d49-a9e7-204091cd6ae8", "b2b50dae-5d30-4514-82b1-26714d91e264"],
       "index": [0, 1],
       "isActive": [true, false],
       "latitude": [-44.600585, 39.655822],
       "longitude": [-9.257008, -70.899696],
       "name": ["Eleanor Baxter", "Talley Reyes"],
       "phone": ["+1 (876) 456-3989", "+1 (895) 435-3714"],
       "registered": ["2014-07-20T11:36:42 +04:00", "2015-03-11T11:45:43 +04:00"]}
      """

