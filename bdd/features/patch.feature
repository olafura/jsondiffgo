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
  Scenario: Patch array reorder
    Given Original JSON:
      """
      {"1": [1,2,3]}
      """
    And Diff:
      """
      {"1": {"_0": ["", 2, 3], "_2": ["", 0, 3], "_t": "a"}}
      """
    And JS compare is skipped
    When I apply the patch
    Then the result equals:
      """
      {"1": [3,2,1]}
      """

  Scenario: Patch array shift one inside
    Given Original JSON:
      """
      {"1": [1,2,3]}
      """
    And Diff:
      """
      {"1": {"2": [0], "_t": "a"}}
      """
    When I apply the patch
    Then the result equals:
      """
      {"1": [1,2,0,3]}
      """

  Scenario: Patch deleted key works
    Given Original JSON:
      """
      {"foo": 1}
      """
    And Diff:
      """
      {"bar": [3], "foo": [1, 0, 0]}
      """
    When I apply the patch
    Then the result equals:
      """
      {"bar": 3}
      """

  Scenario: Patch with two-digit index key
    Given Original JSON:
      """
      {"cards": [{"foo1": true}, {"foo2": true}, {"foo3": true},
                 {"foo4": true}, {"foo5": true}, {"foo6": true},
                 {"foo7": true}, {"foo8": true}, {"foo9": true},
                 {"foo10": true}, {"foo11": true}, {"foo12": true}]}
      """
    And Diff:
      """
      {"cards": {"12": {"foo11": [true, false]}, "_t": "a"}}
      """
    And JS compare is skipped
    When I apply the patch
    Then the result equals:
      """
      {"cards": [{"foo1": true}, {"foo2": true}, {"foo3": true},
                 {"foo4": true}, {"foo5": true}, {"foo6": true},
                 {"foo7": true}, {"foo8": true}, {"foo9": true},
                 {"foo10": true}, {"foo11": true}, {"foo12": true}]}
      """

  Scenario: Big patch from fixtures
    Given Original JSON from file "testdata/big_json1.json"
    And Diff:
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
    When I apply the patch
    Then the result equals:
      """
      {"_id": "56353d1bec3821c78ad14479",
       "index": 1,
       "guid": "b2b50dae-5d30-4514-82b1-26714d91e264",
       "isActive": false,
       "balance": "$3,273.15",
       "picture": "http://placehold.it/32x32",
       "age": 29,
       "eyeColor": "blue",
       "name": "Talley Reyes",
       "gender": "male",
       "company": "ANDRYX",
       "email": "talleyreyes@andryx.com",
       "phone": "+1 (895) 435-3714",
       "address": "919 Lefferts Avenue, Winchester, Colorado, 2905",
       "about": "Consequat ullamco proident anim sunt ipsum esse Lorem tempor pariatur. Nostrud officia mollit aliqua sit consectetur sint minim veniam proident labore anim incididunt ex. Est amet laboris pariatur ut id qui et.\r\n",
       "registered": "2015-03-11T11:45:43 +04:00",
       "latitude": 39.655822,
       "longitude": -70.899696,
       "friends": [
        {"id": 0, "name": "Shelby Barrett"},
        {"id": 1, "name": "Gloria Mccray"},
        {"id": 2, "name": "Hopper Luna"}
       ],
       "greeting": "Hello, Talley Reyes! You have 2 unread messages.",
       "favoriteFruit": "banana"}
      """
